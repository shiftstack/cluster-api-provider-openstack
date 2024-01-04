/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha8"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/cloud/services/networking"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/scope"
	ipamutils "sigs.k8s.io/cluster-api-provider-openstack/pkg/utils/ipam"
)

// IPAddressReconciler reconciles a IPAddress object.
type IPAddressReconciler struct {
	Client           client.Client
	Recorder         record.EventRecorder
	WatchFilterValue string
	ScopeFactory     scope.Factory
	CaCertificates   []byte // PEM encoded ca certificates.

	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ipam.cluster.x-k8x.io.cluster.x-k8s.io,resources=ipaddresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.cluster.x-k8x.io.cluster.x-k8s.io,resources=ipaddresses/status,verbs=get;update;patch

func (r *IPAddressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx).WithValues("ipaddress", req.NamespacedName)

	ipAddress := &ipamv1.IPAddress{}
	if err := r.Client.Get(ctx, req.NamespacedName, ipAddress); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if ipAddress.Spec.PoolRef.Kind != openStackFloatingIPPool {
		return ctrl.Result{}, nil
	}

	pool := &infrav1.OpenStackFloatingIPPool{}

	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: ipAddress.Namespace,
		Name:      ipAddress.Spec.PoolRef.Name,
	}, pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("openStackFloatingIPPool", pool.Name)

	scope, err := r.ScopeFactory.NewClientScopeFromFloatingIPPool(ctx, r.Client, pool, r.CaCertificates, log)
	if err != nil {
		return reconcile.Result{}, err
	}

	networkingService, err := networking.NewService(scope)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !ipAddress.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(ipAddress, infrav1.DeleteFloatingIPFinalizer) {
			// If the pool has ReclaimPolicy=Delete, and the IP is not pre-allocated, delete the IP.
			if pool.Spec.ReclaimPolicy == infrav1.ReclaimDelete && !contains(pool.Spec.PreAllocatedFloatingIPs, ipAddress.Spec.Address) {
				if err = networkingService.DeleteFloatingIP(pool, ipAddress.Spec.Address); err != nil {
					return ctrl.Result{}, fmt.Errorf("delete floating IP %q: %w", ipAddress.Spec.Address, err)
				}
			} else if pool.Spec.ReclaimPolicy == infrav1.ReclaimRetain {
				// Return the IP to available IPs.
				pool.Status.AvailableIPs = append(pool.Status.AvailableIPs, ipAddress.Spec.Address)
				if err := r.Client.Status().Update(ctx, pool); err != nil {
					return ctrl.Result{}, err
				}
			}
			controllerutil.RemoveFinalizer(ipAddress, infrav1.DeleteFloatingIPFinalizer)
			if err := r.Client.Update(ctx, ipAddress); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		scope.Logger().Info("IPAddress is being deleted but has other finalizers, waiting for them to be removed")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPAddressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1.IPAddress{}, builder.WithPredicates(
			predicate.And(
				ipamutils.AddressReferencesPoolKind(
					metav1.GroupKind{
						Group: infrav1.GroupVersion.Group,
						Kind:  openStackFloatingIPPool,
					}),
				ipamutils.HasFinalizerAndIsDeleting(infrav1.DeleteFloatingIPFinalizer),
			)),
		).
		Complete(r)
}

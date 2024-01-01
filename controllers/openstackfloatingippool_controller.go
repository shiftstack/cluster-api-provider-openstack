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
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha7"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/cloud/services/networking"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/scope"
)

const (
	openStackFloatingIPPool = "OpenStackFloatingIPPool"
)

// OpenStackFloatingIPPoolReconciler reconciles a OpenStackFloatingIPPool object.
type OpenStackFloatingIPPoolReconciler struct {
	Client           client.Client
	Recorder         record.EventRecorder
	WatchFilterValue string
	ScopeFactory     scope.Factory
	CaCertificates   []byte // PEM encoded ca certificates.

	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=openstackfloatingippools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=openstackfloatingippools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ipam.cluster.x-k8s.io,resources=ipaddressclaims;ipaddressclaims/status,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups=ipam.cluster.x-k8s.io,resources=ipaddresses;ipaddresses/status,verbs=get;list;watch;create;update;delete
func (r *OpenStackFloatingIPPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	pool := &infrav1.OpenStackFloatingIPPool{}
	if err := r.Client.Get(ctx, req.NamespacedName, pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	scope, err := r.ScopeFactory.NewClientScopeFromFloatingIPPool(ctx, r.Client, pool, r.CaCertificates, log)
	if err != nil {
		return reconcile.Result{}, err
	}

	if pool.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer if it does not exist
		if controllerutil.AddFinalizer(pool, infrav1.OpenStackFloatingIPPoolFinalizer) {
			return ctrl.Result{}, r.Client.Update(ctx, pool)
		}
	} else {
		// Handle deletion
		return ctrl.Result{}, r.reconcileDelete(ctx, scope, pool)
	}

	patchHelper, err := patch.NewHelper(pool, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patchHelper.Patch(ctx, pool); err != nil {
			if reterr == nil {
				reterr = fmt.Errorf("error patching OpenStackFloatingIPPool %s/%s: %w", pool.Namespace, pool.Name, err)
			}
		}
	}()

	if err := r.reconcileFloatingIPNetwork(scope, pool); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.setIPStatuses(ctx, pool); err != nil {
		return ctrl.Result{}, err
	}

	claims := &ipamv1.IPAddressClaimList{}
	if err := r.Client.List(context.Background(), claims, client.InNamespace(req.Namespace), client.MatchingFields{infrav1.OpenStackFloatingIPPoolNameIndex: pool.Name}); err != nil {
		return ctrl.Result{}, err
	}

	for _, claim := range claims.Items {
		claim := claim
		log := log.WithValues("claim", claim.Name)
		if !claim.ObjectMeta.DeletionTimestamp.IsZero() {
			continue
		}

		if claim.Status.AddressRef.Name == "" {
			ipAddress := &ipamv1.IPAddress{}
			if err := r.Client.Get(ctx, client.ObjectKey{Name: claim.Name, Namespace: claim.Namespace}, ipAddress); err == nil {
				// IPAddress already exists, another reconciler is working on it
				log.Info("IPAddress already exists, another reconciler is working on it")
				continue
			}

			ip, err := r.getIP(scope, pool)
			if err != nil {
				return ctrl.Result{}, err
			}

			ipAddress = &ipamv1.IPAddress{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      claim.Name,
					Namespace: claim.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: claim.APIVersion,
							Kind:       claim.Kind,
							Name:       claim.Name,
							UID:        claim.UID,
						},
					},
				},
				Spec: ipamv1.IPAddressSpec{
					ClaimRef: corev1.LocalObjectReference{
						Name: claim.Name,
					},
					PoolRef: corev1.TypedLocalObjectReference{
						APIGroup: pointer.String(infrav1.GroupVersion.Group),
						Kind:     pool.Kind,
						Name:     pool.Name,
					},
					Address: ip,
					Prefix:  32,
				},
			}

			// If the ReclaimPolicy is Delete, we add the finalizer to the IPAddress if they are not pre-allocated
			// this is to ensure that the IPAddress is deleted from openstack when the claim is deleted
			if pool.Spec.ReclaimPolicy == infrav1.ReclaimDelete && !contains(pool.Spec.PreAllocatedFloatingIPs, ip) {
				controllerutil.AddFinalizer(ipAddress, infrav1.DeleteFloatingIPFinalizer)
			}

			if err = r.Client.Create(ctx, ipAddress); err != nil {
				return ctrl.Result{}, err
			}

			claim.Status.AddressRef.Name = ipAddress.Name
			if err = r.Client.Status().Update(ctx, &claim); err != nil {
				log.Error(err, "Failed to update IPAddressClaim status", "claim", claim.Name, "ip", ip)
				return ctrl.Result{}, err
			}
			scope.Logger().Info("Claimed IP", "ip", ip)
		}
	}
	if err = r.setIPStatuses(ctx, pool); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *OpenStackFloatingIPPoolReconciler) reconcileDelete(ctx context.Context, scope scope.Scope, pool *infrav1.OpenStackFloatingIPPool) error {
	log := ctrl.LoggerFrom(ctx)
	ipAddresses := &ipamv1.IPAddressList{}
	if err := r.Client.List(ctx, ipAddresses, client.InNamespace(pool.Namespace), client.MatchingFields{infrav1.OpenStackFloatingIPPoolNameIndex: pool.Name}); err != nil {
		return err
	}
	// If there are still IPAddress objects that are not deleted, there are still claims on this pool and we should not delete it
	// beause the pool is needed to clean up the addresses from openstack
	if len(ipAddresses.Items) > 0 {
		log.Info("Waiting for IPAddress to be deleted before deleting OpenStackFloatingIPPool")
		return errors.New("waiting for IPAddress to be deleted, until we can delete the OpenStackFloatingIPPool")
	}

	networkingService, err := networking.NewService(scope)
	if err != nil {
		return err
	}

	// Clean up ips created by the pool
	for _, ip := range diff(pool.Status.IPs, pool.Spec.PreAllocatedFloatingIPs) {
		if err := networkingService.DeleteFloatingIP(pool, ip); err != nil {
			return fmt.Errorf("delete floating IP: %w", err)
		}
	}

	if controllerutil.RemoveFinalizer(pool, infrav1.OpenStackFloatingIPPoolFinalizer) {
		log.Info("Removing finalizer from OpenStackFloatingIPPool")
		return r.Client.Update(ctx, pool)
	}
	return nil
}

func union(a []string, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	for _, item := range b {
		m[item] = struct{}{}
	}
	result := make([]string, 0, len(m))
	for item := range m {
		result = append(result, item)
	}
	return result
}

func diff(a []string, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	for _, item := range b {
		delete(m, item)
	}
	result := make([]string, 0, len(m))
	for item := range m {
		result = append(result, item)
	}
	return result
}

func (r *OpenStackFloatingIPPoolReconciler) setIPStatuses(ctx context.Context, pool *infrav1.OpenStackFloatingIPPool) error {
	ipAddresses := &ipamv1.IPAddressList{}
	if err := r.Client.List(ctx, ipAddresses, client.InNamespace(pool.Namespace), client.MatchingFields{infrav1.OpenStackFloatingIPPoolNameIndex: pool.Name}); err != nil {
		return err
	}
	pool.Status.ClaimedIPs = []string{}
	for _, ip := range ipAddresses.Items {
		pool.Status.ClaimedIPs = append(pool.Status.ClaimedIPs, ip.Spec.Address)
	}
	pool.Status.IPs = append([]string{}, pool.Spec.PreAllocatedFloatingIPs...)

	pool.Status.IPs = union(pool.Status.IPs, pool.Status.ClaimedIPs)
	pool.Status.AvailableIPs = diff(diff(pool.Status.IPs, pool.Status.ClaimedIPs), pool.Status.FailedIPs)
	return nil
}

func (r *OpenStackFloatingIPPoolReconciler) getIP(scope scope.Scope, pool *infrav1.OpenStackFloatingIPPool) (string, error) {
	var ip string

	networkingService, err := networking.NewService(scope)
	if err != nil {
		scope.Logger().Error(err, "Failed to create networking service")
		return "", err
	}

	if len(pool.Status.AvailableIPs) > 0 {
		ip = pool.Status.AvailableIPs[0]
		pool.Status.AvailableIPs = pool.Status.AvailableIPs[1:]
		pool.Status.ClaimedIPs = append(pool.Status.ClaimedIPs, ip)
	}

	if ip != "" {
		fp, err := networkingService.GetFloatingIP(ip)
		if err != nil {
			return "", fmt.Errorf("get floating IP: %w", err)
		}
		// If the IP exist return it, else we continue and try to allocate it if we fail to allocate it we will mark it as failed
		if fp != nil {
			return fp.FloatingIP, nil
		}
	}

	// ip could be empty meaning we want to get a new one or the IP
	// if ip is not empty it got and ip from availableIPs that does not exist in openstack
	// we try to allocate it, if we fail we mark it as failed and skip it next time
	fp, err := networkingService.CreateFloatingIPForPool(pool, ip)
	if err != nil {
		scope.Logger().Error(err, "Failed to create floating IP", "pool", pool.Name, "ip", ip)
		// If we tried to allocate a specific IP, we should mark it as failed so we don't try again
		// this should only happen if the pool thinks this IP is available and we do not have permission to allocate a specific IP
		if ip != "" {
			pool.Status.FailedIPs = append(pool.Status.FailedIPs, ip)
		}
		return "", err
	}

	ip = fp.FloatingIP
	pool.Status.ClaimedIPs = append(pool.Status.ClaimedIPs, ip)
	pool.Status.IPs = append(pool.Status.IPs, ip)
	return ip, nil
}

func (r *OpenStackFloatingIPPoolReconciler) reconcileFloatingIPNetwork(scope scope.Scope, pool *infrav1.OpenStackFloatingIPPool) error {
	// If the pool already has a network, we don't need to do anything
	if pool.Status.FloatingIPNetwork != nil {
		return nil
	}

	networkingService, err := networking.NewService(scope)
	if err != nil {
		return err
	}

	netListOpts := external.ListOptsExt{
		ListOptsBuilder: pool.Spec.FloatingIPNetwork.ToListOpt(),
		External:        pointer.Bool(true),
	}

	networkList, err := networkingService.GetNetworksByFilter(&netListOpts)
	if err != nil {
		return fmt.Errorf("failed to find network: %w", err)
	}
	if len(networkList) > 1 {
		return fmt.Errorf("found multiple networks, expects filter to match one (result: %v)", networkList)
	}

	if pool.Status.FloatingIPNetwork == nil {
		pool.Status.FloatingIPNetwork = &infrav1.NetworkStatus{}
	}
	pool.Status.FloatingIPNetwork.ID = networkList[0].ID
	pool.Status.FloatingIPNetwork.Name = networkList[0].Name
	pool.Status.FloatingIPNetwork.Tags = networkList[0].Tags
	return nil
}

func (r *OpenStackFloatingIPPoolReconciler) iPAddressClaimToPoolMapper(_ context.Context, o client.Object) []ctrl.Request {
	claim, ok := o.(*ipamv1.IPAddressClaim)
	if !ok {
		panic(fmt.Sprintf("Expected a IPAddressClaim but got a %T", o))
	}
	if claim.Spec.PoolRef.Kind != openStackFloatingIPPool {
		return nil
	}
	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      claim.Spec.PoolRef.Name,
				Namespace: claim.Namespace,
			},
		},
	}
}

func (r *OpenStackFloatingIPPoolReconciler) ipAddressToPoolMapper(_ context.Context, o client.Object) []ctrl.Request {
	ip, ok := o.(*ipamv1.IPAddress)
	if !ok {
		panic(fmt.Sprintf("Expected a IPAddress but got a %T", o))
	}
	if ip.Spec.PoolRef.Kind != openStackFloatingIPPool {
		return nil
	}
	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      ip.Spec.PoolRef.Name,
				Namespace: ip.Namespace,
			},
		},
	}
}

func (r *OpenStackFloatingIPPoolReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1.IPAddressClaim{}, infrav1.OpenStackFloatingIPPoolNameIndex, func(rawObj client.Object) []string {
		claim := rawObj.(*ipamv1.IPAddressClaim)
		if claim.Spec.PoolRef.Kind != openStackFloatingIPPool {
			return nil
		}
		return []string{claim.Spec.PoolRef.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1.IPAddress{}, infrav1.OpenStackFloatingIPPoolNameIndex, func(rawObj client.Object) []string {
		ip := rawObj.(*ipamv1.IPAddress)
		return []string{ip.Spec.PoolRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.OpenStackFloatingIPPool{}).
		Watches(
			&ipamv1.IPAddressClaim{},
			handler.EnqueueRequestsFromMapFunc(r.iPAddressClaimToPoolMapper),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Watches(
			&ipamv1.IPAddress{},
			handler.EnqueueRequestsFromMapFunc(r.ipAddressToPoolMapper),
		).
		Complete(r)
}

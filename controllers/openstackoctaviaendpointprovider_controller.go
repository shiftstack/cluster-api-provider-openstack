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

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha7"
)

const (
	OctaviaEndpointProviderFieldManager = "octavia-endpoint-provider"
)

// OpenStackOctaviaEndpointProviderReconciler reconciles a OpenStackOctaviaEndpointProvider object
type OpenStackOctaviaEndpointProviderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=openstackoctaviaendpointproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=openstackoctaviaendpointproviders/status,verbs=get;update;patch

func (r *OpenStackOctaviaEndpointProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("openstackoctaviaendpointprovider", req.NamespacedName)

	// Fetch the OctaviaEndpointProvider instance.
	endpointProvider := &infrav1.OpenStackOctaviaEndpointProvider{}
	err := r.Client.Get(ctx, req.NamespacedName, endpointProvider)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if annotations.HasPaused(endpointProvider) {
		log.Info("OpenStackOctaviaEndpointProvider is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	if !endpointProvider.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, endpointProvider)
	}

	return r.reconcileNormal(ctx, endpointProvider)
}

func (r *OpenStackOctaviaEndpointProviderReconciler) reconcileNormal(ctx context.Context, endpointProvider *infrav1.OpenStackOctaviaEndpointProvider) (_ ctrl.Result, reterr error) {
	// If the OpenStackMachine doesn't have our finalizer, add it.
	if !controllerutil.ContainsFinalizer(endpointProvider, infrav1.OctaviaEndpointProviderFinalizer) {
		unstructured := unstructured.Unstructured{}
		initPatch(&unstructured, endpointProvider)

		unstructured.SetFinalizers([]string{infrav1.OctaviaEndpointProviderFinalizer})

		// Patch and return immediately
		return ctrl.Result{}, r.Client.Patch(ctx, &unstructured, client.Apply, &client.PatchOptions{FieldManager: OctaviaEndpointProviderFieldManager})
	}

	var err error

	reconciliation := endpointReconciliation{
		r: r,
	}
	initPatch(&reconciliation.patch, endpointProvider)
	reconciliation.networking, err = networking.NewService(scope)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = errors.Join(err, r.resolveVIPs(ctx, endpointProvider, patch, &dirty))

	return ctrl.Result{}, nil
}

type endpointReconciliation struct {
	r *OpenStackOctaviaEndpointProviderReconciler

	dirty bool
	patch infrav1.OpenStackOctaviaEndpointProvider

	network *networking.Service
}

func (r *OpenStackOctaviaEndpointProviderReconciler) resolveVIPs(ctx context.Context, endpointProvider, patch *infrav1.OpenStackOctaviaEndpointProvider, dirty *bool) error {
	// If the VIPs are already set, we don't need to do anything
	if len(endpointProvider.Status.VIPs) > 0 {
		patch.Status.VIPs = endpointProvider.Status.VIPs
		return nil
	}

	networkID, err := 
	vips := make([]infrav1.OctaviaVIP, len(endpointProvider.Spec.VIPs))
	for i, vip := range endpointProvider.Spec.VIPs {

	}
	return nil
}

func (r *OpenStackOctaviaEndpointProviderReconciler) reconcileDelete(ctx context.Context, endpointProvider *infrav1.OpenStackOctaviaEndpointProvider) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenStackOctaviaEndpointProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.OpenStackOctaviaEndpointProvider{}).
		Complete(r)
}

type typedObject interface {
	schema.ObjectKind
	metav1.Object
}

func initPatch(patch typedObject, endpointProvider *infrav1.OpenStackOctaviaEndpointProvider) {
	patch.SetGroupVersionKind(endpointProvider.GroupVersionKind())
	patch.SetName(endpointProvider.Name)
	patch.SetNamespace(endpointProvider.Namespace)
}

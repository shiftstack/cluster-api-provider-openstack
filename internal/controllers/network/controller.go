/*
Copyright 2024 The Kubernetes Authors.

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

package network

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	orcv1alpha1 "github.com/k-orc/openstack-resource-controller/api/v1alpha1"

	"sigs.k8s.io/cluster-api-provider-openstack/pkg/scope"
	ctrlutil "sigs.k8s.io/cluster-api-provider-openstack/pkg/utils/controllers"
)

const (
	Finalizer = "openstack.k-orc.cloud/network"

	FieldOwner = "openstack.k-orc.cloud/networkcontroller"
	// Field owner of the object finalizer
	SSAFinalizerTxn = "finalizer"
	// Field owner of transient status
	SSAStatusTxn = "status"
	// Field owner of persistent id field
	SSAIDTxn = "id"
)

// ssaFieldOwner returns the field owner for a specific named SSA transaction
func ssaFieldOwner(txn string) client.FieldOwner {
	return client.FieldOwner(FieldOwner + "/" + txn)
}

// orcNetworkReconciler reconciles an ORC Network.
type orcNetworkReconciler struct {
	client           client.Client
	recorder         record.EventRecorder
	watchFilterValue string
	scopeFactory     scope.Factory
	caCertificates   []byte // PEM encoded ca certificates.
}

func New(client client.Client, recorder record.EventRecorder, watchFilterValue string, scopeFactory scope.Factory, caCertificates []byte) ctrlutil.SetupWithManager {
	return &orcNetworkReconciler{
		client:           client,
		recorder:         recorder,
		watchFilterValue: watchFilterValue,
		scopeFactory:     scopeFactory,
		caCertificates:   caCertificates,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *orcNetworkReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := mgr.GetLogger()
	return ctrl.NewControllerManagedBy(mgr).
		For(&orcv1alpha1.Network{}).
		WithOptions(options).
		WithEventFilter(needsReconcilePredicate(log)).
		Complete(r)
}

func needsReconcilePredicate(log logr.Logger) predicate.Predicate {
	filter := func(obj client.Object, event string) bool {
		log := log.WithValues("predicate", "NeedsReconcile", "event", event)

		orcResource, ok := obj.(*orcv1alpha1.Network)
		if !ok {
			log.V(0).Info("Expected Network", "type", fmt.Sprintf("%T", obj))
			return false
		}

		// Always reconcile deleted objects. Note that we don't always
		// get a Delete event for a deleted object. If the object was
		// deleted while the controller was not running we will get a
		// Create event for it when the controller syncs.
		if !orcResource.DeletionTimestamp.IsZero() {
			return true
		}

		if !orcv1alpha1.IsReconciliationComplete(orcResource) {
			return true
		}

		log.V(4).Info("not reconciling network due to terminal state", "name", orcResource.GetName(), "namespace", orcResource.GetNamespace(), "generation", orcResource.GetGeneration())
		return false
	}

	// We always reconcile create. We get a create event for every object when
	// the controller restarts as the controller has no previously observed
	// state at that time. This means that upgrading the controller will always
	// re-reconcile objects. This has the advantage of being a way to address
	// invalid state from controller bugs, but the disadvantage of potentially
	// causing a 'thundering herd' when the controller restarts.
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filter(e.ObjectNew, "Update")
		},
	}
}

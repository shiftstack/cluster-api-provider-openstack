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

package v1alpha7

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// OpenStackFloatingIPPoolFinalizer allows ReconcileOpenStackFloatingIPPool to clean up resources associated with OpenStackFloatingIPPool before
	// removing it from the apiserver.
	OpenStackFloatingIPPoolFinalizer = "openstackfloatingippool.infrastructure.cluster.x-k8s.io"

	OpenStackFloatingIPPoolNameIndex = "spec.poolRef.name"

	// OpenStackFloatingIPPoolIP.
	DeleteFloatingIPFinalizer = "openstackfloatingippool.infrastructure.cluster.x-k8s.io/delete-floating-ip"
)

// ReclaimPolicy is a string type alias to represent reclaim policies for floating ips.
type ReclaimPolicy string

const (
	// ReclaimDelete is the reclaim policy for floating ips.
	ReclaimDelete ReclaimPolicy = "Delete"
	// ReclaimRetain is the reclaim policy for floating ips.
	ReclaimRetain ReclaimPolicy = "Retain"
)

// OpenStackFloatingIPPoolSpec defines the desired state of OpenStackFloatingIPPool.
type OpenStackFloatingIPPoolSpec struct {
	PreAllocatedFloatingIPs []string `json:"preAllocatedFloatingIPs,omitempty"`

	// IdentityRef is a reference to a identity to be used when reconciling this pool.
	// +optional
	IdentityRef *OpenStackIdentityReference `json:"identityRef,omitempty"`

	// FloatingIPNetwork is the external network to use for floating ips, if there's only one external network it will be used by default
	FloatingIPNetwork NetworkFilter `json:"floatingIPNetwork"`

	// The name of the cloud to use from the clouds secret
	// +optional
	CloudName string `json:"cloudName"`

	// The stratergy to use for reclaiming floating ips when they are released from a machine
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Retain;Delete
	ReclaimPolicy ReclaimPolicy `json:"reclaimPolicy"`
}

// OpenStackFloatingIPPoolStatus defines the observed state of OpenStackFloatingIPPool.
type OpenStackFloatingIPPoolStatus struct {
	// +kubebuilder:default={}
	// +optional
	ClaimedIPs []string `json:"claimedIPs"`

	// +kubebuilder:default={}
	// +optional
	AvailableIPs []string `json:"availableIPs"`

	// +kubebuilder:default={}
	// +optional
	AllIPs []string `json:"allIPs"`

	// FailedIPs contains a list of floating ips that failed to be allocated
	// +optional
	FailedIPs []string `json:"failedIPs,omitempty"`

	// floatingIPNetwork contains information about the network used for floating ips
	// +optional
	FloatingIPNetwork *NetworkStatus `json:"floatingIPNetwork,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenStackFloatingIPPool is the Schema for the openstackfloatingippools API.
type OpenStackFloatingIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenStackFloatingIPPoolSpec   `json:"spec,omitempty"`
	Status OpenStackFloatingIPPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenStackFloatingIPPoolList contains a list of OpenStackFloatingIPPool.
type OpenStackFloatingIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenStackFloatingIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenStackFloatingIPPool{}, &OpenStackFloatingIPPoolList{})
}

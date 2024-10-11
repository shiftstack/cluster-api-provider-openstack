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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkResourceSpec struct {
	Name *string `json:"name,omitempty"`

	Description *string `json:"description,omitempty"`
}

// NetworkFilter identifies a resource to import
// +kubebuilder:validation:MinProperties:=1
type NetworkFilter struct {
	// Name specifies the name of the resource to import
	// +optional
	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:MaxLength:=1000
	Name *string `json:"name,omitempty"`
}

// NetworkImport specifies an existing resource which will be imported instead
// of creating a new one
// +kubebuilder:validation:MinProperties:=1
// +kubebuilder:validation:MaxProperties:=1
type NetworkImport struct {
	// ID contains the unique identifier of an existing resource. Note that
	// when specifying an import by ID, the resource MUST already exist.
	// This object will enter an error state if the resource does not
	// exist.
	// +optional
	// +kubebuilder:validation:Format:=uuid
	ID *string `json:"id,omitempty"`

	// Filter contains a query which is expected to return a single result.
	// The controller will continue to retry if filter returns no results.
	// If filter returns multiple results the controller will set an error
	// state and will not continue to retry.
	// +optional
	Filter *NetworkFilter `json:"filter,omitempty"`
}

// NetworkSpec defines the desired state of the resource.
// +kubebuilder:validation:XValidation:rule="self.managementPolicy == 'managed' ? has(self.resource) : true",message="resource must be specified when policy is managed"
// +kubebuilder:validation:XValidation:rule="self.managementPolicy == 'managed' ? !has(self.__import__) : true",message="import may not be specified when policy is managed"
// +kubebuilder:validation:XValidation:rule="self.managementPolicy == 'unmanaged' ? !has(self.resource) : true",message="resource may not be specified when policy is unmanaged"
// +kubebuilder:validation:XValidation:rule="self.managementPolicy == 'unmanaged' ? has(self.__import__) : true",message="import must be specified when policy is unmanaged"
// +kubebuilder:validation:XValidation:rule="has(self.managedOptions) ? self.managementPolicy == 'managed' : true",message="managedOptions may only be provided when policy is managed"
// +kubebuilder:validation:XValidation:rule="!has(self.__import__) ? has(self.resource.content) : true",message="resource content must be specified when not importing"
type NetworkSpec struct {
	// Import refers to an existing resource which will be imported instead
	// of creating a new one.
	// +optional
	Import *NetworkImport `json:"import,omitempty"`

	// Resource specifies the desired state of the OpenStack resource.
	//
	// Resource may not be specified if the management policy is `unmanaged`.
	//
	// Resource must be specified when the management policy is `managed`.
	// +optional
	Resource *NetworkResourceSpec `json:"resource,omitempty"`

	// ManagementPolicy defines how ORC will treat the object. Valid values are
	// `managed`: ORC will create, update, and delete the resource; `unmanaged`:
	// ORC will import an existing image, and will not apply updates to it or
	// delete it.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="managementPolicy is immutable"
	// +kubebuilder:default:=managed
	// +optional
	ManagementPolicy ManagementPolicy `json:"managementPolicy,omitempty"`

	// ManagedOptions specifies options which may be applied to managed objects.
	// +optional
	ManagedOptions *ManagedOptions `json:"managedOptions,omitempty"`

	// CloudCredentialsRef points to a secret containing OpenStack credentials
	// +kubebuilder:validation:Required
	CloudCredentialsRef CloudCredentialsReference `json:"cloudCredentialsRef"`
}

// NetworkResourceStatus represents the observed state of the OpenStack resource
type NetworkResourceStatus struct {
	// Name is the resource name as reported by the OpenStack API
	// +optional
	Name *string `json:"name,omitempty"`
}

// NetworkStatus defines the observed state of an Image.
type NetworkStatus struct {
	// Conditions represents the observed status of the object.
	// Known .status.conditions.type are: "Available", "Progressing"
	//
	// Available represents the availability of the glance image. If it is
	// true then the image is ready for use in glance, and its hash has been
	// verified.
	//
	// Progressing indicates the state of the glance image does not currently
	// reflect the desired state, but that reconciliation is progressing.
	// Progressing will be False either because the desired state has been
	// achieved, or some terminal error prevents it from being achieved and the
	// controller is no longer attempting to reconcile.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ID is the unique identifier of the resource in OpenStack
	// +optional
	ID *string `json:"id,omitempty"`

	// Resource contains the observed state of the resource
	// +optional
	Resource *NetworkResourceStatus `json:"resource,omitempty"`
}

var _ ObjectWithConditions = &Network{}

func (i *Network) GetConditions() []metav1.Condition {
	return i.Status.Conditions
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id",description="Glance image ID"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type=='Available')].status",description="Availability status of image"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Available')].message",description="Message describing current availability status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation"

// Network is the Schema for the ORC networks API.
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Image.
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}

func (i *Network) GetCloudCredentialsRef() (*string, *CloudCredentialsReference) {
	if i == nil {
		return nil, nil
	}

	return &i.Namespace, &i.Spec.CloudCredentialsRef
}

var _ CloudCredentialsRefProvider = &Network{}

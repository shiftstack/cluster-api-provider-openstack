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

// +kubebuilder:validation:Enum:=managed;unmanaged;deleteOnDetach
type ManagementPolicy string

const (
	// ManagementPolicyManaged specifies that the controller will reconcile the
	// state of the referenced OpenStack resource with the state of the ORC
	// object.
	ManagementPolicyManaged ManagementPolicy = "managed"

	// ManagementPolicyUnmanaged specifies that the controller will not make any
	// changes to the referenced OpenStack resource.
	ManagementPolicyUnmanaged ManagementPolicy = "unmanaged"

	// ManagementPolicyDetachOnDelete has the same behaviour as
	// ManagementPolicyManaged, except that the controller will not delete the
	// OpenStack resource when the ORC object is deleted.
	ManagementPolicyDetachOnDelete ManagementPolicy = "detachOnDelete"
)

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

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	OctaviaEndpointProviderFinalizer = "openstackoctaviaendpointprovider.infrastructure.cluster.x-k8s.io"
)

// OpenStackOctaviaEndpointProviderSpec defines the desired state of OpenStackOctaviaEndpointProvider
type OpenStackOctaviaEndpointProviderSpec struct {
	// VIPNetwork is the network which all VIPs must be attached to.
	VIPNetwork NetworkFilter `json:"vipNetwork"`

	// VIPs is the list of loadbalancer VIPs.
	// Until we add support for dual-stack loadbalancers this must contain exactly one subnet.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	// +listType=atomic
	VIPs []OctaviaVIP `json:"vips"`

	// Listeners is the list of listeners to create on the loadbalancer
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=port
	Listeners []ListenerSpec `json:"listeners"`

	// Ports is the list of ports to expose on the loadbalancer. These ports
	// must be the same on the member machines.
	// +kubebuilder:validation:MinItems=1
	// +listType=set
	Ports []int32 `json:"port,omitempty"`

	// OctaviaProvider is the name of the Octavia provider to use. If not
	// specified, Octavia will use its default provider.
	OctaviaProvider string `json:"octaviaProvider,omitempty"`

	// AllowInvalidOctaviaProvider will cause the controller to ignore an
	// invalid Octavia provider instead of failing.
	AllowInvalidOctaviaProvider bool `json:"allowInvalidOctaviaProvider,omitempty"`

	// Tags is a list of tags to attach to all OpenStack resources related
	// to this loadbalancer
	Tags []string `json:"tags,omitempty"`
}

type OctaviaVIP struct {
	// Subnet is the subnet to use for the VIP. It must be in the network
	// specified by VIPNetwork. If not specified and IP is specified,
	// Octavia will attempt to find a subnet for which IP is valid. If
	// neither Subnet nor IP are specified, Octavia will choose a subnet
	// automatically from VIPNetwork, preferring IPv4 over IPv6 subnets.
	Subnet *SubnetFilter `json:"subnet,omitempty"`

	// IP is a specific IP address to use for the loadbalancer's VIP. If
	// not specified it will be allocated automatically.
	IP string `json:"vip,omitempty"`

	// FloatingIP is a floating IP to attach to the VIP. It must already
	// exist. If specified, it will be published as the endpoint of this
	// VIP. If not specified, no floating IP will be attached to the VIP and
	// the VIP will be published as the endpoint.
	FloatingIP string `json:"floatingIP,omitempty"`
}

type OctaviaVIPStatus struct {
	// Subnet is the uuid of the subnet used for the loadbalancer's VIP.
	Subnet string `json:"subnet"`
}

type ListenerSpec struct {
	// Port is the port to listen on. Member machines must expose their service on the same port.
	Port int32 `json:"port"`

	// AllowedCIDRs
	AllowedCIDRs []string `json:"allowedCIDRs,omitempty"`

	HealthMonitor HealthMonitorSpec `json:"healthMonitor,omitempty"`
}

type HealthMonitorSpec struct {
	// Delay is the time in seconds between sending probes to members.
	Delay int

	// MaxRetries is the number of successful checks before changing the
	// operating status of the member to ONLINE.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	MaxRetries int

	// MaxRetriesDown is the allowed check failures before changing the
	// operating status of the member to ERROR.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	MaxRetriesDown int

	// Timeout is the maximum time, in seconds, that a monitor waits to
	// connect before it times out. This must be less than Delay.
	Timeout int

	// Check defines the behaviour of the check the monitor will perform
	Check HealthMonitorUnion `json:"check"`
}

type HealthMonitorUnion struct {
	// Type is the type of health monitor to use. The only supported type is
	// TCP.
	// +kubebuilder:validation:Enum=TCP
	Type string `json:"type"`
}

// OpenStackOctaviaEndpointProviderStatus defines the observed state of OpenStackOctaviaEndpointProvider
type OpenStackOctaviaEndpointProviderStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	Endpoints []clusterv1.APIEndpoint `json:"endpoints,omitempty"`

	FloatingIP string `json:"floatingIP,omitempty"`

	VIPNetwork string       `json:"vipNetwork,omitempty"`
	VIPs       []OctaviaVIP `json:"vips,omitempty"`

	LoadBalancerID string `json:"loadBalancerID,omitempty"`

	Listeners []ListenerStatus `json:"listeners,omitempty"`
}

type ListenerStatus struct {
	Port              int32             `json:"port"`
	PoolID            string            `json:"poolID"`
	AllowedCIDRs      []string          `json:"allowedCIDRs,omitempty"`
	HealthMonitorID   string            `json:"healthMonitorID,omitempty"`
	HealthMonitorSpec HealthMonitorSpec `json:"healthMonitorSpec,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenStackOctaviaEndpointProvider is the Schema for the openstackoctaviaendpointproviders API
type OpenStackOctaviaEndpointProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenStackOctaviaEndpointProviderSpec   `json:"spec,omitempty"`
	Status OpenStackOctaviaEndpointProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenStackOctaviaEndpointProviderList contains a list of OpenStackOctaviaEndpointProvider
type OpenStackOctaviaEndpointProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenStackOctaviaEndpointProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenStackOctaviaEndpointProvider{}, &OpenStackOctaviaEndpointProviderList{})
}

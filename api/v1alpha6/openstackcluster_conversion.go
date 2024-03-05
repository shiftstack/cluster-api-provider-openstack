/*
Copyright 2023 The Kubernetes Authors.

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

package v1alpha6

import (
	"reflect"

	apiconversion "k8s.io/apimachinery/pkg/conversion"
	ctrlconversion "sigs.k8s.io/controller-runtime/pkg/conversion"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/utils/conversion"
)

var _ ctrlconversion.Convertible = &OpenStackCluster{}

func (r *OpenStackCluster) ConvertTo(dstRaw ctrlconversion.Hub) error {
	dst := dstRaw.(*infrav1.OpenStackCluster)

	return conversion.ConvertAndRestore(
		r, dst,
		Convert_v1alpha6_OpenStackCluster_To_v1beta1_OpenStackCluster, Convert_v1beta1_OpenStackCluster_To_v1alpha6_OpenStackCluster,
		v1alpha6OpenStackClusterRestorer, v1beta1OpenStackClusterRestorer,
	)
}

func (r *OpenStackCluster) ConvertFrom(srcRaw ctrlconversion.Hub) error {
	src := srcRaw.(*infrav1.OpenStackCluster)

	return conversion.ConvertAndRestore(
		src, r,
		Convert_v1beta1_OpenStackCluster_To_v1alpha6_OpenStackCluster, Convert_v1alpha6_OpenStackCluster_To_v1beta1_OpenStackCluster,
		v1beta1OpenStackClusterRestorer, v1alpha6OpenStackClusterRestorer,
	)
}

var _ ctrlconversion.Convertible = &OpenStackClusterList{}

func (r *OpenStackClusterList) ConvertTo(dstRaw ctrlconversion.Hub) error {
	dst := dstRaw.(*infrav1.OpenStackClusterList)

	return Convert_v1alpha6_OpenStackClusterList_To_v1beta1_OpenStackClusterList(r, dst, nil)
}

func (r *OpenStackClusterList) ConvertFrom(srcRaw ctrlconversion.Hub) error {
	src := srcRaw.(*infrav1.OpenStackClusterList)

	return Convert_v1beta1_OpenStackClusterList_To_v1alpha6_OpenStackClusterList(src, r, nil)
}

/* Restorers */

var v1alpha6OpenStackClusterRestorer = conversion.RestorerFor[*OpenStackCluster]{
	"spec": conversion.HashedFieldRestorer(
		func(c *OpenStackCluster) *OpenStackClusterSpec {
			return &c.Spec
		},
		restorev1alpha6ClusterSpec,
	),
	"status": conversion.HashedFieldRestorer(
		func(c *OpenStackCluster) *OpenStackClusterStatus {
			return &c.Status
		},
		restorev1alpha6ClusterStatus,
	),
}

var v1beta1OpenStackClusterRestorer = conversion.RestorerFor[*infrav1.OpenStackCluster]{
	"externalNetwork": conversion.UnconditionalFieldRestorer(
		func(c *infrav1.OpenStackCluster) *infrav1.NetworkFilter {
			return &c.Spec.ExternalNetwork
		},
	),
	"disableExternalNetwork": conversion.UnconditionalFieldRestorer(
		func(c *infrav1.OpenStackCluster) *bool {
			return &c.Spec.DisableExternalNetwork
		},
	),
	"router": conversion.UnconditionalFieldRestorer(
		func(c *infrav1.OpenStackCluster) **infrav1.RouterFilter {
			return &c.Spec.Router
		},
	),
	"networkMtu": conversion.UnconditionalFieldRestorer(
		func(c *infrav1.OpenStackCluster) *int {
			return &c.Spec.NetworkMTU
		},
	),
	"bastion": conversion.HashedFieldRestorer(
		func(c *infrav1.OpenStackCluster) **infrav1.Bastion {
			return &c.Spec.Bastion
		},
		restorev1beta1Bastion,
	),
	"subnets": conversion.HashedFieldRestorer(
		func(c *infrav1.OpenStackCluster) *[]infrav1.SubnetFilter {
			return &c.Spec.Subnets
		},
		restorev1beta1Subnets,
	),
	"allNodesSecurityGroupRules": conversion.HashedFieldRestorer(
		func(c *infrav1.OpenStackCluster) *infrav1.ManagedSecurityGroups {
			return c.Spec.ManagedSecurityGroups
		},
		restorev1beta1ManagedSecurityGroups,
	),
	"status": conversion.HashedFieldRestorer(
		func(c *infrav1.OpenStackCluster) *infrav1.OpenStackClusterStatus {
			return &c.Status
		},
		restorev1beta1ClusterStatus,
	),
	"managedSubnets": conversion.UnconditionalFieldRestorer(
		func(c *infrav1.OpenStackCluster) *[]infrav1.SubnetSpec {
			return &c.Spec.ManagedSubnets
		},
	),
}

/* OpenStackClusterSpec */

func restorev1alpha6ClusterSpec(previous *OpenStackClusterSpec, dst *OpenStackClusterSpec) {
	for i := range previous.ExternalRouterIPs {
		dstIP := &dst.ExternalRouterIPs[i]
		previousIP := &previous.ExternalRouterIPs[i]

		// Subnet.Filter.ID was overwritten in up-conversion by Subnet.UUID
		dstIP.Subnet.Filter.ID = previousIP.Subnet.Filter.ID

		// If Subnet.UUID was previously unset, we overwrote it with the value of Subnet.Filter.ID
		// Don't unset it again if it doesn't have the previous value of Subnet.Filter.ID, because that means it was genuinely changed
		if previousIP.Subnet.UUID == "" && dstIP.Subnet.UUID == previousIP.Subnet.Filter.ID {
			dstIP.Subnet.UUID = ""
		}
	}

	// We only restore DNSNameservers when these were lossly converted when NodeCIDR is empty.
	if len(previous.DNSNameservers) > 0 && dst.NodeCIDR == "" {
		dst.DNSNameservers = previous.DNSNameservers
	}

	prevBastion := previous.Bastion
	dstBastion := dst.Bastion
	if prevBastion != nil && dstBastion != nil {
		restorev1alpha6MachineSpec(&prevBastion.Instance, &dstBastion.Instance)
	}

	// To avoid lossy conversion, we need to restore AllowAllInClusterTraffic
	// even if ManagedSecurityGroups is set to false
	if previous.AllowAllInClusterTraffic && !previous.ManagedSecurityGroups {
		dst.AllowAllInClusterTraffic = true
	}

	// Conversion to v1beta1 removes the Kind field
	dst.IdentityRef = previous.IdentityRef

	if len(dst.ExternalRouterIPs) == len(previous.ExternalRouterIPs) {
		for i := range dst.ExternalRouterIPs {
			restorev1alpha6SubnetFilter(&previous.ExternalRouterIPs[i].Subnet.Filter, &dst.ExternalRouterIPs[i].Subnet.Filter)
		}
	}

	restorev1alpha6SubnetFilter(&previous.Subnet, &dst.Subnet)

	restorev1alpha6NetworkFilter(&previous.Network, &dst.Network)
}

func Convert_v1alpha6_OpenStackClusterSpec_To_v1beta1_OpenStackClusterSpec(in *OpenStackClusterSpec, out *infrav1.OpenStackClusterSpec, s apiconversion.Scope) error {
	err := autoConvert_v1alpha6_OpenStackClusterSpec_To_v1beta1_OpenStackClusterSpec(in, out, s)
	if err != nil {
		return err
	}

	if in.ExternalNetworkID != "" {
		out.ExternalNetwork = infrav1.NetworkFilter{
			ID: in.ExternalNetworkID,
		}
	}

	emptySubnet := SubnetFilter{}
	if in.Subnet != emptySubnet {
		subnet := infrav1.SubnetFilter{}
		if err := Convert_v1alpha6_SubnetFilter_To_v1beta1_SubnetFilter(&in.Subnet, &subnet, s); err != nil {
			return err
		}
		out.Subnets = []infrav1.SubnetFilter{subnet}
	}

	// DNSNameservers without NodeCIDR doesn't make sense, so we drop that.
	if len(in.NodeCIDR) > 0 {
		out.ManagedSubnets = []infrav1.SubnetSpec{
			{
				CIDR:           in.NodeCIDR,
				DNSNameservers: in.DNSNameservers,
			},
		}
	}

	if in.ManagedSecurityGroups {
		out.ManagedSecurityGroups = &infrav1.ManagedSecurityGroups{}
		if !in.AllowAllInClusterTraffic {
			out.ManagedSecurityGroups.AllNodesSecurityGroupRules = infrav1.LegacyCalicoSecurityGroupRules()
		} else {
			out.ManagedSecurityGroups.AllowAllInClusterTraffic = true
		}
	}

	out.IdentityRef.CloudName = in.CloudName
	if in.IdentityRef != nil {
		out.IdentityRef.Name = in.IdentityRef.Name
	}

	return nil
}

func Convert_v1beta1_OpenStackClusterSpec_To_v1alpha6_OpenStackClusterSpec(in *infrav1.OpenStackClusterSpec, out *OpenStackClusterSpec, s apiconversion.Scope) error {
	err := autoConvert_v1beta1_OpenStackClusterSpec_To_v1alpha6_OpenStackClusterSpec(in, out, s)
	if err != nil {
		return err
	}

	if in.ExternalNetwork.ID != "" {
		out.ExternalNetworkID = in.ExternalNetwork.ID
	}

	if len(in.Subnets) >= 1 {
		if err := Convert_v1beta1_SubnetFilter_To_v1alpha6_SubnetFilter(&in.Subnets[0], &out.Subnet, s); err != nil {
			return err
		}
	}

	if len(in.ManagedSubnets) > 0 {
		out.NodeCIDR = in.ManagedSubnets[0].CIDR
		out.DNSNameservers = in.ManagedSubnets[0].DNSNameservers
	}

	if in.ManagedSecurityGroups != nil {
		out.ManagedSecurityGroups = true
		out.AllowAllInClusterTraffic = in.ManagedSecurityGroups.AllowAllInClusterTraffic
	}

	out.CloudName = in.IdentityRef.CloudName
	out.IdentityRef = &OpenStackIdentityReference{Name: in.IdentityRef.Name}

	return nil
}

/* OpenStackClusterStatus */

func restorev1alpha6ClusterStatus(previous *OpenStackClusterStatus, dst *OpenStackClusterStatus) {
	// PortOpts.SecurityGroups have been removed in v1beta1
	// We restore the whole PortOpts/Networks since they are anyway immutable.
	if previous.ExternalNetwork != nil {
		dst.ExternalNetwork.PortOpts = previous.ExternalNetwork.PortOpts
	}
	if previous.Network != nil {
		dst.Network = previous.Network
	}
	if previous.Bastion != nil && previous.Bastion.Networks != nil {
		dst.Bastion.Networks = previous.Bastion.Networks
	}

	restorev1alpha6SecurityGroup(previous.ControlPlaneSecurityGroup, dst.ControlPlaneSecurityGroup)
	restorev1alpha6SecurityGroup(previous.WorkerSecurityGroup, dst.WorkerSecurityGroup)
	restorev1alpha6SecurityGroup(previous.BastionSecurityGroup, dst.BastionSecurityGroup)
}

func restorev1beta1ClusterStatus(previous *infrav1.OpenStackClusterStatus, dst *infrav1.OpenStackClusterStatus) {
	// It's (theoretically) possible in v1beta1 to have Network nil but
	// Router or APIServerLoadBalancer not nil. In hub-spoke-hub conversion this will
	// result in Network being a pointer to an empty object.
	if previous.Network == nil && dst.Network != nil && reflect.ValueOf(*dst.Network).IsZero() {
		dst.Network = nil
	}

	dst.ControlPlaneSecurityGroup = previous.ControlPlaneSecurityGroup
	dst.WorkerSecurityGroup = previous.WorkerSecurityGroup
	dst.BastionSecurityGroup = previous.BastionSecurityGroup

	if previous.Bastion != nil {
		dst.Bastion.ReferencedResources = previous.Bastion.ReferencedResources
	}
	if previous.Bastion != nil && previous.Bastion.DependentResources.PortsStatus != nil {
		dst.Bastion.DependentResources.PortsStatus = previous.Bastion.DependentResources.PortsStatus
	}
}

func Convert_v1beta1_OpenStackClusterStatus_To_v1alpha6_OpenStackClusterStatus(in *infrav1.OpenStackClusterStatus, out *OpenStackClusterStatus, s apiconversion.Scope) error {
	err := autoConvert_v1beta1_OpenStackClusterStatus_To_v1alpha6_OpenStackClusterStatus(in, out, s)
	if err != nil {
		return err
	}

	// Router and APIServerLoadBalancer have been moved out of Network in v1beta1
	if in.Router != nil || in.APIServerLoadBalancer != nil {
		if out.Network == nil {
			out.Network = &Network{}
		}

		out.Network.Router = (*Router)(in.Router)
		out.Network.APIServerLoadBalancer = (*LoadBalancer)(in.APIServerLoadBalancer)
	}

	return nil
}

func Convert_v1alpha6_OpenStackClusterStatus_To_v1beta1_OpenStackClusterStatus(in *OpenStackClusterStatus, out *infrav1.OpenStackClusterStatus, s apiconversion.Scope) error {
	err := autoConvert_v1alpha6_OpenStackClusterStatus_To_v1beta1_OpenStackClusterStatus(in, out, s)
	if err != nil {
		return err
	}

	// Router and APIServerLoadBalancer have been moved out of Network in v1beta1
	if in.Network != nil {
		out.Router = (*infrav1.Router)(in.Network.Router)
		out.APIServerLoadBalancer = (*infrav1.LoadBalancer)(in.Network.APIServerLoadBalancer)
	}

	return nil
}

/* Bastion */

func restorev1beta1Bastion(previous **infrav1.Bastion, dst **infrav1.Bastion) {
	if *previous != nil && *dst != nil {
		restorev1beta1MachineSpec(&(*previous).Instance, &(*dst).Instance)
	}
}

func Convert_v1alpha6_Instance_To_v1beta1_BastionStatus(in *Instance, out *infrav1.BastionStatus, _ apiconversion.Scope) error {
	// BastionStatus is the same as Instance with unused fields removed
	out.ID = in.ID
	out.Name = in.Name
	out.SSHKeyName = in.SSHKeyName
	out.State = infrav1.InstanceState(in.State)
	out.IP = in.IP
	out.FloatingIP = in.FloatingIP
	return nil
}

func Convert_v1beta1_BastionStatus_To_v1alpha6_Instance(in *infrav1.BastionStatus, out *Instance, _ apiconversion.Scope) error {
	// BastionStatus is the same as Instance with unused fields removed
	out.ID = in.ID
	out.Name = in.Name
	out.SSHKeyName = in.SSHKeyName
	out.State = InstanceState(in.State)
	out.IP = in.IP
	out.FloatingIP = in.FloatingIP
	return nil
}

func Convert_v1alpha6_Bastion_To_v1beta1_Bastion(in *Bastion, out *infrav1.Bastion, s apiconversion.Scope) error {
	err := autoConvert_v1alpha6_Bastion_To_v1beta1_Bastion(in, out, s)
	if err != nil {
		return err
	}

	if in.Instance.ServerGroupID != "" {
		out.Instance.ServerGroup = &infrav1.ServerGroupFilter{ID: in.Instance.ServerGroupID}
	} else {
		out.Instance.ServerGroup = nil
	}

	out.FloatingIP = in.Instance.FloatingIP
	return nil
}

func Convert_v1beta1_Bastion_To_v1alpha6_Bastion(in *infrav1.Bastion, out *Bastion, s apiconversion.Scope) error {
	err := autoConvert_v1beta1_Bastion_To_v1alpha6_Bastion(in, out, s)
	if err != nil {
		return err
	}

	if in.Instance.ServerGroup != nil && in.Instance.ServerGroup.ID != "" {
		out.Instance.ServerGroupID = in.Instance.ServerGroup.ID
	}

	out.Instance.FloatingIP = in.FloatingIP
	return nil
}

/* ManagedSecurityGroups */

func restorev1beta1ManagedSecurityGroups(previous *infrav1.ManagedSecurityGroups, dst *infrav1.ManagedSecurityGroups) {
	dst.AllNodesSecurityGroupRules = previous.AllNodesSecurityGroupRules
}

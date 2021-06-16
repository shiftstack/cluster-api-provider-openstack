package v1alpha1

import (
	"sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha4"
)

func (OpenstackClusterProviderSpec) ToAlpha4() (*v1alpha4.OpenStackClusterSpec, error)

func (OpenstackClusterProviderStatus) ToAlpha4() (*v1alpha4.OpenStackClusterStatus, error)

func (OpenstackProviderSpec) ToAlpha4() (*v1alpha4.OpenStackMachineSpec, error)

func NewAlpha4OpenStackMachine(OpenstackProviderSpec) (*v1alpha4.OpenStackMachine, error)

func newAlpha4OpenStackCluster(OpenstackClusterProviderSpec, *OpenstackClusterProviderStatus) (*v1alpha4.OpenStackCluster, error)

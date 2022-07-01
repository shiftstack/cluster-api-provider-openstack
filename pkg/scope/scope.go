/*
Copyright 2022 The Kubernetes Authors.

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

package scope

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha6"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/clients"
)

// Factory instantiates a new ClientScope using credentials from either a cluster or a machine.
type Factory interface {
	NewClientScopeFromMachine(ctx context.Context, ctrlClient client.Client, openStackMachine *infrav1.OpenStackMachine, logger logr.Logger) (ClientGeneratorScope, error)
	NewClientScopeFromCluster(ctx context.Context, ctrlClient client.Client, openStackCluster *infrav1.OpenStackCluster, logger logr.Logger) (ClientGeneratorScope, error)
}

// ClientGeneratorScope is a Scope that can also be used to create service clients.
type ClientGeneratorScope interface {
	NewComputeClient() (clients.ComputeClient, error)
	NewNetworkClient() (clients.NetworkClient, error)
	NewLbClient() (clients.LbClient, error)
	Scope
}

// Scope contains arguments common to most operations.
type Scope interface {
	Logger() logr.Logger
	ProjectID() string
}

type baseScope struct {
	projectID string
	logger    logr.Logger
}

func NewBaseScope(projectID string, logger logr.Logger) Scope {
	return &baseScope{
		projectID: projectID,
		logger:    logger,
	}
}

func (s *baseScope) Logger() logr.Logger {
	return s.logger
}

func (s *baseScope) ProjectID() string {
	return s.projectID
}

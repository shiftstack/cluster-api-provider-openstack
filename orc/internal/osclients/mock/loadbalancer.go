/*
Copyright 2024 The ORC Authors.

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
// Code generated by MockGen. DO NOT EDIT.
// Source: ../loadbalancer.go
//
// Generated by this command:
//
//	mockgen -package mock -destination=loadbalancer.go -source=../loadbalancer.go github.com/k-orc/openstack-resource-controller/internal/osclients/mock LbClient
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	apiversions "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/apiversions"
	flavors "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/flavors"
	listeners "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/listeners"
	loadbalancers "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/loadbalancers"
	monitors "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/monitors"
	pools "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/pools"
	providers "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/providers"
	gomock "go.uber.org/mock/gomock"
)

// MockLbClient is a mock of LbClient interface.
type MockLbClient struct {
	ctrl     *gomock.Controller
	recorder *MockLbClientMockRecorder
}

// MockLbClientMockRecorder is the mock recorder for MockLbClient.
type MockLbClientMockRecorder struct {
	mock *MockLbClient
}

// NewMockLbClient creates a new mock instance.
func NewMockLbClient(ctrl *gomock.Controller) *MockLbClient {
	mock := &MockLbClient{ctrl: ctrl}
	mock.recorder = &MockLbClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLbClient) EXPECT() *MockLbClientMockRecorder {
	return m.recorder
}

// CreateListener mocks base method.
func (m *MockLbClient) CreateListener(opts listeners.CreateOptsBuilder) (*listeners.Listener, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateListener", opts)
	ret0, _ := ret[0].(*listeners.Listener)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateListener indicates an expected call of CreateListener.
func (mr *MockLbClientMockRecorder) CreateListener(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateListener", reflect.TypeOf((*MockLbClient)(nil).CreateListener), opts)
}

// CreateLoadBalancer mocks base method.
func (m *MockLbClient) CreateLoadBalancer(opts loadbalancers.CreateOptsBuilder) (*loadbalancers.LoadBalancer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLoadBalancer", opts)
	ret0, _ := ret[0].(*loadbalancers.LoadBalancer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateLoadBalancer indicates an expected call of CreateLoadBalancer.
func (mr *MockLbClientMockRecorder) CreateLoadBalancer(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLoadBalancer", reflect.TypeOf((*MockLbClient)(nil).CreateLoadBalancer), opts)
}

// CreateMonitor mocks base method.
func (m *MockLbClient) CreateMonitor(opts monitors.CreateOptsBuilder) (*monitors.Monitor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMonitor", opts)
	ret0, _ := ret[0].(*monitors.Monitor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMonitor indicates an expected call of CreateMonitor.
func (mr *MockLbClientMockRecorder) CreateMonitor(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMonitor", reflect.TypeOf((*MockLbClient)(nil).CreateMonitor), opts)
}

// CreatePool mocks base method.
func (m *MockLbClient) CreatePool(opts pools.CreateOptsBuilder) (*pools.Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePool", opts)
	ret0, _ := ret[0].(*pools.Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePool indicates an expected call of CreatePool.
func (mr *MockLbClientMockRecorder) CreatePool(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePool", reflect.TypeOf((*MockLbClient)(nil).CreatePool), opts)
}

// CreatePoolMember mocks base method.
func (m *MockLbClient) CreatePoolMember(poolID string, opts pools.CreateMemberOptsBuilder) (*pools.Member, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePoolMember", poolID, opts)
	ret0, _ := ret[0].(*pools.Member)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePoolMember indicates an expected call of CreatePoolMember.
func (mr *MockLbClientMockRecorder) CreatePoolMember(poolID, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePoolMember", reflect.TypeOf((*MockLbClient)(nil).CreatePoolMember), poolID, opts)
}

// DeleteListener mocks base method.
func (m *MockLbClient) DeleteListener(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteListener", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteListener indicates an expected call of DeleteListener.
func (mr *MockLbClientMockRecorder) DeleteListener(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteListener", reflect.TypeOf((*MockLbClient)(nil).DeleteListener), id)
}

// DeleteLoadBalancer mocks base method.
func (m *MockLbClient) DeleteLoadBalancer(id string, opts loadbalancers.DeleteOptsBuilder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLoadBalancer", id, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLoadBalancer indicates an expected call of DeleteLoadBalancer.
func (mr *MockLbClientMockRecorder) DeleteLoadBalancer(id, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLoadBalancer", reflect.TypeOf((*MockLbClient)(nil).DeleteLoadBalancer), id, opts)
}

// DeleteMonitor mocks base method.
func (m *MockLbClient) DeleteMonitor(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMonitor", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMonitor indicates an expected call of DeleteMonitor.
func (mr *MockLbClientMockRecorder) DeleteMonitor(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMonitor", reflect.TypeOf((*MockLbClient)(nil).DeleteMonitor), id)
}

// DeletePool mocks base method.
func (m *MockLbClient) DeletePool(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePool", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePool indicates an expected call of DeletePool.
func (mr *MockLbClientMockRecorder) DeletePool(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePool", reflect.TypeOf((*MockLbClient)(nil).DeletePool), id)
}

// DeletePoolMember mocks base method.
func (m *MockLbClient) DeletePoolMember(poolID, lbMemberID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePoolMember", poolID, lbMemberID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePoolMember indicates an expected call of DeletePoolMember.
func (mr *MockLbClientMockRecorder) DeletePoolMember(poolID, lbMemberID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePoolMember", reflect.TypeOf((*MockLbClient)(nil).DeletePoolMember), poolID, lbMemberID)
}

// GetListener mocks base method.
func (m *MockLbClient) GetListener(id string) (*listeners.Listener, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetListener", id)
	ret0, _ := ret[0].(*listeners.Listener)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListener indicates an expected call of GetListener.
func (mr *MockLbClientMockRecorder) GetListener(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListener", reflect.TypeOf((*MockLbClient)(nil).GetListener), id)
}

// GetLoadBalancer mocks base method.
func (m *MockLbClient) GetLoadBalancer(id string) (*loadbalancers.LoadBalancer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLoadBalancer", id)
	ret0, _ := ret[0].(*loadbalancers.LoadBalancer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLoadBalancer indicates an expected call of GetLoadBalancer.
func (mr *MockLbClientMockRecorder) GetLoadBalancer(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLoadBalancer", reflect.TypeOf((*MockLbClient)(nil).GetLoadBalancer), id)
}

// GetPool mocks base method.
func (m *MockLbClient) GetPool(id string) (*pools.Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPool", id)
	ret0, _ := ret[0].(*pools.Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPool indicates an expected call of GetPool.
func (mr *MockLbClientMockRecorder) GetPool(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPool", reflect.TypeOf((*MockLbClient)(nil).GetPool), id)
}

// ListListeners mocks base method.
func (m *MockLbClient) ListListeners(opts listeners.ListOptsBuilder) ([]listeners.Listener, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListListeners", opts)
	ret0, _ := ret[0].([]listeners.Listener)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListListeners indicates an expected call of ListListeners.
func (mr *MockLbClientMockRecorder) ListListeners(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListListeners", reflect.TypeOf((*MockLbClient)(nil).ListListeners), opts)
}

// ListLoadBalancerFlavors mocks base method.
func (m *MockLbClient) ListLoadBalancerFlavors() ([]flavors.Flavor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLoadBalancerFlavors")
	ret0, _ := ret[0].([]flavors.Flavor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLoadBalancerFlavors indicates an expected call of ListLoadBalancerFlavors.
func (mr *MockLbClientMockRecorder) ListLoadBalancerFlavors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLoadBalancerFlavors", reflect.TypeOf((*MockLbClient)(nil).ListLoadBalancerFlavors))
}

// ListLoadBalancerProviders mocks base method.
func (m *MockLbClient) ListLoadBalancerProviders() ([]providers.Provider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLoadBalancerProviders")
	ret0, _ := ret[0].([]providers.Provider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLoadBalancerProviders indicates an expected call of ListLoadBalancerProviders.
func (mr *MockLbClientMockRecorder) ListLoadBalancerProviders() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLoadBalancerProviders", reflect.TypeOf((*MockLbClient)(nil).ListLoadBalancerProviders))
}

// ListLoadBalancers mocks base method.
func (m *MockLbClient) ListLoadBalancers(opts loadbalancers.ListOptsBuilder) ([]loadbalancers.LoadBalancer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLoadBalancers", opts)
	ret0, _ := ret[0].([]loadbalancers.LoadBalancer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLoadBalancers indicates an expected call of ListLoadBalancers.
func (mr *MockLbClientMockRecorder) ListLoadBalancers(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLoadBalancers", reflect.TypeOf((*MockLbClient)(nil).ListLoadBalancers), opts)
}

// ListMonitors mocks base method.
func (m *MockLbClient) ListMonitors(opts monitors.ListOptsBuilder) ([]monitors.Monitor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMonitors", opts)
	ret0, _ := ret[0].([]monitors.Monitor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMonitors indicates an expected call of ListMonitors.
func (mr *MockLbClientMockRecorder) ListMonitors(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMonitors", reflect.TypeOf((*MockLbClient)(nil).ListMonitors), opts)
}

// ListOctaviaVersions mocks base method.
func (m *MockLbClient) ListOctaviaVersions() ([]apiversions.APIVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOctaviaVersions")
	ret0, _ := ret[0].([]apiversions.APIVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOctaviaVersions indicates an expected call of ListOctaviaVersions.
func (mr *MockLbClientMockRecorder) ListOctaviaVersions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOctaviaVersions", reflect.TypeOf((*MockLbClient)(nil).ListOctaviaVersions))
}

// ListPoolMember mocks base method.
func (m *MockLbClient) ListPoolMember(poolID string, opts pools.ListMembersOptsBuilder) ([]pools.Member, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPoolMember", poolID, opts)
	ret0, _ := ret[0].([]pools.Member)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPoolMember indicates an expected call of ListPoolMember.
func (mr *MockLbClientMockRecorder) ListPoolMember(poolID, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPoolMember", reflect.TypeOf((*MockLbClient)(nil).ListPoolMember), poolID, opts)
}

// ListPools mocks base method.
func (m *MockLbClient) ListPools(opts pools.ListOptsBuilder) ([]pools.Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPools", opts)
	ret0, _ := ret[0].([]pools.Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPools indicates an expected call of ListPools.
func (mr *MockLbClientMockRecorder) ListPools(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPools", reflect.TypeOf((*MockLbClient)(nil).ListPools), opts)
}

// UpdateListener mocks base method.
func (m *MockLbClient) UpdateListener(id string, opts listeners.UpdateOpts) (*listeners.Listener, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateListener", id, opts)
	ret0, _ := ret[0].(*listeners.Listener)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateListener indicates an expected call of UpdateListener.
func (mr *MockLbClientMockRecorder) UpdateListener(id, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateListener", reflect.TypeOf((*MockLbClient)(nil).UpdateListener), id, opts)
}
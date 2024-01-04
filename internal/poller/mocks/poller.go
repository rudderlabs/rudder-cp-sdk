// Code generated by MockGen. DO NOT EDIT.
// Source: poller.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	modelv2 "github.com/rudderlabs/rudder-control-plane-sdk/modelv2"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetUpdatedWorkspaceConfigs mocks base method.
func (m *MockClient) GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAt time.Time) (*modelv2.WorkspaceConfigs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatedWorkspaceConfigs", ctx, updatedAt)
	ret0, _ := ret[0].(*modelv2.WorkspaceConfigs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUpdatedWorkspaceConfigs indicates an expected call of GetUpdatedWorkspaceConfigs.
func (mr *MockClientMockRecorder) GetUpdatedWorkspaceConfigs(ctx, updatedAt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatedWorkspaceConfigs", reflect.TypeOf((*MockClient)(nil).GetUpdatedWorkspaceConfigs), ctx, updatedAt)
}

// GetWorkspaceConfigs mocks base method.
func (m *MockClient) GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkspaceConfigs", ctx)
	ret0, _ := ret[0].(*modelv2.WorkspaceConfigs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWorkspaceConfigs indicates an expected call of GetWorkspaceConfigs.
func (mr *MockClientMockRecorder) GetWorkspaceConfigs(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkspaceConfigs", reflect.TypeOf((*MockClient)(nil).GetWorkspaceConfigs), ctx)
}

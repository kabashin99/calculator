// Code generated by MockGen. DO NOT EDIT.
// Source: calculator_app/internal/proto (interfaces: OrchestratorServiceClient)

// Package mock_proto is a generated GoMock package.
package mocks

import (
	proto "calculator_app/internal/proto"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockOrchestratorServiceClient is a mock of OrchestratorServiceClient interface.
type MockOrchestratorServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockOrchestratorServiceClientMockRecorder
}

// MockOrchestratorServiceClientMockRecorder is the mock recorder for MockOrchestratorServiceClient.
type MockOrchestratorServiceClientMockRecorder struct {
	mock *MockOrchestratorServiceClient
}

// NewMockOrchestratorServiceClient creates a new mock instance.
func NewMockOrchestratorServiceClient(ctrl *gomock.Controller) *MockOrchestratorServiceClient {
	mock := &MockOrchestratorServiceClient{ctrl: ctrl}
	mock.recorder = &MockOrchestratorServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrchestratorServiceClient) EXPECT() *MockOrchestratorServiceClientMockRecorder {
	return m.recorder
}

// GetTask mocks base method.
func (m *MockOrchestratorServiceClient) GetTask(arg0 context.Context, arg1 *proto.GetTaskRequest, arg2 ...grpc.CallOption) (*proto.GetTaskResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetTask", varargs...)
	ret0, _ := ret[0].(*proto.GetTaskResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTask indicates an expected call of GetTask.
func (mr *MockOrchestratorServiceClientMockRecorder) GetTask(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTask", reflect.TypeOf((*MockOrchestratorServiceClient)(nil).GetTask), varargs...)
}

// GetTaskResult mocks base method.
func (m *MockOrchestratorServiceClient) GetTaskResult(arg0 context.Context, arg1 *proto.GetTaskResultRequest, arg2 ...grpc.CallOption) (*proto.GetTaskResultResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetTaskResult", varargs...)
	ret0, _ := ret[0].(*proto.GetTaskResultResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTaskResult indicates an expected call of GetTaskResult.
func (mr *MockOrchestratorServiceClientMockRecorder) GetTaskResult(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskResult", reflect.TypeOf((*MockOrchestratorServiceClient)(nil).GetTaskResult), varargs...)
}

// SubmitResult mocks base method.
func (m *MockOrchestratorServiceClient) SubmitResult(arg0 context.Context, arg1 *proto.SubmitResultRequest, arg2 ...grpc.CallOption) (*proto.SubmitResultResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SubmitResult", varargs...)
	ret0, _ := ret[0].(*proto.SubmitResultResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubmitResult indicates an expected call of SubmitResult.
func (mr *MockOrchestratorServiceClientMockRecorder) SubmitResult(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitResult", reflect.TypeOf((*MockOrchestratorServiceClient)(nil).SubmitResult), varargs...)
}

// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

// Code generated by MockGen. DO NOT EDIT.
// Source: go.pinniped.dev/internal/issuer (interfaces: ClientCertIssuer)

// Package issuermocks is a generated GoMock package.
package issuermocks

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockClientCertIssuer is a mock of ClientCertIssuer interface.
type MockClientCertIssuer struct {
	ctrl     *gomock.Controller
	recorder *MockClientCertIssuerMockRecorder
}

// MockClientCertIssuerMockRecorder is the mock recorder for MockClientCertIssuer.
type MockClientCertIssuerMockRecorder struct {
	mock *MockClientCertIssuer
}

// NewMockClientCertIssuer creates a new mock instance.
func NewMockClientCertIssuer(ctrl *gomock.Controller) *MockClientCertIssuer {
	mock := &MockClientCertIssuer{ctrl: ctrl}
	mock.recorder = &MockClientCertIssuerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientCertIssuer) EXPECT() *MockClientCertIssuerMockRecorder {
	return m.recorder
}

// IssueClientCertPEM mocks base method.
func (m *MockClientCertIssuer) IssueClientCertPEM(arg0 string, arg1 []string, arg2 time.Duration) ([]byte, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IssueClientCertPEM", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// IssueClientCertPEM indicates an expected call of IssueClientCertPEM.
func (mr *MockClientCertIssuerMockRecorder) IssueClientCertPEM(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IssueClientCertPEM", reflect.TypeOf((*MockClientCertIssuer)(nil).IssueClientCertPEM), arg0, arg1, arg2)
}

// Name mocks base method.
func (m *MockClientCertIssuer) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockClientCertIssuerMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockClientCertIssuer)(nil).Name))
}

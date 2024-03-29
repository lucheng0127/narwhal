// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/protocol/protocol.go

// Package mock_protocol is a generated GoMock package.
package mock_protocol

import (
	net "net"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	protocol "github.com/lucheng0127/narwhal/pkg/protocol"
)

// MockPKG is a mock of PKG interface.
type MockPKG struct {
	ctrl     *gomock.Controller
	recorder *MockPKGMockRecorder
}

// MockPKGMockRecorder is the mock recorder for MockPKG.
type MockPKGMockRecorder struct {
	mock *MockPKG
}

// NewMockPKG creates a new mock instance.
func NewMockPKG(ctrl *gomock.Controller) *MockPKG {
	mock := &MockPKG{ctrl: ctrl}
	mock.recorder = &MockPKGMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPKG) EXPECT() *MockPKGMockRecorder {
	return m.recorder
}

// Encode mocks base method.
func (m *MockPKG) Encode() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encode")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Encode indicates an expected call of Encode.
func (mr *MockPKGMockRecorder) Encode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encode", reflect.TypeOf((*MockPKG)(nil).Encode))
}

// GetPCode mocks base method.
func (m *MockPKG) GetPCode() byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPCode")
	ret0, _ := ret[0].(byte)
	return ret0
}

// GetPCode indicates an expected call of GetPCode.
func (mr *MockPKGMockRecorder) GetPCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPCode", reflect.TypeOf((*MockPKG)(nil).GetPCode))
}

// GetPayload mocks base method.
func (m *MockPKG) GetPayload() protocol.PL {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayload")
	ret0, _ := ret[0].(protocol.PL)
	return ret0
}

// GetPayload indicates an expected call of GetPayload.
func (mr *MockPKGMockRecorder) GetPayload() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayload", reflect.TypeOf((*MockPKG)(nil).GetPayload))
}

// SendToConn mocks base method.
func (m *MockPKG) SendToConn(conn net.Conn) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendToConn", conn)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendToConn indicates an expected call of SendToConn.
func (mr *MockPKGMockRecorder) SendToConn(conn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendToConn", reflect.TypeOf((*MockPKG)(nil).SendToConn), conn)
}

// MockPL is a mock of PL interface.
type MockPL struct {
	ctrl     *gomock.Controller
	recorder *MockPLMockRecorder
}

// MockPLMockRecorder is the mock recorder for MockPL.
type MockPLMockRecorder struct {
	mock *MockPL
}

// NewMockPL creates a new mock instance.
func NewMockPL(ctrl *gomock.Controller) *MockPL {
	mock := &MockPL{ctrl: ctrl}
	mock.recorder = &MockPLMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPL) EXPECT() *MockPLMockRecorder {
	return m.recorder
}

// Int mocks base method.
func (m *MockPL) Int() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Int")
	ret0, _ := ret[0].(int)
	return ret0
}

// Int indicates an expected call of Int.
func (mr *MockPLMockRecorder) Int() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Int", reflect.TypeOf((*MockPL)(nil).Int))
}

// String mocks base method.
func (m *MockPL) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockPLMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockPL)(nil).String))
}

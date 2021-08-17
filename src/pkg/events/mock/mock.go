// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hr3lxphr6j/bililive-go/src/pkg/events (interfaces: Dispatcher)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	events "github.com/hr3lxphr6j/bililive-go/src/pkg/events"
)

// MockDispatcher is a mock of Dispatcher interface.
type MockDispatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDispatcherMockRecorder
}

// MockDispatcherMockRecorder is the mock recorder for MockDispatcher.
type MockDispatcherMockRecorder struct {
	mock *MockDispatcher
}

// NewMockDispatcher creates a new mock instance.
func NewMockDispatcher(ctrl *gomock.Controller) *MockDispatcher {
	mock := &MockDispatcher{ctrl: ctrl}
	mock.recorder = &MockDispatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDispatcher) EXPECT() *MockDispatcherMockRecorder {
	return m.recorder
}

// AddEventListener mocks base method.
func (m *MockDispatcher) AddEventListener(arg0 events.EventType, arg1 *events.EventListener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddEventListener", arg0, arg1)
}

// AddEventListener indicates an expected call of AddEventListener.
func (mr *MockDispatcherMockRecorder) AddEventListener(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEventListener", reflect.TypeOf((*MockDispatcher)(nil).AddEventListener), arg0, arg1)
}

// Close mocks base method.
func (m *MockDispatcher) Close(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close", arg0)
}

// Close indicates an expected call of Close.
func (mr *MockDispatcherMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDispatcher)(nil).Close), arg0)
}

// DispatchEvent mocks base method.
func (m *MockDispatcher) DispatchEvent(arg0 *events.Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DispatchEvent", arg0)
}

// DispatchEvent indicates an expected call of DispatchEvent.
func (mr *MockDispatcherMockRecorder) DispatchEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DispatchEvent", reflect.TypeOf((*MockDispatcher)(nil).DispatchEvent), arg0)
}

// RemoveAllEventListener mocks base method.
func (m *MockDispatcher) RemoveAllEventListener(arg0 events.EventType) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveAllEventListener", arg0)
}

// RemoveAllEventListener indicates an expected call of RemoveAllEventListener.
func (mr *MockDispatcherMockRecorder) RemoveAllEventListener(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAllEventListener", reflect.TypeOf((*MockDispatcher)(nil).RemoveAllEventListener), arg0)
}

// RemoveEventListener mocks base method.
func (m *MockDispatcher) RemoveEventListener(arg0 events.EventType, arg1 *events.EventListener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveEventListener", arg0, arg1)
}

// RemoveEventListener indicates an expected call of RemoveEventListener.
func (mr *MockDispatcherMockRecorder) RemoveEventListener(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveEventListener", reflect.TypeOf((*MockDispatcher)(nil).RemoveEventListener), arg0, arg1)
}

// Start mocks base method.
func (m *MockDispatcher) Start(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockDispatcherMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockDispatcher)(nil).Start), arg0)
}

package mock_xkcd

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	core "yadro.com/course/update/core"
)

// MockXKCD is a mock of XKCD interface.
type MockXKCD struct {
	ctrl     *gomock.Controller
	recorder *MockXKCDMockRecorder
}

// MockXKCDMockRecorder is the mock recorder for MockXKCD.
type MockXKCDMockRecorder struct {
	mock *MockXKCD
}

// NewMockXKCD creates a new mock instance.
func NewMockXKCD(ctrl *gomock.Controller) *MockXKCD {
	mock := &MockXKCD{ctrl: ctrl}
	mock.recorder = &MockXKCDMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockXKCD) EXPECT() *MockXKCDMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockXKCD) Get(arg0 context.Context, arg1 int) (core.XKCDInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(core.XKCDInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockXKCDMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockXKCD)(nil).Get), arg0, arg1)
}

// LastID mocks base method.
func (m *MockXKCD) LastID(arg0 context.Context) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastID", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastID indicates an expected call of LastID.
func (mr *MockXKCDMockRecorder) LastID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastID", reflect.TypeOf((*MockXKCD)(nil).LastID), arg0)
}

// MissingIds mocks base method.
func (m *MockXKCD) MissingIds(arg0 context.Context) []int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MissingIds", arg0)
	ret0, _ := ret[0].([]int)
	return ret0
}

// MissingIds indicates an expected call of MissingIds.
func (mr *MockXKCDMockRecorder) MissingIds(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MissingIds", reflect.TypeOf((*MockXKCD)(nil).MissingIds), arg0)
}

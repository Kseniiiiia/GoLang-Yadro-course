package mock_words

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

type MockWords struct {
	ctrl     *gomock.Controller
	recorder *MockWordsMockRecorder
}

// MockWordsMockRecorder is the mock recorder for MockWords.
type MockWordsMockRecorder struct {
	mock *MockWords
}

// NewMockWords creates a new mock instance.
func NewMockWords(ctrl *gomock.Controller) *MockWords {
	mock := &MockWords{ctrl: ctrl}
	mock.recorder = &MockWordsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWords) EXPECT() *MockWordsMockRecorder {
	return m.recorder
}

// Norm mocks base method.
func (m *MockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Norm", ctx, phrase)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Norm indicates an expected call of Norm.
func (mr *MockWordsMockRecorder) Norm(ctx, phrase interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Norm", reflect.TypeOf((*MockWords)(nil).Norm), ctx, phrase)
}

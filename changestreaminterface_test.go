// Code generated by mockery. DO NOT EDIT.

package queue

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ChangeStreamInterfaceMock is an autogenerated mock type for the ChangeStreamInterface type
type ChangeStreamInterfaceMock struct {
	mock.Mock
}

type ChangeStreamInterfaceMock_Expecter struct {
	mock *mock.Mock
}

func (_m *ChangeStreamInterfaceMock) EXPECT() *ChangeStreamInterfaceMock_Expecter {
	return &ChangeStreamInterfaceMock_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields: ctx
func (_m *ChangeStreamInterfaceMock) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChangeStreamInterfaceMock_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type ChangeStreamInterfaceMock_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
//   - ctx context.Context
func (_e *ChangeStreamInterfaceMock_Expecter) Close(ctx interface{}) *ChangeStreamInterfaceMock_Close_Call {
	return &ChangeStreamInterfaceMock_Close_Call{Call: _e.mock.On("Close", ctx)}
}

func (_c *ChangeStreamInterfaceMock_Close_Call) Run(run func(ctx context.Context)) *ChangeStreamInterfaceMock_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *ChangeStreamInterfaceMock_Close_Call) Return(_a0 error) *ChangeStreamInterfaceMock_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChangeStreamInterfaceMock_Close_Call) RunAndReturn(run func(context.Context) error) *ChangeStreamInterfaceMock_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Decode provides a mock function with given fields: v
func (_m *ChangeStreamInterfaceMock) Decode(v interface{}) error {
	ret := _m.Called(v)

	if len(ret) == 0 {
		panic("no return value specified for Decode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChangeStreamInterfaceMock_Decode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Decode'
type ChangeStreamInterfaceMock_Decode_Call struct {
	*mock.Call
}

// Decode is a helper method to define mock.On call
//   - v interface{}
func (_e *ChangeStreamInterfaceMock_Expecter) Decode(v interface{}) *ChangeStreamInterfaceMock_Decode_Call {
	return &ChangeStreamInterfaceMock_Decode_Call{Call: _e.mock.On("Decode", v)}
}

func (_c *ChangeStreamInterfaceMock_Decode_Call) Run(run func(v interface{})) *ChangeStreamInterfaceMock_Decode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *ChangeStreamInterfaceMock_Decode_Call) Return(_a0 error) *ChangeStreamInterfaceMock_Decode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChangeStreamInterfaceMock_Decode_Call) RunAndReturn(run func(interface{}) error) *ChangeStreamInterfaceMock_Decode_Call {
	_c.Call.Return(run)
	return _c
}

// Next provides a mock function with given fields: ctx
func (_m *ChangeStreamInterfaceMock) Next(ctx context.Context) bool {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ChangeStreamInterfaceMock_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type ChangeStreamInterfaceMock_Next_Call struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
//   - ctx context.Context
func (_e *ChangeStreamInterfaceMock_Expecter) Next(ctx interface{}) *ChangeStreamInterfaceMock_Next_Call {
	return &ChangeStreamInterfaceMock_Next_Call{Call: _e.mock.On("Next", ctx)}
}

func (_c *ChangeStreamInterfaceMock_Next_Call) Run(run func(ctx context.Context)) *ChangeStreamInterfaceMock_Next_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *ChangeStreamInterfaceMock_Next_Call) Return(_a0 bool) *ChangeStreamInterfaceMock_Next_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChangeStreamInterfaceMock_Next_Call) RunAndReturn(run func(context.Context) bool) *ChangeStreamInterfaceMock_Next_Call {
	_c.Call.Return(run)
	return _c
}

// NewChangeStreamInterfaceMock creates a new instance of ChangeStreamInterfaceMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewChangeStreamInterfaceMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *ChangeStreamInterfaceMock {
	mock := &ChangeStreamInterfaceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

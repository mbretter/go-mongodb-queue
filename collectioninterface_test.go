// Code generated by mockery. DO NOT EDIT.

package queue

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	mongo "go.mongodb.org/mongo-driver/mongo"

	options "go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionInterfaceMock is an autogenerated mock type for the CollectionInterface type
type CollectionInterfaceMock struct {
	mock.Mock
}

type CollectionInterfaceMock_Expecter struct {
	mock *mock.Mock
}

func (_m *CollectionInterfaceMock) EXPECT() *CollectionInterfaceMock_Expecter {
	return &CollectionInterfaceMock_Expecter{mock: &_m.Mock}
}

// FindOneAndUpdate provides a mock function with given fields: ctx, filter, update, opts
func (_m *CollectionInterfaceMock) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, filter, update)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for FindOneAndUpdate")
	}

	var r0 *mongo.SingleResult
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}, ...*options.FindOneAndUpdateOptions) *mongo.SingleResult); ok {
		r0 = rf(ctx, filter, update, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.SingleResult)
		}
	}

	return r0
}

// CollectionInterfaceMock_FindOneAndUpdate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindOneAndUpdate'
type CollectionInterfaceMock_FindOneAndUpdate_Call struct {
	*mock.Call
}

// FindOneAndUpdate is a helper method to define mock.On call
//   - ctx context.Context
//   - filter interface{}
//   - update interface{}
//   - opts ...*options.FindOneAndUpdateOptions
func (_e *CollectionInterfaceMock_Expecter) FindOneAndUpdate(ctx interface{}, filter interface{}, update interface{}, opts ...interface{}) *CollectionInterfaceMock_FindOneAndUpdate_Call {
	return &CollectionInterfaceMock_FindOneAndUpdate_Call{Call: _e.mock.On("FindOneAndUpdate",
		append([]interface{}{ctx, filter, update}, opts...)...)}
}

func (_c *CollectionInterfaceMock_FindOneAndUpdate_Call) Run(run func(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions)) *CollectionInterfaceMock_FindOneAndUpdate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]*options.FindOneAndUpdateOptions, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(*options.FindOneAndUpdateOptions)
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), args[2].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *CollectionInterfaceMock_FindOneAndUpdate_Call) Return(_a0 *mongo.SingleResult) *CollectionInterfaceMock_FindOneAndUpdate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CollectionInterfaceMock_FindOneAndUpdate_Call) RunAndReturn(run func(context.Context, interface{}, interface{}, ...*options.FindOneAndUpdateOptions) *mongo.SingleResult) *CollectionInterfaceMock_FindOneAndUpdate_Call {
	_c.Call.Return(run)
	return _c
}

// Indexes provides a mock function with given fields:
func (_m *CollectionInterfaceMock) Indexes() mongo.IndexView {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Indexes")
	}

	var r0 mongo.IndexView
	if rf, ok := ret.Get(0).(func() mongo.IndexView); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(mongo.IndexView)
	}

	return r0
}

// CollectionInterfaceMock_Indexes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Indexes'
type CollectionInterfaceMock_Indexes_Call struct {
	*mock.Call
}

// Indexes is a helper method to define mock.On call
func (_e *CollectionInterfaceMock_Expecter) Indexes() *CollectionInterfaceMock_Indexes_Call {
	return &CollectionInterfaceMock_Indexes_Call{Call: _e.mock.On("Indexes")}
}

func (_c *CollectionInterfaceMock_Indexes_Call) Run(run func()) *CollectionInterfaceMock_Indexes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *CollectionInterfaceMock_Indexes_Call) Return(_a0 mongo.IndexView) *CollectionInterfaceMock_Indexes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CollectionInterfaceMock_Indexes_Call) RunAndReturn(run func() mongo.IndexView) *CollectionInterfaceMock_Indexes_Call {
	_c.Call.Return(run)
	return _c
}

// InsertOne provides a mock function with given fields: ctx, document, opts
func (_m *CollectionInterfaceMock) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, document)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for InsertOne")
	}

	var r0 *mongo.InsertOneResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)); ok {
		return rf(ctx, document, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, ...*options.InsertOneOptions) *mongo.InsertOneResult); ok {
		r0 = rf(ctx, document, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.InsertOneResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, ...*options.InsertOneOptions) error); ok {
		r1 = rf(ctx, document, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CollectionInterfaceMock_InsertOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertOne'
type CollectionInterfaceMock_InsertOne_Call struct {
	*mock.Call
}

// InsertOne is a helper method to define mock.On call
//   - ctx context.Context
//   - document interface{}
//   - opts ...*options.InsertOneOptions
func (_e *CollectionInterfaceMock_Expecter) InsertOne(ctx interface{}, document interface{}, opts ...interface{}) *CollectionInterfaceMock_InsertOne_Call {
	return &CollectionInterfaceMock_InsertOne_Call{Call: _e.mock.On("InsertOne",
		append([]interface{}{ctx, document}, opts...)...)}
}

func (_c *CollectionInterfaceMock_InsertOne_Call) Run(run func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions)) *CollectionInterfaceMock_InsertOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]*options.InsertOneOptions, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(*options.InsertOneOptions)
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *CollectionInterfaceMock_InsertOne_Call) Return(_a0 *mongo.InsertOneResult, _a1 error) *CollectionInterfaceMock_InsertOne_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CollectionInterfaceMock_InsertOne_Call) RunAndReturn(run func(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)) *CollectionInterfaceMock_InsertOne_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateMany provides a mock function with given fields: ctx, filter, update, opts
func (_m *CollectionInterfaceMock) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, filter, update)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateMany")
	}

	var r0 *mongo.UpdateResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)); ok {
		return rf(ctx, filter, update, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) *mongo.UpdateResult); ok {
		r0 = rf(ctx, filter, update, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.UpdateResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) error); ok {
		r1 = rf(ctx, filter, update, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CollectionInterfaceMock_UpdateMany_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateMany'
type CollectionInterfaceMock_UpdateMany_Call struct {
	*mock.Call
}

// UpdateMany is a helper method to define mock.On call
//   - ctx context.Context
//   - filter interface{}
//   - update interface{}
//   - opts ...*options.UpdateOptions
func (_e *CollectionInterfaceMock_Expecter) UpdateMany(ctx interface{}, filter interface{}, update interface{}, opts ...interface{}) *CollectionInterfaceMock_UpdateMany_Call {
	return &CollectionInterfaceMock_UpdateMany_Call{Call: _e.mock.On("UpdateMany",
		append([]interface{}{ctx, filter, update}, opts...)...)}
}

func (_c *CollectionInterfaceMock_UpdateMany_Call) Run(run func(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions)) *CollectionInterfaceMock_UpdateMany_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]*options.UpdateOptions, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(*options.UpdateOptions)
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), args[2].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *CollectionInterfaceMock_UpdateMany_Call) Return(_a0 *mongo.UpdateResult, _a1 error) *CollectionInterfaceMock_UpdateMany_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CollectionInterfaceMock_UpdateMany_Call) RunAndReturn(run func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)) *CollectionInterfaceMock_UpdateMany_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateOne provides a mock function with given fields: ctx, filter, update, opts
func (_m *CollectionInterfaceMock) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, filter, update)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateOne")
	}

	var r0 *mongo.UpdateResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)); ok {
		return rf(ctx, filter, update, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) *mongo.UpdateResult); ok {
		r0 = rf(ctx, filter, update, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.UpdateResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) error); ok {
		r1 = rf(ctx, filter, update, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CollectionInterfaceMock_UpdateOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateOne'
type CollectionInterfaceMock_UpdateOne_Call struct {
	*mock.Call
}

// UpdateOne is a helper method to define mock.On call
//   - ctx context.Context
//   - filter interface{}
//   - update interface{}
//   - opts ...*options.UpdateOptions
func (_e *CollectionInterfaceMock_Expecter) UpdateOne(ctx interface{}, filter interface{}, update interface{}, opts ...interface{}) *CollectionInterfaceMock_UpdateOne_Call {
	return &CollectionInterfaceMock_UpdateOne_Call{Call: _e.mock.On("UpdateOne",
		append([]interface{}{ctx, filter, update}, opts...)...)}
}

func (_c *CollectionInterfaceMock_UpdateOne_Call) Run(run func(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions)) *CollectionInterfaceMock_UpdateOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]*options.UpdateOptions, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(*options.UpdateOptions)
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), args[2].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *CollectionInterfaceMock_UpdateOne_Call) Return(_a0 *mongo.UpdateResult, _a1 error) *CollectionInterfaceMock_UpdateOne_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CollectionInterfaceMock_UpdateOne_Call) RunAndReturn(run func(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)) *CollectionInterfaceMock_UpdateOne_Call {
	_c.Call.Return(run)
	return _c
}

// Watch provides a mock function with given fields: ctx, pipeline, opts
func (_m *CollectionInterfaceMock) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, pipeline)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Watch")
	}

	var r0 *mongo.ChangeStream
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)); ok {
		return rf(ctx, pipeline, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, ...*options.ChangeStreamOptions) *mongo.ChangeStream); ok {
		r0 = rf(ctx, pipeline, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.ChangeStream)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, ...*options.ChangeStreamOptions) error); ok {
		r1 = rf(ctx, pipeline, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CollectionInterfaceMock_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type CollectionInterfaceMock_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - ctx context.Context
//   - pipeline interface{}
//   - opts ...*options.ChangeStreamOptions
func (_e *CollectionInterfaceMock_Expecter) Watch(ctx interface{}, pipeline interface{}, opts ...interface{}) *CollectionInterfaceMock_Watch_Call {
	return &CollectionInterfaceMock_Watch_Call{Call: _e.mock.On("Watch",
		append([]interface{}{ctx, pipeline}, opts...)...)}
}

func (_c *CollectionInterfaceMock_Watch_Call) Run(run func(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions)) *CollectionInterfaceMock_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]*options.ChangeStreamOptions, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(*options.ChangeStreamOptions)
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *CollectionInterfaceMock_Watch_Call) Return(_a0 *mongo.ChangeStream, _a1 error) *CollectionInterfaceMock_Watch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CollectionInterfaceMock_Watch_Call) RunAndReturn(run func(context.Context, interface{}, ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)) *CollectionInterfaceMock_Watch_Call {
	_c.Call.Return(run)
	return _c
}

// NewCollectionInterfaceMock creates a new instance of CollectionInterfaceMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCollectionInterfaceMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *CollectionInterfaceMock {
	mock := &CollectionInterfaceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

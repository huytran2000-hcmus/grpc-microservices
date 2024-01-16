// Code generated by mockery v2.38.0. DO NOT EDIT.

package ports

import (
	context "context"

	domain "github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
	mock "github.com/stretchr/testify/mock"
)

// MockAPIPort is an autogenerated mock type for the APIPort type
type MockAPIPort struct {
	mock.Mock
}

type MockAPIPort_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAPIPort) EXPECT() *MockAPIPort_Expecter {
	return &MockAPIPort_Expecter{mock: &_m.Mock}
}

// GetOrder provides a mock function with given fields: ctx, id
func (_m *MockAPIPort) GetOrder(ctx context.Context, id int64) (domain.Order, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetOrder")
	}

	var r0 domain.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (domain.Order, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) domain.Order); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(domain.Order)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAPIPort_GetOrder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOrder'
type MockAPIPort_GetOrder_Call struct {
	*mock.Call
}

// GetOrder is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
func (_e *MockAPIPort_Expecter) GetOrder(ctx interface{}, id interface{}) *MockAPIPort_GetOrder_Call {
	return &MockAPIPort_GetOrder_Call{Call: _e.mock.On("GetOrder", ctx, id)}
}

func (_c *MockAPIPort_GetOrder_Call) Run(run func(ctx context.Context, id int64)) *MockAPIPort_GetOrder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *MockAPIPort_GetOrder_Call) Return(_a0 domain.Order, _a1 error) *MockAPIPort_GetOrder_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAPIPort_GetOrder_Call) RunAndReturn(run func(context.Context, int64) (domain.Order, error)) *MockAPIPort_GetOrder_Call {
	_c.Call.Return(run)
	return _c
}

// PlaceOrder provides a mock function with given fields: ctx, order
func (_m *MockAPIPort) PlaceOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	ret := _m.Called(ctx, order)

	if len(ret) == 0 {
		panic("no return value specified for PlaceOrder")
	}

	var r0 domain.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.Order) (domain.Order, error)); ok {
		return rf(ctx, order)
	}
	if rf, ok := ret.Get(0).(func(context.Context, domain.Order) domain.Order); ok {
		r0 = rf(ctx, order)
	} else {
		r0 = ret.Get(0).(domain.Order)
	}

	if rf, ok := ret.Get(1).(func(context.Context, domain.Order) error); ok {
		r1 = rf(ctx, order)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAPIPort_PlaceOrder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PlaceOrder'
type MockAPIPort_PlaceOrder_Call struct {
	*mock.Call
}

// PlaceOrder is a helper method to define mock.On call
//   - ctx context.Context
//   - order domain.Order
func (_e *MockAPIPort_Expecter) PlaceOrder(ctx interface{}, order interface{}) *MockAPIPort_PlaceOrder_Call {
	return &MockAPIPort_PlaceOrder_Call{Call: _e.mock.On("PlaceOrder", ctx, order)}
}

func (_c *MockAPIPort_PlaceOrder_Call) Run(run func(ctx context.Context, order domain.Order)) *MockAPIPort_PlaceOrder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.Order))
	})
	return _c
}

func (_c *MockAPIPort_PlaceOrder_Call) Return(_a0 domain.Order, _a1 error) *MockAPIPort_PlaceOrder_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAPIPort_PlaceOrder_Call) RunAndReturn(run func(context.Context, domain.Order) (domain.Order, error)) *MockAPIPort_PlaceOrder_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAPIPort creates a new instance of MockAPIPort. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAPIPort(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAPIPort {
	mock := &MockAPIPort{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

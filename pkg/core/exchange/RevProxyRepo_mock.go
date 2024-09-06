// Code generated by mockery v2.45.0. DO NOT EDIT.

//go:build !compile

package exchange

import mock "github.com/stretchr/testify/mock"

// MockRevProxyRepo is an autogenerated mock type for the RevProxyRepo type
type MockRevProxyRepo struct {
	mock.Mock
}

type MockRevProxyRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRevProxyRepo) EXPECT() *MockRevProxyRepo_Expecter {
	return &MockRevProxyRepo_Expecter{mock: &_m.Mock}
}

// Find provides a mock function with given fields: nameSpace
func (_m *MockRevProxyRepo) Find(nameSpace string) (*RevProxy, error) {
	ret := _m.Called(nameSpace)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 *RevProxy
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*RevProxy, error)); ok {
		return rf(nameSpace)
	}
	if rf, ok := ret.Get(0).(func(string) *RevProxy); ok {
		r0 = rf(nameSpace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*RevProxy)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(nameSpace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRevProxyRepo_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type MockRevProxyRepo_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - nameSpace string
func (_e *MockRevProxyRepo_Expecter) Find(nameSpace interface{}) *MockRevProxyRepo_Find_Call {
	return &MockRevProxyRepo_Find_Call{Call: _e.mock.On("Find", nameSpace)}
}

func (_c *MockRevProxyRepo_Find_Call) Run(run func(nameSpace string)) *MockRevProxyRepo_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockRevProxyRepo_Find_Call) Return(_a0 *RevProxy, _a1 error) *MockRevProxyRepo_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRevProxyRepo_Find_Call) RunAndReturn(run func(string) (*RevProxy, error)) *MockRevProxyRepo_Find_Call {
	_c.Call.Return(run)
	return _c
}

// Register provides a mock function with given fields: proxy
func (_m *MockRevProxyRepo) Register(proxy *RevProxy) {
	_m.Called(proxy)
}

// MockRevProxyRepo_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type MockRevProxyRepo_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - proxy *RevProxy
func (_e *MockRevProxyRepo_Expecter) Register(proxy interface{}) *MockRevProxyRepo_Register_Call {
	return &MockRevProxyRepo_Register_Call{Call: _e.mock.On("Register", proxy)}
}

func (_c *MockRevProxyRepo_Register_Call) Run(run func(proxy *RevProxy)) *MockRevProxyRepo_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*RevProxy))
	})
	return _c
}

func (_c *MockRevProxyRepo_Register_Call) Return() *MockRevProxyRepo_Register_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockRevProxyRepo_Register_Call) RunAndReturn(run func(*RevProxy)) *MockRevProxyRepo_Register_Call {
	_c.Call.Return(run)
	return _c
}

// Unregister provides a mock function with given fields: proxy
func (_m *MockRevProxyRepo) Unregister(proxy *RevProxy) {
	_m.Called(proxy)
}

// MockRevProxyRepo_Unregister_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Unregister'
type MockRevProxyRepo_Unregister_Call struct {
	*mock.Call
}

// Unregister is a helper method to define mock.On call
//   - proxy *RevProxy
func (_e *MockRevProxyRepo_Expecter) Unregister(proxy interface{}) *MockRevProxyRepo_Unregister_Call {
	return &MockRevProxyRepo_Unregister_Call{Call: _e.mock.On("Unregister", proxy)}
}

func (_c *MockRevProxyRepo_Unregister_Call) Run(run func(proxy *RevProxy)) *MockRevProxyRepo_Unregister_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*RevProxy))
	})
	return _c
}

func (_c *MockRevProxyRepo_Unregister_Call) Return() *MockRevProxyRepo_Unregister_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockRevProxyRepo_Unregister_Call) RunAndReturn(run func(*RevProxy)) *MockRevProxyRepo_Unregister_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRevProxyRepo creates a new instance of MockRevProxyRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRevProxyRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRevProxyRepo {
	mock := &MockRevProxyRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
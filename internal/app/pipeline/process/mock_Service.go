// Code generated by mockery v1.0.0. DO NOT EDIT.

package process

import (
	context "context"

	auth "github.com/banzaicloud/pipeline/src/auth"

	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// GetProcess provides a mock function with given fields: ctx, org, id
func (_m *MockService) GetProcess(ctx context.Context, org auth.Organization, id string) (Process, error) {
	ret := _m.Called(ctx, org, id)

	var r0 Process
	if rf, ok := ret.Get(0).(func(context.Context, auth.Organization, string) Process); ok {
		r0 = rf(ctx, org, id)
	} else {
		r0 = ret.Get(0).(Process)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, auth.Organization, string) error); ok {
		r1 = rf(ctx, org, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListProcesses provides a mock function with given fields: ctx, org, query
func (_m *MockService) ListProcesses(ctx context.Context, org auth.Organization, query map[string]string) ([]Process, error) {
	ret := _m.Called(ctx, org, query)

	var r0 []Process
	if rf, ok := ret.Get(0).(func(context.Context, auth.Organization, map[string]string) []Process); ok {
		r0 = rf(ctx, org, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Process)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, auth.Organization, map[string]string) error); ok {
		r1 = rf(ctx, org, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Log provides a mock function with given fields: ctx, proc
func (_m *MockService) Log(ctx context.Context, proc Process) (Process, error) {
	ret := _m.Called(ctx, proc)

	var r0 Process
	if rf, ok := ret.Get(0).(func(context.Context, Process) Process); ok {
		r0 = rf(ctx, proc)
	} else {
		r0 = ret.Get(0).(Process)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, Process) error); ok {
		r1 = rf(ctx, proc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

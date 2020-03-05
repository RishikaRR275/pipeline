// +build !ignore_autogenerated

// Code generated by mga tool. DO NOT EDIT.

package kubernetes

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// MockDynamicFileClient is an autogenerated mock for the DynamicFileClient type.
type MockDynamicFileClient struct {
	mock.Mock
}

// Create provides a mock function.
func (_m *MockDynamicFileClient) Create(ctx context.Context, file []uint8) error {
	ret := _m.Called(ctx, file)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []uint8) error); ok {
		r0 = rf(ctx, file)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

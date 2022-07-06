package mocks

import (
	"context"

	"github.com/iulianpascalau/node-monitoring/data"
)

// AlarmHandlerStub -
type AlarmHandlerStub struct {
	ShouldQueryCalled func() bool
	QueryCalled       func(ctx context.Context) (data.AlarmResponse, error)
	QueryInfoCalled   func(ctx context.Context) (string, error)
	IdentifierCalled  func() string
}

// ShouldQuery -
func (stub *AlarmHandlerStub) ShouldQuery() bool {
	if stub.ShouldQueryCalled != nil {
		return stub.ShouldQueryCalled()
	}

	return false
}

// Query -
func (stub *AlarmHandlerStub) Query(ctx context.Context) (data.AlarmResponse, error) {
	if stub.QueryCalled != nil {
		return stub.QueryCalled(ctx)
	}

	return data.AlarmResponse{}, nil
}

// QueryInfo -
func (stub *AlarmHandlerStub) QueryInfo(ctx context.Context) (string, error) {
	if stub.QueryInfoCalled != nil {
		return stub.QueryInfoCalled(ctx)
	}

	return "", nil
}

// Identifier -
func (stub *AlarmHandlerStub) Identifier() string {
	if stub.IdentifierCalled != nil {
		return stub.IdentifierCalled()
	}

	return ""
}

// IsInterfaceNil -
func (stub *AlarmHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

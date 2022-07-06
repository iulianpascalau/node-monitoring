package mocks

import (
	"context"

	"github.com/iulianpascalau/node-monitoring/data"
)

// NotifierHandlerStub -
type NotifierHandlerStub struct {
	ProcessAlarmResponseCalled func(ctx context.Context, response data.AlarmResponse) error
}

// ProcessAlarmResponse -
func (stub *NotifierHandlerStub) ProcessAlarmResponse(ctx context.Context, response data.AlarmResponse) error {
	if stub.ProcessAlarmResponseCalled != nil {
		return stub.ProcessAlarmResponseCalled(ctx, response)
	}

	return nil
}

// IsInterfaceNil -
func (stub *NotifierHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

package poll

import (
	"context"

	"github.com/iulianpascalau/node-monitoring/data"
)

// AlarmHandler defines the operations that an alarm handler can perform
type AlarmHandler interface {
	ShouldQuery() bool
	Query(ctx context.Context) (data.AlarmResponse, error)
	Identifier() string
	IsInterfaceNil() bool
}

// NotifierHandler defines the operations implemented by a notifier
type NotifierHandler interface {
	ProcessAlarmResponse(ctx context.Context, response data.AlarmResponse) error
	IsInterfaceNil() bool
}

package notifiers

import "time"

type disabledTimeOfDayNotifier struct {
}

// NewDisabledTimeOfDayNotifier returns a new instance of type disabledTimeOfDayNotifier
func NewDisabledTimeOfDayNotifier() *disabledTimeOfDayNotifier {
	return &disabledTimeOfDayNotifier{}
}

// IsTimeOfDay always returns false
func (notifier *disabledTimeOfDayNotifier) IsTimeOfDay(_ time.Time) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (notifier *disabledTimeOfDayNotifier) IsInterfaceNil() bool {
	return notifier == nil
}

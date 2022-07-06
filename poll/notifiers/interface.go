package notifiers

import "time"

// TimeOfDayNotifier defines what does a time of day notifier should do
type TimeOfDayNotifier interface {
	IsTimeOfDay(t time.Time) bool
	IsInterfaceNil() bool
}

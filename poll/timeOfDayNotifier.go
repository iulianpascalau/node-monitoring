package poll

import (
	"fmt"
	"time"
)

type timeOfDayNotifier struct {
	setHour         int
	setMinute       int
	setSecond       int
	lastDayNotified int
}

func newTimeOfDayNotifier(setHour int, setMinute int, setSecond int) (*timeOfDayNotifier, error) {
	if setHour < 0 || setHour > 23 {
		return nil, fmt.Errorf("%w for setHour: interval 0-23, got %d", errInvalidValue, setHour)
	}
	if setMinute < 0 || setMinute > 59 {
		return nil, fmt.Errorf("%w for setMinute: interval 0-59, got %d", errInvalidValue, setMinute)
	}
	if setSecond < 0 || setSecond > 59 {
		return nil, fmt.Errorf("%w for setSecond: interval 0-59, got %d", errInvalidValue, setSecond)
	}

	return &timeOfDayNotifier{
		setHour:         setHour,
		setMinute:       setMinute,
		setSecond:       setSecond,
		lastDayNotified: 0, // trigger the info in the same day
	}, nil
}

func (notifier *timeOfDayNotifier) isTimeOfDay(t time.Time) bool {
	if notifier.lastDayNotified == t.Day() {
		return false
	}

	setTime := time.Date(t.Year(), t.Month(), t.Day(), notifier.setHour, notifier.setMinute, notifier.setSecond, 0, t.Location())
	if setTime.Unix() <= t.Unix() {
		notifier.lastDayNotified = t.Day()
		return true
	}

	return false
}

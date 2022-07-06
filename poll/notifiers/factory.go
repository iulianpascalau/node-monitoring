package notifiers

// CreateTimeOfDayNotifier will create a new instance of type TimeOfDayNotifier
func CreateTimeOfDayNotifier(
	active bool,
	setHour int,
	setMinute int,
	setSecond int,
) (TimeOfDayNotifier, error) {
	if active {
		return NewTimeOfDayNotifier(setHour, setMinute, setSecond)
	}

	return NewDisabledTimeOfDayNotifier(), nil
}

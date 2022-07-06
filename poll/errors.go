package poll

import "errors"

var errNoAlarmsSet = errors.New("no alarms set")
var errNoActiveNotifiers = errors.New("no active notifiers")
var errNilAlarmHandler = errors.New("nil alarm handler")
var errNilNotifier = errors.New("nil notifier")

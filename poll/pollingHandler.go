package poll

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/iulianpascalau/node-monitoring/data"
	"github.com/iulianpascalau/node-monitoring/poll/notifiers"
)

const pollingInterval = time.Millisecond * 100
const systemIdentifier = "system"
const systemMessage = `System is running. Uptime: %v. 
Number of processing error: %d, number of alarms with error: %d`

var log = logger.GetOrCreate("poll")

// ArgsPollingHandler represents the arguments DTO for the pollingHandler constructor
type ArgsPollingHandler struct {
	Alarms         []AlarmHandler
	Notifiers      []NotifierHandler
	SendInfo       bool
	SendInfoHour   int
	SendInfoMinute int
	SendInfoSecond int
}

type pollingHandler struct {
	notifiers.TimeOfDayNotifier
	*pollingHandlerState
	alarms    []AlarmHandler
	notifiers []NotifierHandler
	startTime time.Time
	cancel    func()
}

// NewPollingHandler creates a new polling handler instance
func NewPollingHandler(args ArgsPollingHandler) (*pollingHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	timeOfDay, err := notifiers.CreateTimeOfDayNotifier(args.SendInfo, args.SendInfoHour, args.SendInfoMinute, args.SendInfoSecond)
	if err != nil {
		return nil, err
	}

	ph := &pollingHandler{
		TimeOfDayNotifier:   timeOfDay,
		pollingHandlerState: &pollingHandlerState{},
		alarms:              args.Alarms,
		notifiers:           args.Notifiers,
		startTime:           time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	ph.cancel = cancel

	go ph.processLoop(ctx)

	return ph, nil
}

func checkArgs(args ArgsPollingHandler) error {
	if len(args.Alarms) == 0 {
		return errNoAlarmsSet
	}
	for idx, alarm := range args.Alarms {
		if check.IfNil(alarm) {
			return fmt.Errorf("%w at index %d", errNilAlarmHandler, idx)
		}
	}

	if len(args.Notifiers) == 0 {
		return errNoActiveNotifiers
	}
	for idx, notifier := range args.Notifiers {
		if check.IfNil(notifier) {
			return fmt.Errorf("%w at index %d", errNilNotifier, idx)
		}
	}

	return nil
}

func (ph *pollingHandler) processLoop(ctx context.Context) {
	ph.setIsRunning()
	log.Debug("polling handler's process loop has started...")
	defer func() {
		log.Debug("polling handler's process loop has been stopped")
		ph.setIsStopped()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollingInterval):
			ph.poll(ctx)
		}
	}
}

func (ph *pollingHandler) poll(ctx context.Context) {
	if ph.IsTimeOfDay(time.Now()) {
		response := ph.createInfoMessage(ctx)
		ph.notifyAll(ctx, response)
	}

	for _, alarm := range ph.alarms {
		if !alarm.ShouldQuery() {
			continue
		}

		response, err := alarm.Query(ctx)
		if err != nil {
			log.Error("error querying alarm", "identifier", alarm.Identifier(), "error", err.Error())
			ph.incrementErrors()
			continue
		}

		ph.notifyAll(ctx, response)
	}
}

func (ph *pollingHandler) notifyAll(ctx context.Context, response data.AlarmResponse) {
	if response.Level == data.Error {
		ph.incrementAlarmsWithError()
	}

	for _, notifier := range ph.notifiers {
		err := notifier.ProcessAlarmResponse(ctx, response)
		if err != nil {
			log.Error("error pushing notification", "error", err.Error())
			ph.incrementErrors()
			continue
		}
	}
}

func (ph *pollingHandler) createInfoMessage(ctx context.Context) data.AlarmResponse {
	response := data.AlarmResponse{
		Identifier: systemIdentifier,
		Level:      data.Info,
		Data: fmt.Sprintf(systemMessage,
			time.Since(time.Now()),
			ph.getNumErrors(),
			ph.getNumAlarmsWithError()),
	}

	if ph.getNumErrors()+ph.getNumAlarmsWithError() > 0 {
		response.Level = data.Error
	}

	for _, alarm := range ph.alarms {
		status, err := alarm.QueryInfo(ctx)
		if err == nil {
			response.Data += fmt.Sprintf("\nStatus for alarm %s: %s", alarm.Identifier(), status)
		} else {
			response.Data += fmt.Sprintf("\nError fetching status info for alarm %s: %s", alarm.Identifier(), err.Error())
			response.Level = data.Error
		}
	}

	ph.resetNumErrors()

	log.Debug("polling handler creating info message",
		"time", time.Now(), "identifier", response.Identifier, "level", response.Level, "message", "\r\n"+response.Data)

	return response
}

// Close will close the running processLoop go routine
func (ph *pollingHandler) Close() error {
	ph.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *pollingHandler) IsInterfaceNil() bool {
	return ph == nil
}

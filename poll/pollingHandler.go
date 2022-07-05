package poll

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/iulianpascalau/node-monitoring/data"
)

const pollingInterval = time.Millisecond * 100
const systemIdentifier = "system"
const systemMessage = `System is running. Uptime: %v. 
Number of processing error: %d, number of alarms with error: %d`

var log = logger.GetOrCreate("poll")

type pollingHandler struct {
	*timeOfDayNotifier
	alarms             []AlarmHandler
	notifiers          []NotifierHandler
	numErrors          uint64
	numAlarmsWithError uint64
	startTime          time.Time
}

func NewPollingHandler() (*pollingHandler, error) {

	return &pollingHandler{
		timeOfDayNotifier:  nil,
		alarms:             nil,
		notifiers:          nil,
		numErrors:          0,
		numAlarmsWithError: 0,
		startTime:          time.Now(),
	}, nil
}

func (ph *pollingHandler) processLoop(ctx context.Context) {
	log.Debug("polling handler's process loop has started...")
	defer log.Debug("polling handler's process loop has been stopped")

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
	if ph.isTimeOfDay(time.Now()) {
		response := ph.createInfoMessage()
		ph.notifyAll(ctx, response)
	}

	for _, alarm := range ph.alarms {
		if !alarm.ShouldQuery() {
			continue
		}

		response, err := alarm.Query(ctx)
		if err != nil {
			log.Error("error querying alarm", "identifier", alarm.Identifier(), "error", err.Error())
			atomic.AddUint64(&ph.numErrors, 1)
			continue
		}

		ph.notifyAll(ctx, response)
	}
}

func (ph *pollingHandler) notifyAll(ctx context.Context, response data.AlarmResponse) {
	if response.Level == data.Error {
		atomic.AddUint64(&ph.numAlarmsWithError, 1)
	}

	for _, notifier := range ph.notifiers {
		err := notifier.ProcessAlarmResponse(ctx, response)
		if err != nil {
			log.Error("error pushing notification", "error", err.Error())
			atomic.AddUint64(&ph.numErrors, 1)
			continue
		}
	}
}

func (ph *pollingHandler) createInfoMessage() data.AlarmResponse {
	response := data.AlarmResponse{
		Identifier: systemIdentifier,
		Level:      data.Info,
		Data: fmt.Sprintf(systemMessage,
			time.Since(time.Now()),
			atomic.LoadUint64(&ph.numErrors),
			atomic.LoadUint64(&ph.numAlarmsWithError)),
	}

	atomic.StoreUint64(&ph.numErrors, 0)
	atomic.StoreUint64(&ph.numAlarmsWithError, 0)

	log.Debug("polling handler creating info message",
		"time", time.Now(), "identifier", response.Identifier, "level", response.Level, "message", response.Data)

	return response
}

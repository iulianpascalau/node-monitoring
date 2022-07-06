package poll

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/iulianpascalau/node-monitoring/data"
	"github.com/iulianpascalau/node-monitoring/mocks"
	"github.com/iulianpascalau/node-monitoring/poll/notifiers"
	"github.com/stretchr/testify/assert"
)

func createMockArgsPollingHandler() ArgsPollingHandler {
	return ArgsPollingHandler{
		Alarms:         []AlarmHandler{&mocks.AlarmHandlerStub{}},
		Notifiers:      []NotifierHandler{&mocks.NotifierHandlerStub{}},
		SendInfo:       false,
		SendInfoHour:   0,
		SendInfoMinute: 0,
		SendInfoSecond: 0,
	}
}

func TestNewPollingHandler(t *testing.T) {
	t.Parallel()

	t.Run("empty alarms should error", func(t *testing.T) {
		args := createMockArgsPollingHandler()
		args.Alarms = nil

		pollHandler, err := NewPollingHandler(args)
		assert.True(t, check.IfNil(pollHandler))
		assert.Equal(t, errNoAlarmsSet, err)
	})
	t.Run("nil alarm should error", func(t *testing.T) {
		args := createMockArgsPollingHandler()
		args.Alarms = append(args.Alarms, nil)

		pollHandler, err := NewPollingHandler(args)
		assert.True(t, check.IfNil(pollHandler))
		assert.True(t, errors.Is(err, errNilAlarmHandler))
		assert.True(t, strings.Contains(err.Error(), "at index 1"))
	})
	t.Run("empty notifiers should error", func(t *testing.T) {
		args := createMockArgsPollingHandler()
		args.Notifiers = nil

		pollHandler, err := NewPollingHandler(args)
		assert.True(t, check.IfNil(pollHandler))
		assert.Equal(t, errNoActiveNotifiers, err)
	})
	t.Run("nil notifier should error", func(t *testing.T) {
		args := createMockArgsPollingHandler()
		args.Notifiers = append(args.Notifiers, nil)

		pollHandler, err := NewPollingHandler(args)
		assert.True(t, check.IfNil(pollHandler))
		assert.True(t, errors.Is(err, errNilNotifier))
		assert.True(t, strings.Contains(err.Error(), "at index 1"))
	})
	t.Run("invalid time of day should error", func(t *testing.T) {
		args := createMockArgsPollingHandler()
		args.SendInfoMinute = 60
		args.SendInfo = true

		pollHandler, err := NewPollingHandler(args)
		assert.True(t, check.IfNil(pollHandler))
		assert.True(t, errors.Is(err, notifiers.ErrInvalidValue))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsPollingHandler()

		pollHandler, err := NewPollingHandler(args)
		assert.False(t, check.IfNil(pollHandler))
		assert.Nil(t, err)

		_ = pollHandler.Close()
	})
}

func TestPollingHandler_Close(t *testing.T) {
	t.Parallel()

	args := createMockArgsPollingHandler()
	pollHandler, _ := NewPollingHandler(args)

	time.Sleep(time.Second) // wait for go routine to start
	assert.True(t, pollHandler.IsRunning())

	t.Run("first close should work", func(t *testing.T) {
		err := pollHandler.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second) // wait for go routine to finish
		assert.False(t, pollHandler.IsRunning())

	})
	t.Run("double close should be ok", func(t *testing.T) {
		err := pollHandler.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second)
		assert.False(t, pollHandler.IsRunning())
	})
}

func TestPollingHandler_AlarmShouldNotQueryIfNotRequired(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	wg := sync.WaitGroup{}
	wg.Add(3)
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return false
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")

				return data.AlarmResponse{}, nil
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				assert.Fail(t, "should have not called process alarm response")

				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()

	_ = pollHandler.Close()
}

func TestPollingHandler_AlarmFailsOnQueryShouldNotNotify(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	expectedErr := errors.New("expected error")
	wg := sync.WaitGroup{}
	wg.Add(3)
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				return data.AlarmResponse{}, expectedErr
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				assert.Fail(t, "should have not called process alarm response")

				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()
	_ = pollHandler.Close()

	assert.Equal(t, 2, pollHandler.getNumErrors())
}

func TestPollingHandler_AlarmReturnsResultShouldNotifyInfoLevel(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	wg := sync.WaitGroup{}
	wg.Add(3)
	numNotified := uint64(0)
	alarmResponse := data.AlarmResponse{
		Identifier: "test",
		Level:      data.Info,
		Data:       "test message",
	}
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				return alarmResponse, nil
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()
	_ = pollHandler.Close()

	assert.Equal(t, 0, pollHandler.getNumErrors())
	assert.Equal(t, 0, pollHandler.getNumAlarmsWithError())
	assert.Equal(t, uint64(2), atomic.LoadUint64(&numNotified))
}

func TestPollingHandler_AlarmReturnsResultShouldNotifyInfoLevel2Alarms3Notifiers(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	wg := sync.WaitGroup{}
	wg.Add(3)
	numNotified := uint64(0)
	alarmResponse := data.AlarmResponse{
		Identifier: "test",
		Level:      data.Info,
		Data:       "test message",
	}
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				fmt.Println("first alarm query")
				return alarmResponse, nil
			},
		},
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				fmt.Println("second alarm query")
				return alarmResponse, nil
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				fmt.Println("first notifier")
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return nil
			},
		},
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				fmt.Println("second notifier")
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return nil
			},
		},
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				fmt.Println("third notifier")
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()
	_ = pollHandler.Close()

	assert.Equal(t, 0, pollHandler.getNumErrors())
	assert.Equal(t, 0, pollHandler.getNumAlarmsWithError())
	assert.Equal(t, uint64(15), atomic.LoadUint64(&numNotified)) // 3 notifiers * 2 alarms * 2 times + 3 notifiers from first alarm 3-rd time
}

func TestPollingHandler_AlarmReturnsResultShouldNotifyErrorLevel(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	wg := sync.WaitGroup{}
	wg.Add(3)
	numNotified := uint64(0)
	alarmResponse := data.AlarmResponse{
		Identifier: "test",
		Level:      data.Error,
		Data:       "test message",
	}
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				return alarmResponse, nil
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()
	_ = pollHandler.Close()

	assert.Equal(t, 0, pollHandler.getNumErrors())
	assert.Equal(t, 2, pollHandler.getNumAlarmsWithError())
	assert.Equal(t, uint64(2), atomic.LoadUint64(&numNotified))
}

func TestPollingHandler_AlarmReturnsResultNotifierErrors(t *testing.T) {
	t.Parallel()
	args := createMockArgsPollingHandler()

	wg := sync.WaitGroup{}
	wg.Add(3)
	numNotified := uint64(0)
	expectedErr := errors.New("expected error")
	alarmResponse := data.AlarmResponse{
		Identifier: "test",
		Level:      data.Info,
		Data:       "test message",
	}
	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				wg.Done()
				time.Sleep(time.Second)
				return true
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				assert.Fail(t, "should have not called QueryInfoCalled")

				return "", nil
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				return alarmResponse, nil
			},
		},
	}
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				assert.Equal(t, alarmResponse, response)
				atomic.AddUint64(&numNotified, 1)
				return expectedErr
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()
	_ = pollHandler.Close()

	assert.Equal(t, 2, pollHandler.getNumErrors())
	assert.Equal(t, 0, pollHandler.getNumAlarmsWithError())
	assert.Equal(t, uint64(2), atomic.LoadUint64(&numNotified))
}

func TestPollingHandler_CreateInfoMessageWithErrors(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	args := createMockArgsPollingHandler()
	args.SendInfo = true
	wg := sync.WaitGroup{}
	wg.Add(1)

	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "query string 1", nil
			},
			IdentifierCalled: func() string {
				return "1"
			},
		},
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "query string 2", nil
			},
			IdentifierCalled: func() string {
				return "2"
			},
		},
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "", expectedErr
			},
			IdentifierCalled: func() string {
				return "3"
			},
		},
	}
	var receivedResponse data.AlarmResponse
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				receivedResponse = response
				wg.Done()

				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	pollHandler.incrementErrors()
	pollHandler.incrementAlarmsWithError()
	pollHandler.incrementAlarmsWithError()

	wg.Wait()

	expectedPartialString := `Number of processing error: 1, number of alarms with error: 2
Status for alarm 1: query string 1
Status for alarm 2: query string 2
Error fetching status info for alarm 3: expected error`

	assert.Equal(t, data.Error, receivedResponse.Level)
	assert.Equal(t, systemIdentifier, receivedResponse.Identifier)
	fmt.Println(receivedResponse.Data)
	assert.True(t, strings.Contains(receivedResponse.Data, expectedPartialString))

	_ = pollHandler.Close()
}

func TestPollingHandler_CreateInfoMessageNoErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgsPollingHandler()
	args.SendInfo = true
	wg := sync.WaitGroup{}
	wg.Add(1)

	args.Alarms = []AlarmHandler{
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "query string 1", nil
			},
			IdentifierCalled: func() string {
				return "1"
			},
		},
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "query string 2", nil
			},
			IdentifierCalled: func() string {
				return "2"
			},
		},
		&mocks.AlarmHandlerStub{
			ShouldQueryCalled: func() bool {
				return false
			},
			QueryCalled: func(ctx context.Context) (data.AlarmResponse, error) {
				assert.Fail(t, "should have not called Query")
				return data.AlarmResponse{}, nil
			},
			QueryInfoCalled: func(ctx context.Context) (string, error) {
				return "query string 3", nil
			},
			IdentifierCalled: func() string {
				return "3"
			},
		},
	}
	var receivedResponse data.AlarmResponse
	args.Notifiers = []NotifierHandler{
		&mocks.NotifierHandlerStub{
			ProcessAlarmResponseCalled: func(ctx context.Context, response data.AlarmResponse) error {
				receivedResponse = response
				wg.Done()

				return nil
			},
		},
	}

	pollHandler, _ := NewPollingHandler(args)

	wg.Wait()

	expectedPartialString := `Number of processing error: 0, number of alarms with error: 0
Status for alarm 1: query string 1
Status for alarm 2: query string 2
Status for alarm 3: query string 3`

	assert.Equal(t, data.Info, receivedResponse.Level)
	assert.Equal(t, systemIdentifier, receivedResponse.Identifier)
	fmt.Println(receivedResponse.Data)
	assert.True(t, strings.Contains(receivedResponse.Data, expectedPartialString))

	_ = pollHandler.Close()
}

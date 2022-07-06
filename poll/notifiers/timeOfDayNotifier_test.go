package notifiers

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeOfDayNotifier(t *testing.T) {
	t.Parallel()

	t.Run("invalid setHour", func(t *testing.T) {
		notifier, err := NewTimeOfDayNotifier(-1, 0, 0)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setHour: interval 0-23, got -1"))

		notifier, err = NewTimeOfDayNotifier(24, 0, 0)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setHour: interval 0-23, got 24"))
	})
	t.Run("invalid setMinute", func(t *testing.T) {
		notifier, err := NewTimeOfDayNotifier(0, -1, 0)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setMinute: interval 0-59, got -1"))

		notifier, err = NewTimeOfDayNotifier(0, 60, 0)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setMinute: interval 0-59, got 60"))
	})
	t.Run("invalid setSecond", func(t *testing.T) {
		notifier, err := NewTimeOfDayNotifier(0, 0, -1)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setSecond: interval 0-59, got -1"))

		notifier, err = NewTimeOfDayNotifier(0, 0, 60)
		assert.True(t, check.IfNil(notifier))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for setSecond: interval 0-59, got 60"))
	})
	t.Run("should work", func(t *testing.T) {
		notifier, err := NewTimeOfDayNotifier(0, 0, 0)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(notifier))
		assert.Equal(t, 0, notifier.setHour)
		assert.Equal(t, 0, notifier.setMinute)
		assert.Equal(t, 0, notifier.setSecond)

		notifier, err = NewTimeOfDayNotifier(23, 59, 59)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(notifier))
		assert.Equal(t, 23, notifier.setHour)
		assert.Equal(t, 59, notifier.setMinute)
		assert.Equal(t, 59, notifier.setSecond)

		notifier, err = NewTimeOfDayNotifier(23, 59, 58)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(notifier))
		assert.Equal(t, 23, notifier.setHour)
		assert.Equal(t, 59, notifier.setMinute)
		assert.Equal(t, 58, notifier.setSecond)
	})
}

func TestTimeOfDayNotifier_IsTimeOfDay(t *testing.T) {
	t.Parallel()

	t.Run("new instance should test the time set", func(t *testing.T) {
		notifier, _ := NewTimeOfDayNotifier(12, 0, 0)
		currentTime := time.Date(2022, 07, 06, 11, 59, 59, 0, time.UTC)
		assert.False(t, notifier.IsTimeOfDay(currentTime))

		currentTime = time.Date(2022, 07, 06, 0, 0, 0, 0, time.UTC)
		assert.False(t, notifier.IsTimeOfDay(currentTime))

		currentTime = time.Date(2022, 07, 06, 12, 0, 0, 0, time.UTC)
		assert.True(t, notifier.IsTimeOfDay(currentTime))
	})
	t.Run("should work", func(t *testing.T) {
		notifier, _ := NewTimeOfDayNotifier(12, 0, 0)

		t.Run("first time returns true", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 06, 12, 0, 0, 0, time.UTC)
			assert.True(t, notifier.IsTimeOfDay(currentTime))
		})
		t.Run("second time returns false", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 06, 12, 0, 0, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))
		})
		t.Run("same day returns false", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 06, 23, 59, 59, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))
		})
		t.Run("next day returns false until time", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 07, 0, 0, 0, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))

			currentTime = time.Date(2022, 07, 07, 10, 0, 0, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))

			currentTime = time.Date(2022, 07, 07, 11, 59, 59, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))
		})
		t.Run("next day returns true once", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 07, 12, 0, 0, 0, time.UTC)
			assert.True(t, notifier.IsTimeOfDay(currentTime))

			currentTime = time.Date(2022, 07, 07, 12, 0, 0, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))
		})
		t.Run("next day returns false", func(t *testing.T) {
			currentTime := time.Date(2022, 07, 07, 18, 0, 0, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))

			currentTime = time.Date(2022, 07, 07, 23, 59, 59, 0, time.UTC)
			assert.False(t, notifier.IsTimeOfDay(currentTime))
		})
	})
}

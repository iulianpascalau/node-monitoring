package notifiers

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTimeOfDayNotifier(t *testing.T) {
	t.Parallel()

	notifier := NewDisabledTimeOfDayNotifier()
	assert.False(t, check.IfNil(notifier))
}

func TestDisabledTimeOfDayNotifier_IsTimeOfDay(t *testing.T) {
	t.Parallel()

	notifier := NewDisabledTimeOfDayNotifier()
	assert.False(t, notifier.IsTimeOfDay(time.Now()))
}

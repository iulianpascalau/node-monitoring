package notifiers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestCreateTimeOfDayNotifier(t *testing.T) {
	t.Parallel()

	t.Run("inactive should return the disabled instance", func(t *testing.T) {
		instance, err := CreateTimeOfDayNotifier(false, 0, 0, 0)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(instance))
		assert.Equal(t, "*notifiers.disabledTimeOfDayNotifier", fmt.Sprintf("%T", instance))
	})
	t.Run("active should return the real instance", func(t *testing.T) {
		instance, err := CreateTimeOfDayNotifier(true, 0, 0, 0)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(instance))
		assert.Equal(t, "*notifiers.timeOfDayNotifier", fmt.Sprintf("%T", instance))
	})
	t.Run("active but with wrong parameters should error", func(t *testing.T) {
		instance, err := CreateTimeOfDayNotifier(true, 24, 0, 0)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, check.IfNil(instance))
	})
}

package tool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsTimeToUpdate(t *testing.T) {
	now := time.Now()

	justNow := now.Add(-5 * time.Second)
	yesterday := now.Add(-24 * time.Hour)
	twoWeeksAgo := now.Add(-15 * 24 * time.Hour)
	coupleMonthsBack := now.Add(-60 * 24 * time.Hour)

	assert := assert.New(t)

	assert.True(UpdateAlways.IsTimeToUpdate(justNow))
	assert.True(UpdateAlways.IsTimeToUpdate(yesterday))
	assert.True(UpdateAlways.IsTimeToUpdate(twoWeeksAgo))
	assert.True(UpdateAlways.IsTimeToUpdate(coupleMonthsBack))

	assert.False(UpdateDaily.IsTimeToUpdate(justNow))
	assert.True(UpdateDaily.IsTimeToUpdate(yesterday))
	assert.True(UpdateDaily.IsTimeToUpdate(twoWeeksAgo))
	assert.True(UpdateDaily.IsTimeToUpdate(coupleMonthsBack))

	assert.False(UpdateWeekly.IsTimeToUpdate(justNow))
	assert.False(UpdateWeekly.IsTimeToUpdate(yesterday))
	assert.True(UpdateWeekly.IsTimeToUpdate(twoWeeksAgo))
	assert.True(UpdateWeekly.IsTimeToUpdate(coupleMonthsBack))

	assert.False(UpdateMonthly.IsTimeToUpdate(justNow))
	assert.False(UpdateMonthly.IsTimeToUpdate(yesterday))
	assert.False(UpdateMonthly.IsTimeToUpdate(twoWeeksAgo))
	assert.True(UpdateMonthly.IsTimeToUpdate(coupleMonthsBack))

	assert.False(UpdateNever.IsTimeToUpdate(justNow))
	assert.False(UpdateNever.IsTimeToUpdate(yesterday))
	assert.False(UpdateNever.IsTimeToUpdate(twoWeeksAgo))
	assert.False(UpdateNever.IsTimeToUpdate(coupleMonthsBack))
}

func TestIsTimeToUpdateDoesNotHandleDefaults(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	UpdateDefault.IsTimeToUpdate(time.Now())
}

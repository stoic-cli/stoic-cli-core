package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStoic(t *testing.T) {
	assert := assert.New(t)

	assert.True(isStoic("stoic"))
	assert.True(isStoic("stoic.bin"))
	assert.True(isStoic("stoic.exe"))
	assert.True(isStoic("stoic1"))
	assert.True(isStoic("stoic2"))
	assert.True(isStoic("stoic3"))
	assert.True(isStoic("stoic-1.0.0-darwin-amd64"))
	assert.True(isStoic("stoic_1.0.0_darwin_amd64"))

	assert.False(isStoic("stoical"))
	assert.False(isStoic("stoicism"))
	assert.False(isStoic("stoichastic"))

	assert.False(isStoic("unrelated"))
}

package engine

import (
	"testing"
	"time"

	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stoic-cli/stoic-cli-core/util"
)

func TestEngineState(t *testing.T) {
	tid := util.SetupTestInDir(t)
	defer tid.Close()

	stoic, err := NewWithOptions(EngineOptions{
		Root: tid.TestDir(),
	})
	if err != nil {
		t.Fatalf("unable to setup stoic instance: %v", err)
	}

	engine, ok := stoic.(*engine)
	if !ok {
		t.Fatalf("unable to cast Stoic instance to engine")
	}

	t.Run("EmptyState", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)

		state := engine.LoadState(testName)
		assert.Equal(testName, state.ToolId())
		assert.NotEqual("", state.Filename())

		assert.Equal(tool.NullVersion, state.UpstreamVersion(tool.DefaultChannel))
		assert.True(state.LastUpstreamUpdate(tool.DefaultChannel).IsZero())

		tc := tool.Channel("testing")
		assert.Equal(tool.NullVersion, state.UpstreamVersion(tc))
		assert.True(state.LastUpstreamUpdate(tc).IsZero())
	})
	t.Run("DefaultChannelFlow", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)

		state := engine.LoadState(testName)
		assert.Equal(testName, state.ToolId())
		assert.NotEqual("", state.Filename())

		baseTimestamp := time.Now().Add(-time.Second)

		ts, ok := state.(*toolState)
		assert.True(ok)

		tc := tool.DefaultChannel
		ts.setUpstreamVersion(tc, tool.Version("1"))
		assert.Equal(tool.Version("1"), state.UpstreamVersion(tc))
		assert.Truef(
			baseTimestamp.Before(state.LastUpstreamUpdate(tc)),
			"%v is NOT before %v", baseTimestamp, state.LastUpstreamUpdate(tc))
		assert.Nil(state.CurrentCheckout())

		ts.addCheckout(tool.Version("1"), "checkout-v1-xyz", false)
		assert.Nil(state.CurrentCheckout())

		ts.setCurrentCheckout("checkout-v1-xyz")
		checkout := state.CurrentCheckout()

		assert.NotNil(checkout)
		assert.Equal("checkout-v1-xyz", checkout.Path())
		assert.Equal(tool.Version("1"), checkout.Version())
	})
	t.Run("AlternateChannelFlow", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)

		state := engine.LoadState(testName)
		assert.Equal(testName, state.ToolId())
		assert.NotEqual("", state.Filename())

		baseTimestamp := time.Now().Add(-time.Second)

		ts, ok := state.(*toolState)
		assert.True(ok)

		tc := tool.Channel("unstable")
		ts.setUpstreamVersion(tc, tool.Version("1"))
		assert.Equal(tool.Version("1"), state.UpstreamVersion(tc))
		assert.Truef(
			baseTimestamp.Before(state.LastUpstreamUpdate(tc)),
			"%v is NOT before %v", baseTimestamp, state.LastUpstreamUpdate(tc))
		assert.Nil(state.CurrentCheckout())

		ts.addCheckout(tool.Version("1"), "checkout-v1-xyz", false)
		assert.Nil(state.CurrentCheckout())

		ts.setCurrentCheckout("checkout-v1-xyz")
		checkout := state.CurrentCheckout()

		assert.NotNil(checkout)
		assert.Equal("checkout-v1-xyz", checkout.Path())
		assert.Equal(tool.Version("1"), checkout.Version())
	})
}

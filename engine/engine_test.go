package engine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stoic-cli/stoic-cli-core/util"
)

func TestNewWithOptions(t *testing.T) {
	tid := util.SetupTestInDir(t)
	defer tid.Close()

	t.Run("DefaultConstructor", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)
		defer util.SwitchHome(t, testName).Close()

		stoic, err := New()
		assert.Nil(err)

		engine, ok := stoic.(*engine)
		assert.True(ok)

		assert.Equal(filepath.Join(tid.TestDir(), testName, ".stoic"), stoic.Root())
		assert.Equal("", engine.ConfigFile())
		assert.Equal(tool.UpdateDefault, engine.updateFrequencyOverride)
		assert.Equal(tool.UpdateWeekly, engine.updateFrequencyFallback)
	})
	t.Run("OverrideRoot", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)

		stoic, err := NewWithOptions(EngineOptions{Root: testName})
		assert.Nil(err)

		engine, ok := stoic.(*engine)
		assert.True(ok)

		assert.Equal(filepath.Join(tid.TestDir(), testName), engine.Root())
		assert.Equal("", engine.ConfigFile())
		assert.Equal(tool.UpdateDefault, engine.updateFrequencyOverride)
		assert.Equal(tool.UpdateWeekly, engine.updateFrequencyFallback)
	})
	t.Run("CustomConfig", func(t *testing.T) {
		assert, testName := tid.SetupTest(t)

		err := os.MkdirAll(testName, 0755)
		assert.Nil(err)

		err = ioutil.WriteFile(filepath.Join(testName, "config"), []byte("update: never\n"), 0644)
		assert.Nil(err)

		stoic, err := NewWithOptions(EngineOptions{Root: testName})
		assert.Nil(err)

		engine, ok := stoic.(*engine)
		assert.True(ok)

		assert.Equal(filepath.Join(tid.TestDir(), testName), engine.Root())
		assert.Equal(filepath.Join(engine.Root(), "config"), engine.ConfigFile())
		assert.Equal(tool.UpdateDefault, engine.updateFrequencyOverride)
		assert.Equal(tool.UpdateNever, engine.updateFrequencyFallback)
	})
}

package engine

import (
	"fmt"
	"os"
	"testing"

	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stoic-cli/stoic-cli-core/util"
	"github.com/stretchr/testify/assert"
)

type mockToolGetter struct {
	FetchLatestFunc  func() (tool.Version, error)
	FetchVersionFunc func(tool.Version) error

	CheckoutToFunc func(tool.Version, string) error

	Options map[string]interface{}
}

var (
	mtg_FetchLatest_callCount  uint
	mtg_FetchVersion_callCount uint
	mtg_CheckoutTo_callCount   uint
)

func (m *mockToolGetter) FetchLatest() (tool.Version, error) {
	mtg_FetchLatest_callCount += 1
	if m.FetchLatestFunc == nil {
		if latest, ok := m.Options["latest"].(string); ok {
			return tool.Version(latest), nil
		}
		return tool.NullVersion, nil
	}
	return m.FetchLatestFunc()
}

func (m *mockToolGetter) FetchVersion(version tool.Version) error {
	mtg_FetchVersion_callCount += 1
	if m.FetchVersionFunc == nil {
		return nil
	}
	return m.FetchVersionFunc(version)
}

func (m *mockToolGetter) CheckoutTo(version tool.Version, path string) error {
	mtg_CheckoutTo_callCount += 1
	if m.CheckoutToFunc == nil {
		return nil
	}
	return m.CheckoutToFunc(version, path)
}

type mockToolRunner struct {
	SetupFunc func(tool.Checkout) error
	RunFunc   func(tool.Checkout, string, []string) error

	Options map[string]interface{}
}

var (
	mtr_Setup_callCount uint
	mtr_Run_callCount   uint
)

func (m *mockToolRunner) Setup(checkout tool.Checkout) error {
	mtr_Setup_callCount += 1
	if m.SetupFunc == nil {
		return nil
	}
	return m.SetupFunc(checkout)
}

func (m *mockToolRunner) Run(checkout tool.Checkout, name string, args []string) error {
	mtr_Run_callCount += 1
	if m.RunFunc == nil {
		return nil
	}
	return m.RunFunc(checkout, name, args)
}

func TestRunTool(t *testing.T) {
	tid := util.SetupTestInDir(t)
	defer tid.Close()

	RegisterGetter(t.Name(), func(s stoic.Stoic, toolConfig tool.ConfigFormat) (tool.Getter, error) {
		return &mockToolGetter{Options: toolConfig.Getter.Options}, nil
	})
	RegisterRunner(t.Name(), func(s stoic.Stoic, toolConfig tool.ConfigFormat) (tool.Runner, error) {
		return &mockToolRunner{Options: toolConfig.Runner.Options}, nil
	})

	configFile, err := os.Create("config")
	if err != nil {
		t.Fatalf("unable to create config file: %v", err)
	}

	fmt.Fprintf(configFile, `
tools:
  test1:
    endpoint: github.com/stoic-cli/stoic-cli-core
    getter:
      type: '%[1]v'
      latest: v1.0.0
    runner:
      type: '%[1]v'
`, t.Name())
	configFile.Close()

	stoic, err := NewWithOptions(EngineOptions{
		Root: tid.TestDir(),
	})
	if err != nil {
		t.Fatalf("unable to set up stoic instance: %v", err)
	}

	err = stoic.RunTool("test1", []string{"a1", "a2", "a3"})
	assert.Nil(t, err)

	assert.Equal(t, uint(1), mtg_FetchLatest_callCount)
	assert.Equal(t, uint(0), mtg_FetchVersion_callCount)
	assert.Equal(t, uint(1), mtg_CheckoutTo_callCount)

	assert.Equal(t, uint(1), mtr_Setup_callCount)
	assert.Equal(t, uint(1), mtr_Run_callCount)
}

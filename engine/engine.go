package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/format"
	"github.com/stoic-cli/stoic-cli-core/tool"
	"gopkg.in/yaml.v2"
)

type engine struct {
	root         string
	configFile   string
	stateDir     string
	checkoutsDir string

	updateFrequencyFallback tool.UpdateFrequency
	updateFrequencyOverride tool.UpdateFrequency

	tools map[string]format.ToolConfig
}

type EngineOptions struct {
	Root            string
	UpdateFrequency tool.UpdateFrequency
}

func New() (stoic.Stoic, error) {
	return NewWithOptions(EngineOptions{})
}

func NewWithOptions(o EngineOptions) (stoic.Stoic, error) {
	if o.Root == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		o.Root = filepath.Join(home, ".stoic")
	} else {
		root, err := filepath.Abs(o.Root)
		if err != nil {
			return nil, err
		}
		o.Root = root
	}

	var tools map[string]format.ToolConfig

	updateFrequencyFallback := DefaultToolUpdateFrequency

	configFilename := filepath.Join(o.Root, "config")
	if configFile, err := os.Open(configFilename); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"unable to load config from %v: %v", configFilename, err)
		}
		configFilename = ""
	} else {
		var sc format.StoicConfig

		decoder := yaml.NewDecoder(configFile)
		if err := decoder.Decode(&sc); err != nil {
			return nil, fmt.Errorf(
				"unable to load config from %v: %v", configFilename, err)
		}

		if sc.UpdateFrequency != tool.UpdateDefault {
			updateFrequencyFallback = sc.UpdateFrequency
		}
		tools = sc.Tools
	}

	return &engine{
		root:         o.Root,
		configFile:   configFilename,
		stateDir:     filepath.Join(o.Root, ".state"),
		checkoutsDir: filepath.Join(o.Root, "checkout"),

		updateFrequencyFallback: updateFrequencyFallback,
		updateFrequencyOverride: o.UpdateFrequency,

		tools: tools,
	}, nil
}

func (e *engine) Root() string {
	return e.root
}

func (e *engine) ConfigFile() string {
	return e.configFile
}

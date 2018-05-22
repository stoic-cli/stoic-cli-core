package format

import (
	"github.com/stoic-cli/stoic-cli-core/tool"
)

type StoicConfig struct {
	UpdateFrequency tool.UpdateFrequency  `yaml:"update,omitempty"`
	Tools           map[string]ToolConfig `yaml:",omitempty"`
}

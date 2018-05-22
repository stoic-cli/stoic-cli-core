package format

import (
	"testing"

	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestToolConfig(t *testing.T) {
	data := []struct {
		Name            string
		Config          []byte
		Endpoint        string
		Channel         tool.Channel
		UpdateFrequency tool.UpdateFrequency
		PinVersion      tool.Version
		GetterType      string
		GetterOptions   map[string]interface{}
		RunnerType      string
		RunnerOptions   map[string]interface{}
	}{
		{
			Name:   "Defaults",
			Config: nil,
		},
		{
			Name: "FullySpecified",
			Config: []byte(`
endpoint: github.com/stoic-cli/stoic-cli-core
channel: beta
update: never
pin-version: 0.1.0
getter: {type: labrador}
runner: {type: cheetah}`),
			Endpoint:        "github.com/stoic-cli/stoic-cli-core",
			Channel:         tool.Channel("beta"),
			UpdateFrequency: tool.UpdateNever,
			PinVersion:      tool.Version("0.1.0"),
			GetterType:      "labrador",
			RunnerType:      "cheetah",
		},
		{
			Name:   "GetterOptions",
			Config: []byte(`getter: {option: value, other-option: other-value}`),
			GetterOptions: map[string]interface{}{
				"option":       "value",
				"other-option": "other-value",
			},
		},
		{
			Name:   "RunnerOptions",
			Config: []byte(`runner: {option: value, other-option: other-value}`),
			RunnerOptions: map[string]interface{}{
				"option":       "value",
				"other-option": "other-value",
			},
		},
		{
			Name: "GetterAndRunnerWithOptions",
			Config: []byte(`
getter: {type: fetch, getter-option: value, other-option: other-value}
runner: {type: gallop, runner-option: value, another-option: another-value}`),
			GetterType: "fetch",
			GetterOptions: map[string]interface{}{
				"getter-option": "value",
				"other-option":  "other-value",
			},
			RunnerType: "gallop",
			RunnerOptions: map[string]interface{}{
				"runner-option":  "value",
				"another-option": "another-value",
			},
		},
		{
			Name:       "GetterShorthand",
			Config:     []byte(`getter: labrador`),
			GetterType: "labrador",
		},
		{
			Name:       "RunnerShorthand",
			Config:     []byte(`runner: cheetah`),
			RunnerType: "cheetah",
		},
		{
			Name:            "UpdateDaily",
			Config:          []byte(`update: daily`),
			UpdateFrequency: tool.UpdateDaily,
		},
		{
			Name:            "UpdateWeekly",
			Config:          []byte(`update: weekly`),
			UpdateFrequency: tool.UpdateWeekly,
		},
		{
			Name:            "UpdateMonthly",
			Config:          []byte(`update: monthly`),
			UpdateFrequency: tool.UpdateMonthly,
		},
		{
			Name:            "UpdateAlways",
			Config:          []byte(`update: always`),
			UpdateFrequency: tool.UpdateAlways,
		},
		{
			Name:            "UpdateNever",
			Config:          []byte(`update: never`),
			UpdateFrequency: tool.UpdateNever,
		},
	}

	for _, test := range data {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)

			var config ToolConfig

			err := yaml.Unmarshal(test.Config, &config)
			assert.Nil(err)

			assert.Equal(test.Endpoint, config.Endpoint)
			assert.Equal(test.Channel, config.Channel)
			assert.Equal(test.UpdateFrequency, config.UpdateFrequency)
			assert.Equal(test.PinVersion, config.PinVersion)

			assert.Equal(test.GetterType, config.Getter.Type)
			assert.Len(config.Getter.Options, len(test.GetterOptions))
			for k, v := range test.GetterOptions {
				_ = assert.Contains(config.Getter.Options, k) &&
					assert.Equal(v, config.Getter.Options[k])
			}

			assert.Equal(test.RunnerType, config.Runner.Type)
			assert.Len(config.Runner.Options, len(test.RunnerOptions))
			for k, v := range test.RunnerOptions {
				_ = assert.Contains(config.Runner.Options, k) &&
					assert.Equal(v, config.Runner.Options[k])
			}

		})
	}
}

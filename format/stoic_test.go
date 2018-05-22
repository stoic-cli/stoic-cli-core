package format

import (
	"testing"

	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestStoicConfig(t *testing.T) {
	data := []struct {
		Name            string
		Config          []byte
		UpdateFrequency tool.UpdateFrequency
		ToolNames       []string
	}{
		{
			Name:            "Defaults",
			Config:          nil,
			UpdateFrequency: tool.UpdateDefault,
			ToolNames:       nil,
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
		{
			Name:            "FullyDefined",
			Config:          []byte("update: daily\ntools: {beetle: {}, walrus: {}}\n"),
			UpdateFrequency: tool.UpdateDaily,
			ToolNames:       []string{"beetle", "walrus"},
		},
	}

	for _, test := range data {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)

			var config StoicConfig

			err := yaml.Unmarshal(test.Config, &config)
			assert.Nil(err)

			assert.Equal(test.UpdateFrequency, config.UpdateFrequency)

			assert.Len(config.Tools, len(test.ToolNames))
			for _, name := range test.ToolNames {
				assert.Contains(config.Tools, name)
			}
		})
	}
}

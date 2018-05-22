package format

import (
	"fmt"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

type ToolConfig struct {
	Endpoint        string
	Channel         tool.Channel
	UpdateFrequency tool.UpdateFrequency `yaml:"update,omitempty"`
	PinVersion      tool.Version         `yaml:"pin-version,omitempty"`
	Getter          ToolGetterConfig     `yaml:"getter,omitempty"`
	Runner          ToolRunnerConfig     `yaml:"runner,omitempty"`
}

type ToolGetterConfig TypedOptions
type ToolRunnerConfig TypedOptions

type TypedOptions struct {
	Type    string
	Options map[string]interface{}
}

func (tc *ToolConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data struct {
		Endpoint        string
		Channel         tool.Channel
		UpdateFrequency tool.UpdateFrequency `yaml:"update,omitempty"`
		PinVersion      tool.Version         `yaml:"pin-version,omitempty"`
		Getter          interface{}          `yaml:"getter,omitempty"`
		Runner          interface{}          `yaml:"runner,omitempty"`
	}

	if err := unmarshal(&data); err != nil {
		return err
	}

	tc.Endpoint = data.Endpoint
	tc.Channel = data.Channel
	tc.UpdateFrequency = data.UpdateFrequency
	tc.PinVersion = data.PinVersion

	var to TypedOptions

	if err := toTypedOptions(data.Getter, &to); err != nil {
		return err
	}
	tc.Getter.Type = to.Type
	tc.Getter.Options = to.Options

	if err := toTypedOptions(data.Runner, &to); err != nil {
		return err
	}
	tc.Runner.Type = to.Type
	tc.Runner.Options = to.Options

	return nil
}

func toTypedOptions(data interface{}, to *TypedOptions) error {
	to.Type = ""
	to.Options = map[string]interface{}{}

	if data == nil {
		return nil
	}

	if typ, ok := data.(string); ok {
		to.Type = typ
		return nil
	}

	if mapData, ok := data.(map[interface{}]interface{}); ok {
		for k, v := range mapData {
			stringKey, ok := k.(string)
			if !ok {
				jww.WARN.Printf("ignoring non-string parameter of type %T", stringKey)
				continue
			}

			if stringKey == "type" {
				typ, ok := v.(string)
				if !ok {
					return fmt.Errorf("invalid type: %T", typ)
				}

				to.Type = typ
				continue
			}

			to.Options[stringKey] = v
		}

		return nil
	}

	return fmt.Errorf("expected string or map, got %T", data)
}

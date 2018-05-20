package engine

import (
	"net/url"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core/format"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func (e *engine) Tools() []tool.Tool {
	var tools []tool.Tool
	for name, config := range e.tools {
		tools = append(tools, e.getTool(name, config))
	}
	return tools
}

type engineTool struct {
	name   string
	config format.ToolConfig
	state  State
}

func (e *engine) getTool(name string, config format.ToolConfig) tool.Tool {
	var state State

	url, err := config.Endpoint.MarshalBinary()
	if err == nil {
		state = e.LoadState(string(url))
	} else {
		jww.WARN.Printf(
			"unable to load state for %v, endpoint unmarshalable: %v",
			name, err)
	}

	if config.Getter.Type == "" {
		config.Getter.Type = DefaultToolGetterType
	}
	if config.Runner.Type == "" {
		config.Runner.Type = DefaultToolRunnerType
	}

	return engineTool{
		name:   name,
		config: config,
		state:  state,
	}
}

func (t engineTool) Name() string          { return t.name }
func (t engineTool) Endpoint() *url.URL    { return t.config.Endpoint }
func (t engineTool) Channel() tool.Channel { return t.config.Channel }

func (t engineTool) Version() tool.Version {
	checkout := t.state.CurrentCheckout()
	if checkout == nil {
		return tool.NullVersion
	}
	return checkout.Version()
}

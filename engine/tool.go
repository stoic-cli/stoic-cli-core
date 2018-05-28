package engine

import (
	"net/url"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/format"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

var (
	hasScheme = regexp.MustCompile("^[a-zA-Z][-+.a-zA-Z0-9]*:")
)

func (e *engine) Tools() []stoic.Tool {
	var tools []stoic.Tool
	for name := range e.tools {
		tool, err := e.getTool(name)
		if err == nil {
			tools = append(tools, tool)
		}
	}
	return tools
}

type engineTool struct {
	name     string
	endpoint *url.URL
	config   format.ToolConfig
	state    State
}

func (e *engine) getTool(name string) (stoic.Tool, error) {
	config, ok := e.tools[name]
	if !ok {
		return nil, errors.Errorf("unknown tool, '%v'", name)
	}

	endpoint := config.Endpoint
	if !hasScheme.MatchString(endpoint) {
		endpoint = "https://" + endpoint
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if config.Getter.Type == "" {
		config.Getter.Type = DefaultToolGetterType
	}
	if config.Runner.Type == "" {
		config.Runner.Type = DefaultToolRunnerType
	}

	state := e.LoadState(config.Endpoint)
	return engineTool{
		name:     name,
		endpoint: url,
		config:   config,
		state:    state,
	}, nil
}

func (t engineTool) Name() string              { return t.name }
func (t engineTool) Config() format.ToolConfig { return t.config }
func (t engineTool) Endpoint() *url.URL        { return t.endpoint }
func (t engineTool) Channel() tool.Channel     { return t.config.Channel }

func (t engineTool) IsVersionPinned() bool {
	return t.config.PinVersion != tool.NullVersion
}

func (t engineTool) UpdateFrequency() tool.UpdateFrequency {
	return t.config.UpdateFrequency
}

func (t engineTool) UpstreamVersion() tool.Version {
	return t.state.UpstreamVersion(t.Channel())
}

func (t engineTool) LastUpdate() time.Time {
	return t.state.LastUpstreamUpdate(t.Channel())
}

func (t engineTool) CurrentVersion() tool.Version {
	if t.IsVersionPinned() {
		return t.config.PinVersion
	}

	checkout := t.CurrentCheckout()
	if checkout == nil {
		return tool.NullVersion
	}
	return checkout.Version()
}

func (t engineTool) CurrentCheckout() tool.Checkout {
	if t.IsVersionPinned() {
		return t.CheckoutForVersion(t.config.PinVersion)
	}
	return t.state.CurrentCheckout()
}

func (t engineTool) CheckoutForVersion(version tool.Version) tool.Checkout {
	return t.state.CheckoutForVersion(version)
}

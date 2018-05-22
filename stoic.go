package stoic

import (
	"io"
	"net/url"
	"time"

	"github.com/stoic-cli/stoic-cli-core/format"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

type Cache interface {
	Get(string) io.ReadCloser
	Put(string, io.Reader) error
}

type Tool interface {
	Name() string
	Config() format.ToolConfig

	Endpoint() *url.URL
	Channel() tool.Channel

	IsVersionPinned() bool

	UpdateFrequency() tool.UpdateFrequency
	UpstreamVersion() tool.Version
	LastUpdate() time.Time

	CurrentVersion() tool.Version

	CurrentCheckout() tool.Checkout
	CheckoutForVersion(tool.Version) tool.Checkout
}

type Stoic interface {
	Root() string
	ConfigFile() string

	Parameters() map[string]interface{}

	Cache() Cache

	Tools() []Tool

	RunTool(name string, args []string) error
}

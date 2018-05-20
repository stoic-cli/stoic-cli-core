package stoic

import (
	"io"

	"github.com/stoic-cli/stoic-cli-core/tool"
)

type Cache interface {
	Get(string) io.ReadCloser
	Put(string, io.Reader) error
}

type Stoic interface {
	Root() string
	ConfigFile() string

	Parameters() map[string]interface{}

	Cache() Cache

	Tools() []tool.Tool

	RunTool(name string, args []string) error
}

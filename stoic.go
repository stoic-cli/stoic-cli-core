package stoic

import (
	"io"
)

type Cache interface {
	Get(string) io.ReadCloser
	Put(string, io.Reader) error
}

type Stoic interface {
	Root() string
	ConfigFile() string

	Cache() Cache

	RunTool(name string, args []string) error
}

package tool

import (
	"net/url"
)

type Tool interface {
	Name() string

	Endpoint() *url.URL
	Channel() Channel

	Version() Version
}

package client

import (
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

type Config struct {
	SkipSSL           bool
	Threads           int
	UserAgent         string
	RandomDelay       int
	DisableKeepAlives bool
	Proxys            []string
	RequestTimeout    time.Duration
}

type Context interface {
	Client() HttpClient
}

type CollectorContext struct {
	Context
	client *CollectorClient
}

func NewCollectorContext(config *Config) *CollectorContext {
	client := NewCollectorClient(config, nil)
	return &CollectorContext{client: client}
}

func (c *CollectorContext) Client() HttpClient {
	return c.client
}

func (c *CollectorContext) GetCollector() *crawler.Collector {
	return c.client.GetCollector()
}

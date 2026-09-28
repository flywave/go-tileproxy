package client

import (
	"net/http"

	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type MapClient interface {
	Retrieve(query *layer.MapQuery, format *tile.TileFormat) []byte
	CombinedClient(other MapClient, query *layer.MapQuery) MapClient
}

type BaseClient struct {
	ctx Context
}

// nullHttpClient is used when no context is configured so that an
// unconfigured client degrades to failed requests instead of panicking.
type nullHttpClient struct{}

func (nullHttpClient) Open(url string, data []byte, hdr http.Header) (int, []byte) {
	return 0, nil
}

func (c *BaseClient) GetContext() Context {
	return c.ctx
}

func (c *BaseClient) httpClient() HttpClient {
	if c.ctx == nil {
		return nullHttpClient{}
	}
	return c.ctx.Client()
}

package client

import (
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

func (c *BaseClient) GetContext() Context {
	return c.ctx
}

func (c *BaseClient) httpClient() HttpClient {
	return c.ctx.Client()
}

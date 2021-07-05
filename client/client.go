package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type Client interface {
	Retrieve(*layer.MapQuery, images.ImageFormat) []byte
	CombinedClient(other Client, query *layer.MapQuery) Client
}

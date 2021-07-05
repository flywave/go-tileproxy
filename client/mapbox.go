package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type MapboxClient struct {
	Client
}

func (c *MapboxClient) Retrieve(*layer.MapQuery, images.ImageFormat) []byte {
	return nil
}

func (c *MapboxClient) GetSprite(query *layer.MapQuery) *resource.Sprite {
	return nil
}

func (c *MapboxClient) GetStyle(query *layer.MapQuery) *resource.Style {
	return nil
}

func (c *MapboxClient) GetGlyphs(query *layer.MapQuery) *resource.Glyphs {
	return nil
}

package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type MapboxVectorClient struct {
	Client
}

func (c *MapboxVectorClient) GetVector(*layer.MapQuery) []byte {
	return nil
}

type MapboxRasterClient struct {
	Client
}

func (c *MapboxRasterClient) GetRaster(*layer.MapQuery, images.ImageFormat) []byte {
	return nil
}

type MapboxRasterDemClient struct {
	Client
}

func (c *MapboxRasterDemClient) GetRasterDem(*layer.MapQuery, images.ImageFormat) []byte {
	return nil
}

type MapboxSpriteClient struct {
	Client
}

func (c *MapboxSpriteClient) GetSprite(query *layer.MapQuery) *resource.Sprite {
	return nil
}

type MapboxStyleClient struct {
	Client
}

func (c *MapboxStyleClient) GetStyle(query *layer.MapQuery) *resource.Style {
	return nil
}

type MapboxGlyphsClient struct {
	Client
}

func (c *MapboxGlyphsClient) GetGlyphs(query *layer.MapQuery) *resource.Glyphs {
	return nil
}

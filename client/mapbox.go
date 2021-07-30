package client

import (
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxVectorClient struct {
	BaseClient
}

func (c *MapboxVectorClient) GetVector(*layer.MapQuery) []byte {
	return nil
}

type MapboxRasterClient struct {
	BaseClient
}

func (c *MapboxRasterClient) GetRaster(*layer.MapQuery, tile.TileFormat) []byte {
	return nil
}

type MapboxRasterDemClient struct {
	BaseClient
}

func (c *MapboxRasterDemClient) GetRasterDem(*layer.MapQuery, tile.TileFormat) []byte {
	return nil
}

type MapboxSpriteClient struct {
	BaseClient
}

func (c *MapboxSpriteClient) GetSprite(query *layer.MapQuery) *resource.Sprite {
	return nil
}

type MapboxStyleClient struct {
	BaseClient
}

func (c *MapboxStyleClient) GetStyle(query *layer.MapQuery) *resource.Style {
	return nil
}

type MapboxGlyphsClient struct {
	BaseClient
}

func (c *MapboxGlyphsClient) GetGlyphs(query *layer.MapQuery) *resource.Glyphs {
	return nil
}

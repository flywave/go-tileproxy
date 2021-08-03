package sources

import (
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxTileSource struct {
	Client *client.MapboxTileClient
}

func (s *MapboxTileSource) GetTile(query *layer.TileQuery) tile.Source {
	return nil
}

type MapboxSpriteSource struct {
	Client *client.MapboxSpriteClient
}

func (s *MapboxSpriteSource) GetSprite(query *layer.SpriteQuery) *resource.Sprite {
	return nil
}

type MapboxStyleSource struct {
	Client *client.MapboxStyleClient
}

func (s *MapboxStyleSource) GetStyle(query *layer.StyleQuery) *resource.Style {
	return nil
}

type MapboxGlyphsSource struct {
	Client *client.MapboxGlyphsClient
}

func (s *MapboxGlyphsSource) GetGlyphs(query *layer.GlyphsQuery) *resource.Glyphs {
	return nil
}

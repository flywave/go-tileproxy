package sources

import (
	"errors"

	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

var ErrInvalidBBOX = errors.New("Invalid BBOX")

var ErrInvalidSourceQuery = errors.New("Invalid source query")

type Source interface {
}

type ImagerySource interface {
	Source
	GetMap(query *layer.MapQuery) tile.Source
}

type InfoSource interface {
	Source
	GetInfo(query *layer.MapQuery) *resource.FeatureInfo
}

type LegendSource interface {
	Source
	GetLegend(query *layer.MapQuery) *resource.Legend
}

type StyleSource interface {
	Source
	GetStyle(query *layer.StyleQuery) *resource.Style
}

type SpriteSource interface {
	Source
	GetSprite(query *layer.MapQuery) *resource.Sprite
}

type GlyphsSource interface {
	Source
	GetGlyphs(query *layer.MapQuery) *resource.Glyphs
}

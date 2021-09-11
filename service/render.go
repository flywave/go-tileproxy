package service

import (
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type TileResponse interface {
	getBuffer() []byte
	getTimestamp() *time.Time
	getFormat() string
	GetFormatMime() string
	getSize() int
	getCacheable() bool
}

type Provider interface {
	GetName() string
	GetGrid() *geo.TileGrid
	GetBBox() vec2d.Rect
	GetSrs() geo.Proj
	GetExtent() *geo.MapExtent
	GetFormatMimeType() string
	GetFormat() string
	GetTileBBox(request request.TiledRequest, use_profiles bool, limit bool) (*RequestError, vec2d.Rect)
	Render(tile_request request.TiledRequest, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse)
}

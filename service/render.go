package service

import (
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type RenderResponse interface {
	getBuffer() []byte
	getTimestamp() *time.Time
	getFormat() string
	getSize() int
	getCacheable() bool
}

type RenderLayer interface {
	GetName() string
	GetGrid() *geo.TileGrid
	GetBBox() vec2d.Rect
	GetSrs() geo.Proj
	GetFormat() string
	GetTileBBox(request request.Request, use_profiles bool, limit bool) vec2d.Rect
	Render(tile_request request.Request, use_profiles bool, coverage geo.Coverage, decorate_tile func(image tile.Source) tile.Source) RenderResponse
}

package static

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
)

type TileProvider interface {
	Attribution() string
	Grid() *geo.TileGrid
}

type TileFetcher interface {
	Fetch(tile *cache.Tile) error
}

func NewTileFetcher(p TileProvider) TileFetcher {
	return nil
}

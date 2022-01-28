package static

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
)

type TileProvider interface {
	Attribution() string
	Grid() *geo.TileGrid
	Manager() cache.Manager
}

type TileFetcher interface {
	Fetch(coord [3]int) (*cache.Tile, error)
}

func NewTileFetcher(p TileProvider) TileFetcher {
	return nil
}

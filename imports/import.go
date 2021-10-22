package imports

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type ImportProvider interface {
	Open() error
	GetTileFormat() tile.TileFormat
	GetGrid() *geo.TileGrid
	GetZoomLevels() []int
	GetCoverage() geo.Coverage
	LoadTileCoord(t [3]int) (*cache.Tile, error)
	LoadTileCoords(t [][3]int) (*cache.TileCollection, error)
	Close() error
}

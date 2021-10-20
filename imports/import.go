package imports

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
)

type ImportProvider interface {
	GetFormat() string
	GetFormatMimeType() string
	GetGrid() geo.Grid
	GetCoverage() geo.Coverage
	LoadTileCoord(t [3]int) (*cache.Tile, error)
	LoadTileCoords(t [][3]int) (*cache.TileCollection, error)
}

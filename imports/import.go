package imports

import (
	"errors"
	"strings"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type Import interface {
	Open() error
	GetTileFormat() tile.TileFormat
	GetGrid() *geo.TileGrid
	GetZoomLevels() []int
	GetCoverage() geo.Coverage
	LoadTileCoord(t [3]int, grid *geo.TileGrid) (*cache.Tile, error)
	LoadTileCoords(t [][3]int, grid *geo.TileGrid) (*cache.TileCollection, error)
	Close() error
}

func New(fileName string, opts tile.TileOptions) (Import, error) {
	if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".zip") {
		return NewArchiveImport(fileName, opts)
	} else if strings.HasSuffix(fileName, ".gpkg") {
		return NewGeoPackageImport(fileName, opts)
	} else if strings.HasSuffix(fileName, ".mbtiles") {
		return NewMBTilesImport(fileName, opts)
	}
	return nil, errors.New("import not fount")
}

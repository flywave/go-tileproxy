package imports

import (
	"github.com/flywave/go-geo"
	_ "github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type MBTilesOptions struct {
}

type MBTilesImport struct {
	ImportProvider
	Options     tile.TileOptions
	Grid        geo.Grid
	TileSize    [2]int
	Levels      []int
	Resolutions []float64
	Coverage    geo.Coverage
	Creater     tile.SourceCreater
}

func (a *MBTilesImport) GetTileFormat() tile.TileFormat {
	return a.Options.GetFormat()
}

func (a *MBTilesImport) GetGrid() geo.Grid {
	return a.Grid
}

func (a *MBTilesImport) GetCoverage() geo.Coverage {
	return a.Coverage
}

func (a *MBTilesImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	return nil, nil
}

func (a *MBTilesImport) LoadTileCoords(t [][3]int) (*cache.TileCollection, error) {
	return nil, nil
}

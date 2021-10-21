package imports

import (
	"github.com/flywave/go-geo"
	_ "github.com/flywave/go-gpkg/gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type GPKGOptions struct {
}

type GPKGImport struct {
	ImportProvider
	Options     tile.TileOptions
	Grid        geo.Grid
	TileSize    [2]int
	Levels      []int
	Resolutions []float64
	Coverage    geo.Coverage
	Creater     tile.SourceCreater
}

func (a *GPKGImport) GetTileFormat() tile.TileFormat {
	return a.Options.GetFormat()
}

func (a *GPKGImport) GetGrid() geo.Grid {
	return a.Grid
}

func (a *GPKGImport) GetCoverage() geo.Coverage {
	return a.Coverage
}

func (a *GPKGImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	return nil, nil
}

func (a *GPKGImport) LoadTileCoords(t [][3]int) (*cache.TileCollection, error) {
	return nil, nil
}

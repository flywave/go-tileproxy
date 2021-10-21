package exports

import (
	"github.com/flywave/go-geo"
	_ "github.com/flywave/go-gpkg/gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type GPKGOptions struct {
}

type GPKGExport struct {
	ExportIO
	Format *tile.TileFormat
	Grid   geo.Grid
}

func (a *GPKGExport) GetTileFormat() *tile.TileFormat {
	return a.Format
}

func NewGPKGExport(g geo.Grid) *GPKGExport {
	return &GPKGExport{Grid: g}
}

func (a *GPKGExport) StoreTile(t *cache.Tile) error {
	return nil
}

func (a *GPKGExport) StoreTileCollection(ts *cache.TileCollection) error {
	return nil
}

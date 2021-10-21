package exports

import (
	"github.com/flywave/go-geo"
	_ "github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type MBTilesOptions struct {
}

type MBTilesExport struct {
	ExportIO
	Format *tile.TileFormat
	Grid   geo.Grid
}

func NewMBTilesExport(g geo.Grid) *MBTilesExport {
	return &MBTilesExport{Grid: g}
}

func (a *MBTilesExport) GetTileFormat() *tile.TileFormat {
	return nil
}

func (a *MBTilesExport) StoreTile(t *cache.Tile) error {
	return nil
}

func (a *MBTilesExport) StoreTileCollection(ts *cache.TileCollection) error {
	return nil
}

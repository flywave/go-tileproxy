package exports

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoTIFFExport struct {
	Export
	Uri string
}

func (e *GeoTIFFExport) GetTileFormat() tile.TileFormat {
	return tile.TileFormat("")
}

func (e *GeoTIFFExport) StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
	return nil
}

func (e *GeoTIFFExport) StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error {
	return nil
}

func (e *GeoTIFFExport) Close() error {
	return nil
}

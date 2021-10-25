package exports

import (
	"errors"
	"path"
	"strings"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type Export interface {
	GetTileFormat() tile.TileFormat
	StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error
	StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error
	Close() error
}

func New(fileName string, g *geo.TileGrid, optios tile.TileOptions, settings map[string]interface{}) (Export, error) {
	if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".zip") {
		if dl, ok := settings["directory_layout"]; ok {
			return NewArchiveExport(fileName, g, optios, dl.(string))
		}
		return NewArchiveExport(fileName, g, optios, mbtiles.DEFAULT_DIRECTORY_LAYOUT)
	} else if strings.HasSuffix(fileName, ".gpkg") {
		if tname, ok := settings["table_name"]; ok {
			return NewArchiveExport(fileName, g, optios, tname.(string))
		}
		tableName := path.Base(fileName)
		tableName = strings.TrimSuffix(tableName, ".gpkg")
		return NewGeoPackageExport(fileName, tableName, g, optios)
	} else if strings.HasSuffix(fileName, ".mbtiles") {
		return NewMBTilesExport(fileName, g, optios)
	}
	return nil, errors.New("export not fount")
}

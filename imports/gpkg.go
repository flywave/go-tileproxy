package imports

import (
	"errors"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoPackageImport struct {
	ImportProvider
	filename  string
	options   tile.TileOptions
	creater   tile.SourceCreater
	db        *gpkg.GeoPackage
	tableName string
}

func NewGeoPackageImport(filename string, opts tile.TileOptions) *GeoPackageImport {
	return &GeoPackageImport{filename: filename, options: opts}
}

func (a *GeoPackageImport) Open() error {
	a.db = gpkg.New(a.filename)

	if !a.db.Exists() {
		return errors.New("file not found")
	}

	if err := a.db.Init(); err != nil {
		return err
	}

	if tms, err := a.db.GetTileMatrixSets(); err != nil {
		return err
	} else if len(tms) > 0 {
		a.tableName = tms[0].Name
	} else {
		return errors.New("not found tile table")
	}

	if a.options == nil {
		format, err := a.db.GetTileFormat(a.tableName)
		if err != nil {
			return nil
		}

		switch format.String() {
		case "png":
			a.options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
		case "jpg":
			a.options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
		case "webp":
			a.options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
		case "pbf":
			a.options = &imagery.ImageOptions{Format: tile.TileFormat("mvt")}
		}
	}

	if a.options != nil {
		a.creater = cache.GetSourceCreater(a.options)
	} else {
		return errors.New("format not found")
	}

	return nil
}

func (a *GeoPackageImport) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *GeoPackageImport) GetTileFormat() tile.TileFormat {
	return a.options.GetFormat()
}

func (a *GeoPackageImport) GetGrid() *geo.TileGrid {
	grid, err := a.db.GetTileGrid(a.tableName)
	if err != nil {
		return nil
	}

	return grid
}

func (a *GeoPackageImport) GetCoverage() geo.Coverage {
	cov, err := a.db.GetCoverage(a.tableName)
	if err != nil {
		return nil
	}

	return cov
}

func (a *GeoPackageImport) GetZoomLevels() []int {
	levels, err := a.db.GetTileZoomLevels(a.tableName)
	if err != nil {
		return nil
	}
	return levels
}

func (a *GeoPackageImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	data, err := a.db.GetTile(a.tableName, t[2], t[0], t[1])
	if err != nil {
		return nil, err
	}
	tile := cache.NewTile(t)
	tile.Source = a.creater.Create(data, tile.Coord)
	return tile, nil
}

func (a *GeoPackageImport) LoadTileCoords(t [][3]int) (*cache.TileCollection, error) {
	var errs error
	tiles := cache.NewTileCollection(nil)
	for _, tc := range t {
		if t, err := a.LoadTileCoord(tc); err != nil {
			errs = err
		} else if t != nil {
			tiles.SetItem(t)
		}
	}
	return tiles, errs
}

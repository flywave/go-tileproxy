package imports

import (
	"errors"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoPackageImport struct {
	ImportProvider
	filename string
	Options  tile.TileOptions
	Levels   []int
	Creater  tile.SourceCreater
	db       *gpkg.GeoPackage
	tiles    gpkg.TileMatrixSet
}

func NewGeoPackageImport(filename string) *GeoPackageImport {
	return &GeoPackageImport{filename: filename}
}

func (a *GeoPackageImport) Open() error {
	a.db = gpkg.New(a.filename)

	if !a.db.Exists() {
		return errors.New("file not found!")
	}

	if err := a.db.Init(); err != nil {
		return err
	}

	if tms, err := a.db.GetTileMatrixSets(); err != nil {
		return err
	} else if len(tms) > 0 {
		a.tiles = tms[0]
	} else {
		return errors.New("not found tile table!")
	}

	format, err := a.db.GetTileFormat(a.tiles.Name)
	if err != nil {
		return nil
	}
	switch format.String() {
	case "png":
		a.Options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "jpg":
		a.Options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "webp":
		a.Options = &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "pbf":
		a.Options = &imagery.ImageOptions{Format: tile.TileFormat("mvt")}
	}

	if a.Options != nil {
		a.Creater = cache.GetSourceCreater(a.Options)
	} else {
		return errors.New("format not found!")
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
	return a.Options.GetFormat()
}

func (a *GeoPackageImport) GetGrid() geo.Grid {
	_, res, err := a.db.GetZoomLevelsAndResolutions(a.tiles.Name)
	if err != nil {
		return nil
	}
	tileWith, err := a.db.GetTileWidth(a.tiles.Name)
	if err != nil {
		return nil
	}
	tileHeight, err := a.db.GetTileHeight(a.tiles.Name)
	if err != nil {
		return nil
	}
	srsid, err := a.db.GetTileSrsId(a.tiles.Name)
	if err != nil {
		return nil
	}

	conf := geo.DefaultTileGridOptions()

	conf[geo.TILEGRID_SRS] = geo.NewProj(srsid)
	conf[geo.TILEGRID_RES] = res
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{uint32(tileWith), uint32(tileHeight)}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	return geo.NewTileGrid(conf)
}

func (a *GeoPackageImport) GetCoverage() geo.Coverage {
	bbox := vec2d.Rect{Min: vec2d.T{*a.tiles.MinX, *a.tiles.MinY}, Max: vec2d.T{*a.tiles.MaxX, *a.tiles.MaxY}}
	prj := geo.NewProj(*a.tiles.SpatialReferenceSystemId)
	return geo.NewBBoxCoverage(bbox, prj, false)
}

func (a *GeoPackageImport) GetZoomLevels() []int {
	levels, _, err := a.db.GetZoomLevelsAndResolutions(a.tiles.Name)
	if err != nil {
		return nil
	}
	return levels
}

func (a *GeoPackageImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	data, err := a.db.GetTile(a.tiles.Name, t[2], t[0], t[1])
	if err != nil {
		return nil, err
	}
	tile := cache.NewTile(t)
	tile.Source = a.Creater.Create(data, tile.Coord)
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

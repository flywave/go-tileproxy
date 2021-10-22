package imports

import (
	"errors"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

type MBTilesImport struct {
	ImportProvider
	filename string
	md       *mbtiles.Metadata
	options  tile.TileOptions
	grid     geo.Grid
	coverage geo.Coverage
	creater  tile.SourceCreater
	db       *mbtiles.DB
}

func NewMBTilesImport(filename string, opts tile.TileOptions) *MBTilesImport {
	return &MBTilesImport{filename: filename, options: opts}
}

func (a *MBTilesImport) Open() error {
	var err error
	a.db, err = mbtiles.NewDB(a.filename)
	if err != nil {
		return err
	}

	a.md, err = a.db.GetMetadata()
	if err != nil {
		return err
	}

	if a.options == nil {
		a.options = a.getTileOptions(a.md)
	}

	a.grid = a.getTileGrid(a.md)
	a.coverage = a.getTileCoverage(a.md)
	a.creater = cache.GetSourceCreater(a.options)

	if a.options == nil || a.grid == nil || a.coverage == nil || a.creater == nil {
		return errors.New("count open mbtiles")
	}

	return nil
}

func (a *MBTilesImport) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *MBTilesImport) GetTileFormat() tile.TileFormat {
	return a.options.GetFormat()
}

func (a *MBTilesImport) GetExtension() string {
	format := a.GetTileFormat()
	return format.Extension()
}

func (a *MBTilesImport) GetGrid() geo.Grid {
	return a.grid
}

func (a *MBTilesImport) GetCoverage() geo.Coverage {
	return a.coverage
}

func (a *MBTilesImport) GetZoomLevels() []int {
	rets := []int{}
	for i := a.md.MinZoom; i <= a.md.MaxZoom; i++ {
		rets = append(rets, i)
	}
	return rets
}

func (a *MBTilesImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	var data []byte
	err := a.db.ReadTile(uint8(t[2]), uint64(t[0]), uint64(t[1]), &data)
	if err != nil {
		return nil, err
	}
	tile := cache.NewTile(t)
	tile.Source = a.creater.Create(data, tile.Coord)
	return tile, nil
}

func (a *MBTilesImport) LoadTileCoords(t [][3]int) (*cache.TileCollection, error) {
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

func (a *MBTilesImport) getTileOptions(md *mbtiles.Metadata) tile.TileOptions {
	format := md.Format.String()
	switch format {
	case "png":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "jpg":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "webp":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "pbf":
		return &vector.VectorOptions{Format: tile.TileFormat("mvt")}
	}
	return nil
}

func (a *MBTilesImport) getTileGrid(md *mbtiles.Metadata) geo.Grid {
	conf := geo.DefaultTileGridOptions()
	if md.Srs == "" {
		conf[geo.TILEGRID_SRS] = geo.NewProj(mbtiles.DEFAULT_SRS)
	} else {
		conf[geo.TILEGRID_SRS] = geo.NewProj(md.Srs)
	}
	if md.Origin == "" {
		conf[geo.TILEGRID_ORIGIN] = geo.OriginFromString(mbtiles.DEFAULT_ORIGIN)
	} else {
		conf[geo.TILEGRID_ORIGIN] = geo.OriginFromString(md.Origin)
	}
	if md.ResFactor == nil {
		conf[geo.TILEGRID_RES_FACTOR] = mbtiles.DEFAULT_RES_FACTOR
	} else {
		conf[geo.TILEGRID_RES_FACTOR] = md.ResFactor
	}

	if md.TileSize != nil {
		conf[geo.TILEGRID_TILE_SIZE] = []uint32{uint32(md.TileSize[0]), uint32(md.TileSize[1])}
	}

	return geo.NewTileGrid(conf)
}

func (a *MBTilesImport) getTileCoverage(md *mbtiles.Metadata) geo.Coverage {
	bbox := vec2d.Rect{Min: vec2d.T{md.Bounds[0], md.Bounds[1]}, Max: vec2d.T{md.Bounds[2], md.Bounds[3]}}
	var prj geo.Proj
	if md.BoundsSrs != "" {
		prj = geo.NewProj(md.BoundsSrs)
	} else {
		prj = geo.NewProj("EPSG:4326")
	}
	return geo.NewBBoxCoverage(bbox, prj, false)
}

package imports

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	"github.com/mholt/archiver/v3"
)

type ArchiveImport struct {
	ImportProvider
	fileName     string
	tempDir      string
	md           *mbtiles.Metadata
	Options      tile.TileOptions
	Grid         geo.Grid
	Coverage     geo.Coverage
	Creater      tile.SourceCreater
	tileLocation func(*cache.Tile, string, string, bool) string
}

func NewArchiveImport(fileName string) *ArchiveImport {
	return &ArchiveImport{fileName: fileName}
}

func (a *ArchiveImport) Open() error {
	p, err := ioutil.TempDir(os.TempDir(), "import")

	if err != nil {
		return err
	}

	a.tempDir = p

	err = archiver.Unarchive(a.fileName, a.tempDir)
	if err != nil {
		return err
	}

	mdpath := path.Join(a.tempDir, mbtiles.METADATA_JSON)

	if !utils.FileExists(mdpath) {
		return errors.New("not found metadata.json")
	}

	f, err := os.Open(mdpath)
	if err != nil {
		return err
	}

	mddata, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var md mbtiles.Metadata
	err = json.Unmarshal(mddata, &md)
	if err != nil {
		return err
	}

	a.md = &md
	a.Options = a.getTileOptions(a.md)
	a.Grid = a.getTileGrid(a.md)
	a.Coverage = a.getTileCoverage(a.md)
	a.Creater = cache.GetSourceCreater(a.Options)

	var directoryLayout string
	if a.md.DirectoryLayout == "" {
		directoryLayout = mbtiles.DEFAULT_DIRECTORY_LAYOUT
	} else {
		directoryLayout = a.md.DirectoryLayout
	}

	pathLoc, _, err := cache.LocationPaths(directoryLayout)
	if err != nil {
		return nil
	}

	a.tileLocation = pathLoc

	return nil
}

func (a *ArchiveImport) Close() error {
	return os.Remove(a.tempDir)
}

func (a *ArchiveImport) GetTileFormat() tile.TileFormat {
	return a.Options.GetFormat()
}

func (a *ArchiveImport) GetExtension() string {
	format := a.GetTileFormat()
	return format.Extension()
}

func (a *ArchiveImport) GetGrid() geo.Grid {
	return a.Grid
}

func (a *ArchiveImport) GetCoverage() geo.Coverage {
	return a.Coverage
}

func (a *ArchiveImport) GetZoomLevels() []int {
	rets := []int{}
	for i := a.md.MinZoom; i <= a.md.MaxZoom; i++ {
		rets = append(rets, i)
	}
	return rets
}

func (a *ArchiveImport) LoadTileCoord(t [3]int) (*cache.Tile, error) {
	tile := cache.NewTile(t)
	location := a.TileLocation(tile)

	if utils.FileExists(location) {
		data, _ := os.ReadFile(location)
		tile.Source = a.Creater.Create(data, tile.Coord)
		return tile, nil
	}
	return nil, nil
}

func (a *ArchiveImport) LoadTileCoords(t [][3]int) (*cache.TileCollection, error) {
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

func (a *ArchiveImport) TileLocation(tile *cache.Tile) string {
	return a.tileLocation(tile, "", a.GetExtension(), false)
}

func (a *ArchiveImport) buildTilePath(tile *cache.Tile) string {
	return a.TileLocation(tile)
}

func (a *ArchiveImport) getTileOptions(md *mbtiles.Metadata) tile.TileOptions {
	format := md.Format.String()
	switch format {
	case "png":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "jpg":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "webp":
		return &imagery.ImageOptions{Format: tile.TileFormat(format)}
	case "pbf":
		return &imagery.ImageOptions{Format: tile.TileFormat("mvt")}
	}
	return nil
}

func (a *ArchiveImport) getTileGrid(md *mbtiles.Metadata) geo.Grid {
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

func (a *ArchiveImport) getTileCoverage(md *mbtiles.Metadata) geo.Coverage {
	bbox := vec2d.Rect{Min: vec2d.T{md.Bounds[0], md.Bounds[1]}, Max: vec2d.T{md.Bounds[2], md.Bounds[3]}}
	var prj geo.Proj
	if md.BoundsSrs != "" {
		prj = geo.NewProj(md.BoundsSrs)
	} else {
		prj = geo.NewProj("EPSG:4326")
	}
	return geo.NewBBoxCoverage(bbox, prj, false)
}

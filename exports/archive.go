package exports

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/mholt/archiver/v3"
)

const (
	TARGZ = ".tar.gz"
	ZIP   = ".zip"
)

type ArchiveExport struct {
	Export
	Name         string
	layout       string
	optios       tile.TileOptions
	grid         *geo.TileGrid
	writer       archiver.Writer
	bounds       vec2d.Rect
	boundsSrs    geo.Proj
	minZoom      int
	maxZoom      int
	tileLocation func(*cache.Tile, string, string, bool) string
}

func tileFormatToMBTileFormat(t tile.TileFormat) mbtiles.TileFormat {
	switch t {
	case "png":
		return mbtiles.PNG
	case "jpg":
		return mbtiles.JPG
	case "pbf":
		return mbtiles.PBF
	case "webp":
		return mbtiles.WEBP
	default:
		return mbtiles.UNKNOWN
	}
}

func NewArchiveExport(filename string, g *geo.TileGrid, optios tile.TileOptions, directoryLayout string) (*ArchiveExport, error) {
	pathLoc, _, err := cache.LocationPaths(directoryLayout)
	if err != nil {
		return nil, err
	}

	var writer archiver.Writer
	if strings.HasSuffix(filename, TARGZ) {
		writer = archiver.NewTarGz()
	} else if strings.HasSuffix(filename, ZIP) {
		writer = archiver.NewZip()
	} else {
		return nil, errors.New("only support .tar.gz or .zip")
	}

	if f, err := os.Create(filename); err != nil {
		return nil, err
	} else {
		if err := writer.Create(f); err != nil {
			return nil, err
		}
	}

	return &ArchiveExport{
		Name:         filename,
		grid:         g,
		layout:       directoryLayout,
		writer:       writer,
		optios:       optios,
		bounds:       vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal},
		minZoom:      30,
		maxZoom:      0,
		tileLocation: pathLoc,
		boundsSrs:    geo.NewProj("EPSG:4326"),
	}, nil
}

func (a *ArchiveExport) GetTileFormat() tile.TileFormat {
	return a.optios.GetFormat()
}

func (a *ArchiveExport) GetExtension() string {
	format := a.GetTileFormat()
	return format.Extension()
}

func (a *ArchiveExport) StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
	return a.writeTile(t, srcGrid)
}

func (a *ArchiveExport) StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error {
	for _, t := range ts.GetSlice() {
		if err := a.writeTile(t, srcGrid); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArchiveExport) Close() error {
	md := a.buildMetadata()
	data, _ := json.Marshal(md)

	p, err := ioutil.TempFile(os.TempDir(), "metadata-")

	if err != nil {
		return err
	}

	defer func() {
		p.Close()
		os.Remove(p.Name())
	}()

	p.Write(data)

	p.Sync()

	info, err := p.Stat()

	if err != nil {
		return err
	}

	err = a.writer.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   info,
			CustomName: mbtiles.METADATA_JSON,
		},
		ReadCloser: &readerCloser{buf: bytes.NewBuffer(data)},
	})

	if err != nil {
		return err
	}

	return a.writer.Close()
}

func (a *ArchiveExport) buildMetadata() *mbtiles.Metadata {
	md := &mbtiles.Metadata{
		Name:            a.Name,
		Format:          tileFormatToMBTileFormat(a.GetTileFormat()),
		Bounds:          [4]float64{a.bounds.Min[0], a.bounds.Min[1], a.bounds.Max[0], a.bounds.Max[1]},
		Center:          [3]float64{(a.bounds.Max[0] + a.bounds.Min[0]) / 2, (a.bounds.Max[1] + a.bounds.Min[1]) / 2, 0},
		MinZoom:         a.minZoom,
		MaxZoom:         a.maxZoom,
		Type:            mbtiles.Overlay,
		DirectoryLayout: a.layout,
		Origin:          geo.OriginToString(a.grid.Origin),
		Srs:             a.grid.Srs.GetSrsCode(),
		BoundsSrs:       a.boundsSrs.GetSrsCode(),
	}

	if a.grid.Levels == 40 {
		md.ResFactor = "sqrt2"
	} else {
		md.ResFactor = 2.0
	}

	md.TileSize = new([2]int)

	md.TileSize[0] = int(a.grid.TileSize[0])
	md.TileSize[1] = int(a.grid.TileSize[1])

	return md
}

func (a *ArchiveExport) TileLocation(tile *cache.Tile) string {
	tile.Location = ""
	return a.tileLocation(tile, "", a.GetExtension(), false)
}

type readerCloser struct {
	buf *bytes.Buffer
}

func (r *readerCloser) Close() error {
	return nil
}

func (r *readerCloser) Read(p []byte) (n int, err error) {
	return r.buf.Read(p)
}

func (a *ArchiveExport) expand(t *cache.Tile) error {
	bbox := a.grid.TileBBox(t.Coord, false)
	bbox = a.grid.Srs.TransformRectTo(a.boundsSrs, bbox, 16)
	a.bounds.Join(&bbox)

	if a.minZoom > t.Coord[2] {
		a.minZoom = t.Coord[2]
	}

	if a.maxZoom < t.Coord[2] {
		a.maxZoom = t.Coord[2]
	}

	return nil
}

func (a *ArchiveExport) writeTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
	dc, err := cache.TransformCoord(t.Coord, srcGrid, a.grid)

	if err != nil {
		return err
	}

	dstTile := *t
	dstTile.Coord = dc

	data, err := cache.EncodeTile(a.optios, dstTile.Coord, dstTile.Source)

	if err != nil {
		return err
	}

	name := a.TileLocation(&dstTile)

	p, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("*-%s", path.Base(name)))

	if err != nil {
		return err
	}

	defer func() {
		p.Close()
		os.Remove(p.Name())
	}()

	p.Write(data)

	p.Sync()

	info, err := p.Stat()

	if err != nil {
		return err
	}

	err = a.writer.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   info,
			CustomName: name,
		},
		ReadCloser: &readerCloser{buf: bytes.NewBuffer(data)},
	})

	if err == nil {
		a.expand(&dstTile)
	}

	return err
}

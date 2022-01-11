package exports

import (
	"bytes"
	"compress/gzip"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type MBTilesExport struct {
	Export
	Name      string
	Uri       string
	optios    tile.TileOptions
	grid      *geo.TileGrid
	db        *mbtiles.DB
	bounds    vec2d.Rect
	boundsSrs geo.Proj
	minZoom   int
	maxZoom   int
}

func NewMBTilesExport(uri string, g *geo.TileGrid, optios tile.TileOptions) (*MBTilesExport, error) {
	db, err := mbtiles.CreateDB(uri, tileFormatToMBTileFormat(optios.GetFormat()), nil)
	if err != nil {
		return nil, err
	}
	return &MBTilesExport{
		Uri:    uri,
		grid:   g,
		optios: optios,
		bounds: vec2d.Rect{
			Min: vec2d.MaxVal,
			Max: vec2d.MinVal,
		},
		db:        db,
		boundsSrs: geo.NewProj("EPSG:4326"),
		minZoom:   0,
		maxZoom:   30,
	}, nil
}

func (a *MBTilesExport) GetTileFormat() tile.TileFormat {
	return a.optios.GetFormat()
}

func (a *MBTilesExport) GetExtension() string {
	format := a.GetTileFormat()
	return format.Extension()
}

func (a *MBTilesExport) StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
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

	if a.GetExtension() == "pbf" || a.GetExtension() == "mvt" {
		var in bytes.Buffer
		w := gzip.NewWriter(&in)
		w.Write(data)
		w.Close()
		data = in.Bytes()
	}

	if err := a.db.StoreTile(uint8(dc[2]), uint64(dc[0]), uint64(dc[1]), data); err != nil {
		return err
	} else {
		a.expand(&dstTile)
	}

	return nil
}

func (a *MBTilesExport) StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error {
	for _, t := range ts.GetSlice() {
		if err := a.StoreTile(t, srcGrid); err != nil {
			return err
		}
	}
	return nil
}

func (a *MBTilesExport) Close() error {
	md := a.buildMetadata()
	err := a.db.UpdateMetadata(md)

	if err != nil {
		return err
	}

	return a.db.Close()
}

func (a *MBTilesExport) buildMetadata() *mbtiles.Metadata {
	md := &mbtiles.Metadata{
		Name:            a.Name,
		Format:          tileFormatToMBTileFormat(a.GetTileFormat()),
		Bounds:          [4]float64{a.bounds.Min[0], a.bounds.Min[1], a.bounds.Max[0], a.bounds.Max[1]},
		Center:          [3]float64{(a.bounds.Max[0] + a.bounds.Min[0]) / 2, (a.bounds.Max[1] + a.bounds.Min[1]) / 2, 0},
		MinZoom:         a.minZoom,
		MaxZoom:         a.maxZoom,
		Type:            mbtiles.Overlay,
		DirectoryLayout: "",
		Origin:          geo.OriginToString(a.grid.Origin),
		Srs:             a.grid.Srs.SrsCode,
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

func (a *MBTilesExport) expand(t *cache.Tile) error {
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

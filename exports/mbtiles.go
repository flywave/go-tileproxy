package exports

import (
	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type MBTilesExport struct {
	ExportIO
	Name    string
	Uri     string
	Optios  tile.TileOptions
	Grid    *geo.TileGrid
	db      *mbtiles.DB
	bounds  vec2d.Rect
	minZoom int
	maxZoom int
}

func NewMBTilesExport(uri string, g *geo.TileGrid, optios tile.TileOptions) (*MBTilesExport, error) {
	db, err := mbtiles.CreateDB(uri, tileFormatToMBTileFormat(optios.GetFormat()), nil)
	if err != nil {
		return nil, err
	}
	return &MBTilesExport{Uri: uri, Grid: g, Optios: optios, bounds: vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}, db: db}, nil
}

func (a *MBTilesExport) GetTileFormat() tile.TileFormat {
	return a.Optios.GetFormat()
}

func (a *MBTilesExport) StoreTile(t *cache.Tile) error {
	data := t.Source.GetBuffer(nil, a.Optios)

	if err := a.db.StoreTile(uint8(t.Coord[2]), uint64(t.Coord[0]), uint64(t.Coord[1]), data); err != nil {
		return err
	} else {
		a.expand(t)
	}

	return nil
}

func (a *MBTilesExport) StoreTileCollection(ts *cache.TileCollection) error {
	for _, t := range ts.GetSlice() {
		if err := a.StoreTile(t); err != nil {
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
		MinZoom:         a.minZoom,
		MaxZoom:         a.maxZoom,
		Type:            mbtiles.Overlay,
		DirectoryLayout: "",
		Origin:          geo.OriginToString(a.Grid.Origin),
		Srs:             a.Grid.Srs.SrsCode,
		BoundsSrs:       a.Grid.Srs.SrsCode,
	}

	if a.Grid.Levels == 40 {
		md.ResFactor = "sqrt2"
	} else {
		md.ResFactor = 2.0
	}

	md.TileSize = new([2]int)

	md.TileSize[0] = int(a.Grid.TileSize[0])
	md.TileSize[1] = int(a.Grid.TileSize[1])

	return nil
}

func (a *MBTilesExport) expand(t *cache.Tile) error {
	bbox := a.Grid.TileBBox(t.Coord, false)
	a.bounds.Join(&bbox)

	if a.minZoom > t.Coord[2] {
		a.minZoom = t.Coord[2]
	}

	if a.maxZoom < t.Coord[2] {
		a.maxZoom = t.Coord[2]
	}

	return nil
}

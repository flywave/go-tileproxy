package exports

import (
	"bytes"
	"compress/gzip"
	"errors"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoPackageExport struct {
	Export
	Uri    string
	Name   string
	optios tile.TileOptions
	grid   *geo.TileGrid
	db     *gpkg.GeoPackage
	bounds vec2d.Rect
}

func NewGeoPackageExport(uri string, name string, g *geo.TileGrid, optios tile.TileOptions) (*GeoPackageExport, error) {
	if g.Origin != geo.ORIGIN_UL {
		return nil, errors.New("gpkg only support ul origin")
	}

	ge := &GeoPackageExport{
		Uri:    uri,
		Name:   name,
		grid:   g,
		optios: optios,
		bounds: vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal},
		db:     gpkg.Create(uri),
	}

	if err := ge.db.AddTilesTable(name, g, nil); err != nil {
		return nil, err
	}

	return ge, nil
}

func (a *GeoPackageExport) GetTileFormat() tile.TileFormat {
	return a.optios.GetFormat()
}

func (a *GeoPackageExport) GetExtension() string {
	format := a.GetTileFormat()
	return format.Extension()
}

func (a *GeoPackageExport) StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
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

	// if a.GetExtension() == "pbf" || a.GetExtension() == "mvt" {
	var in bytes.Buffer
	w := gzip.NewWriter(&in)
	w.Write(data)
	w.Close()
	data = in.Bytes()
	// }

	if err := a.db.StoreTile(a.Name, dc[2], dc[0], dc[1], data); err != nil {
		return err
	} else {
		a.expand(&dstTile)
	}

	return nil
}

func (a *GeoPackageExport) expand(t *cache.Tile) error {
	bbox := a.grid.TileBBox(t.Coord, false)
	a.bounds.Join(&bbox)
	return nil
}

func (a *GeoPackageExport) StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error {
	for _, t := range ts.GetSlice() {
		if err := a.StoreTile(t, srcGrid); err != nil {
			return err
		}
	}
	return nil
}

func (a *GeoPackageExport) Close() error {
	err := a.db.UpdateGeometryExtent(a.Name, &general.Extent{a.bounds.Min[0], a.bounds.Min[1], a.bounds.Max[0], a.bounds.Max[1]})

	if err != nil {
		return err
	}

	return a.db.Close()
}

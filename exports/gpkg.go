package exports

import (
	"errors"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-gpkg"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoPackageExport struct {
	ExportIO
	Uri    string
	Name   string
	Optios tile.TileOptions
	Grid   geo.Grid
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
		Grid:   g,
		Optios: optios,
		bounds: vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal},
		db:     gpkg.Create(uri),
	}

	if err := ge.db.AddTilesTable(name, g, nil); err != nil {
		return nil, err
	}

	return ge, nil
}

func (a *GeoPackageExport) GetTileFormat() tile.TileFormat {
	return a.Optios.GetFormat()
}

func (a *GeoPackageExport) StoreTile(t *cache.Tile) error {
	data := t.Source.GetBuffer(nil, a.Optios)

	if err := a.db.StoreTile(a.Name, t.Coord[2], t.Coord[0], t.Coord[1], data); err != nil {
		return err
	} else {
		a.expand(t)
	}

	return nil
}

func (a *GeoPackageExport) expand(t *cache.Tile) error {
	bbox := a.Grid.TileBBox(t.Coord, false)
	a.bounds.Join(&bbox)
	return nil
}

func (a *GeoPackageExport) StoreTileCollection(ts *cache.TileCollection) error {
	for _, t := range ts.GetSlice() {
		if err := a.StoreTile(t); err != nil {
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

package setting

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-geom"
	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-geos"
	"github.com/flywave/go3d/float64/vec2"
)

func loadExpireTiles(tiles [][3]int, grid *geo.TileGrid) []*geos.Geometry {
	if grid == nil {
		opts := geo.DefaultTileGridOptions()
		opts[geo.TILEGRID_SRS] = "EPSG:3857"
		opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_NW
		grid = geo.NewTileGrid(opts)
	}

	boxes := []*geos.Geometry{}
	for _, tile := range tiles {
		geom := geo.BBoxPolygon(grid.TileBBox(tile, false))
		boxes = append(boxes, geom)
	}

	return boxes
}

func LoadCoverage(cov *Coverage) geo.Coverage {
	clip := false
	if cov.Clip != nil {
		clip = *cov.Clip
	}

	if cov.Union != nil {
		parts := []geo.Coverage{}
		for _, cov := range cov.Union {
			parts = append(parts, LoadCoverage(&cov))
		}
		return geo.UnionCoverage(parts, clip)
	} else if cov.Intersection != nil {
		parts := []geo.Coverage{}
		for _, cov := range cov.Intersection {
			parts = append(parts, LoadCoverage(&cov))
		}
		return geo.IntersectionCoverage(parts, clip)
	} else if cov.Difference != nil {
		parts := []geo.Coverage{}
		for _, cov := range cov.Difference {
			parts = append(parts, LoadCoverage(&cov))
		}
		return geo.DiffCoverage(parts, clip)
	} else if cov.Polygons != "" {
		geom := geos.CreateFromWKT(cov.Polygons)
		return geo.NewGeosCoverage(geom, geo.NewProj(cov.PolygonsSrs), clip)
	} else if cov.Geometry != "" {
		geomdata, _ := geom.UnmarshalGeometry([]byte(cov.Geometry))
		geom := general.GeometryDataAsGeometry(geomdata)
		return geo.NewGeomCoverage(geom, geo.NewProj(cov.PolygonsSrs), clip)
	} else if cov.BBox != nil {
		return geo.NewBBoxCoverage(vec2.Rect{
			Min: vec2.T{cov.BBox[0], cov.BBox[1]},
			Max: vec2.T{cov.BBox[2], cov.BBox[3]},
		}, geo.NewProj(cov.BBoxSrs), clip)
	} else if cov.ExpireTiles != nil {
		geoms := loadExpireTiles(cov.ExpireTiles, nil)
		multipolygon := geos.CreateEmptyPolygon()
		for _, g := range geoms {
			multipolygon = multipolygon.Union(g)
		}
		return geo.NewGeosCoverage(multipolygon, geo.NewProj(3857), clip)
	}
	return nil
}

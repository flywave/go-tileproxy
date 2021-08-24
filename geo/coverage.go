package geo

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geom"
	"github.com/flywave/go-geos"
)

type Coverage interface {
	GetSrs() Proj
	GetBBox() vec2d.Rect
	GetExtent() *MapExtent
	GetGeom() *geos.Geometry
	Intersects(bbox vec2d.Rect, srs Proj) bool
	Contains(bbox vec2d.Rect, srs Proj) bool
	ContainsPoint(pt []float64, srs Proj) bool
	Intersection(bbox vec2d.Rect, srs Proj) Coverage
	TransformTo(srs Proj) Coverage
	Equals(cc Coverage) bool
	IsClip() bool
}

type MultiCoverage struct {
	Coverage
	Coverages []Coverage
	BBox      vec2d.Rect
}

func getExtent(coves []Coverage) vec2d.Rect {
	if len(coves) != 0 {
		ret := coves[0].GetBBox()
		for _, b := range coves[1:] {
			bbox := b.GetBBox()
			ret.Join(&bbox)
		}
		return ret
	}
	return vec2d.NewRect(&vec2d.Zero, &vec2d.Zero)
}

func NewMultiCoverage(coves []Coverage) *MultiCoverage {
	return &MultiCoverage{Coverages: coves, BBox: getExtent(coves)}
}

func (c *MultiCoverage) IsClip() bool {
	if len(c.Coverages) == 0 {
		return false
	}
	return c.Coverages[0].IsClip()
}

func (c *MultiCoverage) GetSrs() Proj {
	if len(c.Coverages) == 0 {
		return nil
	}
	return c.Coverages[0].GetSrs()
}

func (c *MultiCoverage) GetBBox() vec2d.Rect {
	return getExtent(c.Coverages)
}

func (c *MultiCoverage) GetExtent() *MapExtent {
	return &MapExtent{BBox: c.GetBBox(), Srs: c.GetSrs()}
}

func (c *MultiCoverage) Intersects(bbox vec2d.Rect, srs Proj) bool {
	for i := range c.Coverages {
		if c.Coverages[i].Intersects(bbox, srs) {
			return true
		}
	}
	return false
}

func (c *MultiCoverage) Contains(bbox vec2d.Rect, srs Proj) bool {
	for i := range c.Coverages {
		if c.Coverages[i].Contains(bbox, srs) {
			return true
		}
	}
	return false
}

func (c *MultiCoverage) ContainsPoint(pt []float64, srs Proj) bool {
	for i := range c.Coverages {
		if c.Coverages[i].ContainsPoint(pt, srs) {
			return true
		}
	}
	return false
}

func (c *MultiCoverage) TransformTo(srs Proj) Coverage {
	ret := make([]Coverage, len(c.Coverages))
	for i := range c.Coverages {
		ret[i] = c.Coverages[i].TransformTo(srs)
	}
	return NewMultiCoverage(ret)
}

func (c *MultiCoverage) Equals(cc Coverage) bool {
	if cb, ok := cc.(*MultiCoverage); !ok {
		return false
	} else {
		if c.BBox.Min != cb.BBox.Min || c.BBox.Max != cb.BBox.Max {
			return false
		}

		if len(c.Coverages) != len(cb.Coverages) {
			return false
		}

		for i := range c.Coverages {
			if !c.Coverages[i].Equals(cb.Coverages[i]) {
				return false
			}
		}

		return true
	}
}

func geosBoundsToRect(geom *geos.Geometry) vec2d.Rect {
	minx, miny, maxx, maxy := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	if geom.GetType() == geos.LINESTRING || geom.GetType() == geos.POINT {
		points := geom.GetCoords()
		for _, p := range points {
			minx = math.Min(minx, p.X)
			maxx = math.Max(maxx, p.X)
			miny = math.Min(miny, p.Y)
			maxy = math.Max(maxy, p.Y)
		}
	} else if geom.GetType() == geos.POLYGON {
		points := geom.GetExteriorRing().GetCoords()
		for _, p := range points {
			minx = math.Min(minx, p.X)
			maxx = math.Max(maxx, p.X)
			miny = math.Min(miny, p.Y)
			maxy = math.Max(maxy, p.Y)
		}
	}
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

type GeomCoverage struct {
	Coverage
	BBox vec2d.Rect
	Srs  Proj
	Geom *geos.Geometry
	Clip bool
}

func NewGeomCoverage(geom geom.Geometry, srs Proj, clip bool) *GeomCoverage {
	geo := geos.ConvertGeomToGeos(geom)
	return NewGeosCoverage(geo, srs, clip)
}

func NewGeosCoverage(geom *geos.Geometry, srs Proj, clip bool) *GeomCoverage {
	bounds := geosBoundsToRect(geom)
	return &GeomCoverage{BBox: bounds, Srs: srs, Geom: geom, Clip: clip}
}

func NewBBoxCoverage(rect vec2d.Rect, srs Proj, clip bool) *GeomCoverage {
	geom := BBoxPolygon(rect)
	bounds := geosBoundsToRect(geom)
	return &GeomCoverage{BBox: bounds, Srs: srs, Geom: geom, Clip: clip}
}

func (c *GeomCoverage) IsClip() bool {
	return c.Clip
}

func (c *GeomCoverage) GetBBox() vec2d.Rect {
	return c.BBox
}

func (c *GeomCoverage) GetGeom() *geos.Geometry {
	return c.Geom
}

func (c *GeomCoverage) GetClip() bool {
	return c.Clip
}

func (c *GeomCoverage) GetSrs() Proj {
	return c.Srs
}

func (c *GeomCoverage) GetExtent() *MapExtent {
	return &MapExtent{BBox: c.BBox, Srs: c.Srs}
}

func (c *GeomCoverage) geomInCoverageSrs(gin interface{}, srs Proj) *geos.Geometry {
	var geom *geos.Geometry
	switch g := gin.(type) {
	case *geos.Geometry:
		if !srs.Eq(c.Srs) {
			geom = transformGeometry(srs, c.Srs, geom)
		}
	case vec2d.T:
		if !srs.Eq(c.Srs) {
			g = srs.TransformTo(c.Srs, []vec2d.T{g})[0]
		}
		geom = geos.CreatePoint(g[0], g[1])
	case vec2d.Rect:
		if !srs.Eq(c.Srs) {
			g = srs.TransformRectTo(c.Srs, g, 16)
		}
		geom = BBoxPolygon(g)
	}

	return geom
}

func (c *GeomCoverage) Intersects(bbox vec2d.Rect, srs Proj) bool {
	geom := c.geomInCoverageSrs(bbox, srs)
	return c.Geom.Intersects(geom)
}

func (c *GeomCoverage) Intersection(bbox vec2d.Rect, srs Proj) Coverage {
	geom := c.geomInCoverageSrs(bbox, srs)
	return &GeomCoverage{Geom: c.Geom.Intersection(geom), Srs: c.Srs, Clip: c.Clip}
}

func (c *GeomCoverage) Contains(bbox vec2d.Rect, srs Proj) bool {
	geom := c.geomInCoverageSrs(bbox, srs)
	return c.Geom.Contains(geom)
}

func (c *GeomCoverage) ContainsPoint(pt []float64, srs Proj) bool {
	geom := c.geomInCoverageSrs(vec2d.T{pt[0], pt[1]}, srs)
	return c.Geom.Contains(geom)
}

func (c *GeomCoverage) TransformTo(srs Proj) Coverage {
	if srs.Eq(c.Srs) {
		return c
	}

	bbox := c.Srs.TransformRectTo(srs, c.BBox, 16)
	return &GeomCoverage{BBox: bbox, Srs: c.Srs, Clip: c.Clip}
}

func (c *GeomCoverage) Equals(cc Coverage) bool {
	if cb, ok := cc.(*GeomCoverage); !ok {
		return false
	} else {
		if !c.Srs.Eq(cb.Srs) {
			return false
		}

		if c.BBox.Min != cb.BBox.Min || c.BBox.Max != cb.BBox.Max {
			return false
		}

		if !c.Geom.Equals(cb.Geom) {
			return false
		}
		return true
	}
}

func BBoxPolygon(bbox vec2d.Rect) *geos.Geometry {
	shell := []geos.Coord{
		{X: bbox.Min[0], Y: bbox.Min[1]},
		{X: bbox.Max[0], Y: bbox.Min[1]},
		{X: bbox.Max[0], Y: bbox.Max[1]},
		{X: bbox.Min[0], Y: bbox.Max[1]},
		{X: bbox.Min[0], Y: bbox.Min[1]},
	}

	return geos.CreatePolygon(shell)
}

func transformXY(from_srs Proj, to_srs Proj, xy []geos.Coord) []geos.Coord {
	xyv := make([]vec2d.T, len(xy))
	for i := range xy {
		xyv[i] = vec2d.T{xy[i].X, xy[i].Y}
	}
	result := from_srs.TransformTo(to_srs, xyv)
	resultC := make([]geos.Coord, len(result))
	for i := range result {
		resultC[i] = geos.Coord{X: result[i][0], Y: result[i][1]}
	}
	return resultC
}

func transformPolygon(transf func([]geos.Coord) []geos.Coord, polygon *geos.Geometry) *geos.Geometry {
	ext := transf(polygon.GetExteriorRing().GetCoords())
	ints := make([][]geos.Coord, 0)

	for i := 0; i < polygon.GetNumInteriorRings(); i++ {
		in := polygon.GetInteriorRingN(i)
		ints = append(ints, transf(in.GetCoords()))
	}

	return geos.CreatePolygon(ext, ints...)
}

func transformMultipolygon(transf func([]geos.Coord) []geos.Coord, multipolygon *geos.Geometry) *geos.Geometry {
	transformed_polygons := make([]*geos.Geometry, 0)
	for i := 0; i < multipolygon.GetNumGeometries(); i++ {
		transformed_polygons = append(transformed_polygons, transformPolygon(transf, multipolygon.GetGeometryN(i)))
	}
	return geos.CreateMultiGeometry(transformed_polygons, geos.MULTIPOLYGON)
}

func transformGeometry(from_srs Proj, to_srs Proj, geometry *geos.Geometry) *geos.Geometry {
	transf := func(coords []geos.Coord) []geos.Coord {
		return transformXY(from_srs, to_srs, coords)
	}
	var result *geos.Geometry
	if geometry.GetType() == geos.POLYGON {
		result = transformPolygon(transf, geometry)
	} else if geometry.GetType() == geos.MULTIPOLYGON {
		result = transformMultipolygon(transf, geometry)
	} else {
		return nil
	}

	return result
}

func UnionCoverage(coverages []Coverage, clip bool) Coverage {
	srs := coverages[0].GetSrs()

	for i := range coverages {
		coverages[i] = coverages[i].TransformTo(srs)
	}

	geoms := make([]*geos.Geometry, 0)

	for c := range coverages {
		if coverages[c].GetGeom() != nil {
			geoms = append(geoms, coverages[c].GetGeom())
		} else {
			geoms = append(geoms, BBoxPolygon(coverages[c].GetBBox()))
		}
	}

	union := geos.CreateEmptyPolygon()
	for _, g := range geoms {
		union = union.Union(g)
	}

	return NewGeosCoverage(union, srs, clip)
}

func DiffCoverage(coverages []Coverage, clip bool) Coverage {
	srs := coverages[0].GetSrs()

	for i := range coverages {
		coverages[i] = coverages[i].TransformTo(srs)
	}

	geoms := make([]*geos.Geometry, 0)

	for c := range coverages {
		if coverages[c].GetGeom() != nil {
			geoms = append(geoms, coverages[c].GetGeom())
		} else {
			geoms = append(geoms, BBoxPolygon(coverages[c].GetBBox()))
		}
	}

	sub := geos.CreateEmptyPolygon()
	for _, g := range geoms[1:] {
		sub = sub.Union(g)
	}

	diff := geoms[0].Difference(sub)

	if diff.IsEmpty() {
		panic("diff did not return any geometry")
	}

	return NewGeosCoverage(diff, srs, clip)
}

func IntersectionCoverage(coverages []Coverage, clip bool) Coverage {
	srs := coverages[0].GetSrs()

	for i := range coverages {
		coverages[i] = coverages[i].TransformTo(srs)
	}

	geoms := make([]*geos.Geometry, 0)

	for c := range coverages {
		if coverages[c].GetGeom() != nil {
			geoms = append(geoms, coverages[c].GetGeom())
		} else {
			geoms = append(geoms, BBoxPolygon(coverages[c].GetBBox()))
		}
	}
	intersection := geoms[0]
	for i := 1; i < len(geoms); i++ {
		b := geoms[i]
		intersection = intersection.Intersection(b)
	}

	if intersection.IsEmpty() {
		panic("intersection did not return any geometry")
	}

	return NewGeosCoverage(intersection, srs, clip)
}

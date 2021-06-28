package maths

import (
	"testing"

	"github.com/flywave/go-geos"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestGeomCoverage(t *testing.T) {
	geom := geos.CreateFromWKT("POLYGON((10 10, 10 50, -10 60, 10 80, 80 80, 80 10, 10 10))")
	srs := NewSRSProj4("EPSG:4326")
	srs900913 := NewSRSProj4("EPSG:900913")

	coverage := NewGeomCoverage(geom, srs, false)

	if !BBoxEquals(coverage.BBox, vec2d.Rect{Min: vec2d.T{-10, 10}, Max: vec2d.T{80, 80}}, 0.0001, 0.0001) {
		t.FailNow()
	}

	if !coverage.Contains(vec2d.Rect{Min: vec2d.T{15, 15}, Max: vec2d.T{20, 20}}, srs) {
		t.FailNow()
	}

	if !coverage.Contains(vec2d.Rect{Min: vec2d.T{15, 15}, Max: vec2d.T{80, 20}}, srs) {
		t.FailNow()
	}

	if coverage.Contains(vec2d.Rect{Min: vec2d.T{9, 10}, Max: vec2d.T{20, 20}}, srs) {
		t.FailNow()
	}

	if !coverage.Intersects(vec2d.Rect{Min: vec2d.T{15, 15}, Max: vec2d.T{20, 20}}, srs) {
		t.FailNow()
	}

	if !coverage.Intersects(vec2d.Rect{Min: vec2d.T{15, 15}, Max: vec2d.T{80, 20}}, srs) {
		t.FailNow()
	}

	if !coverage.Intersects(vec2d.Rect{Min: vec2d.T{9, 10}, Max: vec2d.T{20, 20}}, srs) {
		t.FailNow()
	}

	if !coverage.Intersects(vec2d.Rect{Min: vec2d.T{-30, 10}, Max: vec2d.T{-8, 70}}, srs) {
		t.FailNow()
	}

	if coverage.Intersects(vec2d.Rect{Min: vec2d.T{-30, 10}, Max: vec2d.T{-11, 70}}, srs) {
		t.FailNow()
	}

	if coverage.Intersects(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1000, 1000}}, srs900913) {
		t.FailNow()
	}

	if !coverage.Intersects(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1500000, 1500000}}, srs900913) {
		t.FailNow()
	}
}

func TestGeomCoverageEq(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	srs2 := NewSRSProj4("EPSG:31467")

	g1 := geos.CreateFromWKT("POLYGON((10 10, 10 50, -10 60, 10 80, 80 80, 80 10, 10 10))")
	g2 := geos.CreateFromWKT("POLYGON((10 10, 10 50, -10 60, 10 80, 80 80, 80 10, 10 10))")
	c1 := NewGeomCoverage(g1, srs, false)
	c2 := NewGeomCoverage(g2, srs, false)
	c3 := NewGeomCoverage(g2, srs2, false)
	if !c1.Equals(c2) {
		t.FailNow()
	}

	if c1.Equals(c3) {
		t.FailNow()
	}

	g3 := geos.CreateFromWKT("POLYGON((10.0 10, 10 50.0, -10 60, 10 80, 80 80, 80 10, 10 10))")
	c4 := NewGeomCoverage(g3, srs, false)

	if !c1.Equals(c4) {
		t.FailNow()
	}

	g4 := geos.CreateFromWKT("POLYGON((10 10, 10.1 50, -10 60, 10 80, 80 80, 80 10, 10 10))")
	c5 := NewGeomCoverage(g4, srs, false)

	if c1.Equals(c5) {
		t.FailNow()
	}
}

func TestBBOXCoverage(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	coverage := NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{-10, 10}, Max: vec2d.T{80, 80}}, srs, false)

	if !BBoxEquals(coverage.BBox, vec2d.Rect{Min: vec2d.T{-10, 10}, Max: vec2d.T{80, 80}}, 0.0001, 0.0001) {
		t.FailNow()
	}

}

func TestUnionCoverage(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	srs2 := NewSRSProj4("EPSG:3857")
	coverage1 := NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, srs, false)

	g2 := geos.CreateFromWKT("POLYGON((10 0, 20 0, 20 10, 10 10, 10 0))")
	coverage2 := NewGeomCoverage(g2, srs, false)

	g3 := geos.CreateFromWKT("POLYGON((-1000000 0, 0 0, 0 1000000, -1000000 1000000, -1000000 0))")
	coverage3 := NewGeomCoverage(g3, srs2, false)

	coverage := UnionCoverage([]Coverage{
		coverage1,
		coverage2,
		coverage3,
	}, false)

	if !BBoxEquals(coverage.GetBBox(), vec2d.Rect{Min: vec2d.T{-8.98315284, 0.0}, Max: vec2d.T{20.0, 10.0}}, 0.0001, 0.0001) {
		t.FailNow()
	}

}

func TestDiffCoverage(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	srs2 := NewSRSProj4("EPSG:3857")
	g1 := geos.CreateFromWKT("POLYGON((-10 0, 20 0, 20 10, -10 10, -10 0))")
	coverage1 := NewGeomCoverage(g1, srs, false)

	coverage2 := NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 2}, Max: vec2d.T{8, 8}}, srs, false)

	g3 := geos.CreateFromWKT("POLYGON((-1000000 0, 0 0, 0 1000000, -1000000 1000000, -1000000 0))")
	coverage3 := NewGeomCoverage(g3, srs2, false)

	coverage := DiffCoverage([]Coverage{coverage1, coverage2, coverage3}, false)

	if !BBoxEquals(coverage.GetBBox(), vec2d.Rect{Min: vec2d.T{-10, 0.0}, Max: vec2d.T{20.0, 10.0}}, 0.0001, 0.0001) {
		t.FailNow()
	}

}

func TestIntersectionCoverage(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	g1 := geos.CreateFromWKT("POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))")
	coverage1 := NewGeomCoverage(g1, srs, false)

	coverage2 := NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{5, 5}, Max: vec2d.T{15, 15}}, srs, false)

	coverage := IntersectionCoverage([]Coverage{coverage1, coverage2}, false)

	if !BBoxEquals(coverage.GetBBox(), vec2d.Rect{Min: vec2d.T{5.0, 5.0}, Max: vec2d.T{10.0, 10.0}}, 0.0001, 0.0001) {
		t.FailNow()
	}

}

func TestMultiCoverage(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	g1 := geos.CreateFromWKT("POLYGON((10 10, 10 50, -10 60, 10 80, 80 80, 80 10, 10 10))")

	coverage1 := NewGeomCoverage(g1, srs, false)
	coverage2 := NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{100, 0}, Max: vec2d.T{120, 20}}, srs, false)
	coverage := NewMultiCoverage([]Coverage{coverage1, coverage2})

	if !BBoxEquals(coverage.GetBBox(), vec2d.Rect{Min: vec2d.T{-10, 0}, Max: vec2d.T{120, 80}}, 0.0001, 0.0001) {
		t.FailNow()
	}
}

package geo

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MapExtent struct {
	BBox   vec2d.Rect
	Srs    Proj
	llbbox *vec2d.Rect
}

func MapExtentFromGrid(grid *TileGrid) *MapExtent {
	return &MapExtent{BBox: *grid.BBox, Srs: grid.Srs}
}

func (m *MapExtent) BBoxFor(to Proj) vec2d.Rect {
	if to != nil && to.Eq(m.Srs) {
		return m.BBox
	}
	return m.Srs.TransformRectTo(to, m.BBox, 16)
}

func (m *MapExtent) Eq(o *MapExtent) bool {
	if o.Srs != nil && m.Srs != nil {
		if !o.Srs.Eq(m.Srs) {
			return false
		}
	}

	if m.BBox.Min != o.BBox.Min || m.BBox.Max != o.BBox.Max {
		return false
	}

	return true
}

func (m *MapExtent) IsDefault() bool {
	return false
}

func (m *MapExtent) GetLLBBox() vec2d.Rect {
	if m.llbbox == nil {
		lb := m.Srs.TransformRectTo(NewSRSProj4("EPSG:4326"), m.BBox, 16)
		m.llbbox = &lb
	}
	return *m.llbbox
}

func (m *MapExtent) Add(o *MapExtent) *MapExtent {
	if o.IsDefault() {
		return m
	}
	if m.IsDefault() {
		return o
	}
	return &MapExtent{BBox: MergeBBox(m.GetLLBBox(), o.GetLLBBox()), Srs: NewSRSProj4("EPSG:4326")}
}

func (m *MapExtent) Contains(other *MapExtent) bool {
	if m.IsDefault() {
		return true
	}
	return BBoxContains(m.BBox, other.BBoxFor(m.Srs))
}

func (m *MapExtent) Intersects(other *MapExtent) bool {
	return BBoxIntersects(m.BBox, other.BBoxFor(m.Srs))
}

func (m *MapExtent) Intersection(other *MapExtent) *MapExtent {
	source := m.BBox
	sub := other.BBoxFor(m.Srs)

	return &MapExtent{BBox: vec2d.Rect{
		Min: vec2d.T{
			math.Max(source.Min[0], sub.Min[0]),
			math.Max(source.Min[1], sub.Min[1])},
		Max: vec2d.T{
			math.Min(source.Max[0], sub.Max[0]),
			math.Min(source.Max[1], sub.Max[1])}},
		Srs: m.Srs}
}

func (m *MapExtent) transform(srs Proj) *MapExtent {
	return &MapExtent{BBox: m.BBoxFor(srs), Srs: srs}
}

type DefaultMapExtent struct {
	MapExtent
}

func (m *DefaultMapExtent) IsDefault() bool {
	return false
}

func MapExtentFromDefault() *MapExtent {
	return &MapExtent{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Srs: NewSRSProj4("EPSG:4326")}
}

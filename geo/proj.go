package geo

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo/gcj02"
)

type Proj interface {
	TransformTo(o Proj, points []vec2d.T) []vec2d.T
	TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect
	Eq(o Proj) bool
	AlignBBox(t vec2d.Rect) vec2d.Rect
	GetSrsCode() string
	GetDef() string
	IsLatLong() bool
	ToString() string
	IsAxisOrderNE() bool
}

type GCJ02Proj struct {
	Proj
	Exact bool
}

func NewGCJ02Proj(exact bool) *GCJ02Proj {
	return &GCJ02Proj{Exact: exact}
}

func (p *GCJ02Proj) IsLatLong() bool {
	return true
}

func (p *GCJ02Proj) IsAxisOrderNE() bool {
	return true
}

func (p *GCJ02Proj) TransformTo(o Proj, points []vec2d.T) []vec2d.T {
	switch prj := o.(type) {
	case *GCJ02Proj:
		return points
	case *BD09Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.GCJ02toBD09(p[0], p[1])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *BD09MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.GCJ02toBDMC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *GCJ02MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.GCJ02toGCJ02MC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *SRSProj4:
		wpoints, wproj := p.transformToWGS84(points)
		return wproj.TransformTo(prj, wpoints)
	}
	return nil
}

func (p *GCJ02Proj) TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect {
	if p.Eq(o) {
		return rect
	}
	bbox := p.AlignBBox(rect)
	points := GenerateEnvelopePoints(bbox, withPoints)
	transf_pts := p.TransformTo(o, points)
	result := CalculateBBox(transf_pts)
	return result
}

func (p *GCJ02Proj) AlignBBox(t vec2d.Rect) vec2d.Rect {
	delta := 0.00000001
	minx, miny, maxx, maxy := t.Min[0], t.Min[1], t.Max[0], t.Max[1]
	if math.Abs(miny - -90.0) < 1e-6 {
		miny = -90.0 + delta
	}
	if math.Abs(maxy-90.0) < 1e-6 {
		maxy = 90.0 - delta
	}
	bbox := vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	return bbox
}

func (p *GCJ02Proj) transformFromWGS84(points []vec2d.T) []vec2d.T {
	ret := make([]vec2d.T, len(points))
	for i, p := range points {
		lat, lng := gcj02.WGS84toGCJ02(p[1], p[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret
}

func (p *GCJ02Proj) transformToWGS84(points []vec2d.T) ([]vec2d.T, *SRSProj4) {
	ret := make([]vec2d.T, len(points))
	for i, pt := range points {
		if p.Exact {
			lat, lng := gcj02.GCJ02toWGS84Exact(pt[1], pt[0])
			ret[i] = vec2d.T{lng, lat}
		} else {
			lat, lng := gcj02.GCJ02toWGS84(pt[1], pt[0])
			ret[i] = vec2d.T{lng, lat}
		}
	}
	return ret, NewSRSProj4("EPSG:4326")
}

func (p *GCJ02Proj) Eq(o Proj) bool {
	if _, ok := o.(*GCJ02Proj); ok {
		return true
	}
	return false
}

func (p *GCJ02Proj) GetSrsCode() string {
	return "EPSG:GCJ02"
}

func (p *GCJ02Proj) GetDef() string {
	return "EPSG:GCJ02"
}

func (p *GCJ02Proj) ToString() string {
	return "GCJ02"
}

type BD09Proj struct {
	Proj
	Exact bool
}

func NewBD09Proj(exact bool) *BD09Proj {
	return &BD09Proj{Exact: exact}
}

func (p *BD09Proj) TransformTo(o Proj, points []vec2d.T) []vec2d.T {
	switch prj := o.(type) {
	case *BD09Proj:
		return points
	case *GCJ02Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BD09toGCJ02(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *BD09MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BD09toBDMC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *GCJ02MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BD09toGCJ02MC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *SRSProj4:
		wpoints, wproj := p.transformToWGS84(points)
		return wproj.TransformTo(prj, wpoints)
	}
	return nil
}

func (p *BD09Proj) AlignBBox(t vec2d.Rect) vec2d.Rect {
	delta := 0.00000001
	minx, miny, maxx, maxy := t.Min[0], t.Min[1], t.Max[0], t.Max[1]
	if math.Abs(miny - -90.0) < 1e-6 {
		miny = -90.0 + delta
	}
	if math.Abs(maxy-90.0) < 1e-6 {
		maxy = 90.0 - delta
	}
	bbox := vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	return bbox
}

func (p *BD09Proj) TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect {
	if p.Eq(o) {
		return rect
	}
	bbox := p.AlignBBox(rect)
	points := GenerateEnvelopePoints(bbox, withPoints)
	transf_pts := p.TransformTo(o, points)
	result := CalculateBBox(transf_pts)
	return result
}

func (p *BD09Proj) transformFromWGS84(points []vec2d.T) []vec2d.T {
	ret := make([]vec2d.T, len(points))
	for i, p := range points {
		lat, lng := gcj02.WGS84toBD09(p[1], p[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret
}

func (p *BD09Proj) transformToWGS84(points []vec2d.T) ([]vec2d.T, *SRSProj4) {
	ret := make([]vec2d.T, len(points))
	for i, pt := range points {
		if p.Exact {
			lat, lng := gcj02.GCJ02toWGS84Exact(pt[1], pt[0])
			ret[i] = vec2d.T{lng, lat}
		} else {
			lat, lng := gcj02.BD09toWGS84(pt[1], pt[0])
			ret[i] = vec2d.T{lng, lat}
		}
	}
	return ret, NewSRSProj4("EPSG:4326")
}

func (p *BD09Proj) GetSrsCode() string {
	return "EPSG:BD09"
}

func (p *BD09Proj) GetDef() string {
	return "EPSG:BD09"
}

func (p *BD09Proj) ToString() string {
	return "BD09"
}

func (p *BD09Proj) IsLatLong() bool {
	return true
}

func (p *BD09Proj) IsAxisOrderNE() bool {
	return true
}

func (p *BD09Proj) Eq(o Proj) bool {
	if _, ok := o.(*BD09Proj); ok {
		return true
	}
	return false
}

type GCJ02MCProj struct {
	Proj
}

func (p *GCJ02MCProj) TransformTo(o Proj, points []vec2d.T) []vec2d.T {
	switch prj := o.(type) {
	case *GCJ02MCProj:
		return points
	case *BD09Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.GCJ02MCtoBD09(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *BD09MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.GCJ02MCtoBDMC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *GCJ02Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BDMCtoGCJ02(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *SRSProj4:
		wpoints, wproj := p.transformToWGS84(points)
		return wproj.TransformTo(prj, wpoints)
	}
	return nil
}

func (p *GCJ02MCProj) TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect {
	if p.Eq(o) {
		return rect
	}
	bbox := p.AlignBBox(rect)
	points := GenerateEnvelopePoints(bbox, withPoints)
	transf_pts := p.TransformTo(o, points)
	result := CalculateBBox(transf_pts)
	return result
}

func (p *GCJ02MCProj) transformFromWGS84(points []vec2d.T) []vec2d.T {
	ret := make([]vec2d.T, len(points))
	for i, p := range points {
		lat, lng := gcj02.WGS84toGCJ02MC(p[1], p[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret
}

func (p *GCJ02MCProj) transformToWGS84(points []vec2d.T) ([]vec2d.T, *SRSProj4) {
	ret := make([]vec2d.T, len(points))
	for i, pt := range points {
		lat, lng := gcj02.GCJ02MCtoWGS84(pt[1], pt[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret, NewSRSProj4("EPSG:4326")
}

func (p *GCJ02MCProj) Eq(o Proj) bool {
	if _, ok := o.(*GCJ02MCProj); ok {
		return true
	}
	return false
}

func (p *GCJ02MCProj) AlignBBox(t vec2d.Rect) vec2d.Rect {
	var delta float64
	minx, miny, maxx, maxy := t.Min[0], t.Min[1], t.Max[0], t.Max[1]
	if math.Abs(miny - -90.0) < 1e-6 {
		miny = -90.0 + delta
	}
	if math.Abs(maxy-90.0) < 1e-6 {
		maxy = 90.0 - delta
	}
	bbox := vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	return bbox
}

func (p *GCJ02MCProj) GetSrsCode() string {
	return "EPSG:GCJ02MC"
}

func (p *GCJ02MCProj) GetDef() string {
	return "EPSG:GCJ02MC"
}

func (p *GCJ02MCProj) IsLatLong() bool {
	return false
}

func (p *GCJ02MCProj) ToString() string {
	return "GCJ02MC"
}

func (p *GCJ02MCProj) IsAxisOrderNE() bool {
	return true
}

type BD09MCProj struct {
	Proj
}

func (p *BD09MCProj) TransformTo(o Proj, points []vec2d.T) []vec2d.T {
	switch prj := o.(type) {
	case *BD09MCProj:
		return points
	case *BD09Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BDMCtoBD09(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *GCJ02Proj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BDMCtoGCJ02(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *GCJ02MCProj:
		ret := make([]vec2d.T, len(points))
		for i, p := range points {
			lat, lng := gcj02.BDMCtoGCJ02MC(p[1], p[0])
			ret[i] = vec2d.T{lng, lat}
		}
		return ret
	case *SRSProj4:
		wpoints, wproj := p.transformToWGS84(points)
		return wproj.TransformTo(prj, wpoints)
	}
	return nil
}

func (p *BD09MCProj) TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect {
	if p.Eq(o) {
		return rect
	}
	bbox := p.AlignBBox(rect)
	points := GenerateEnvelopePoints(bbox, withPoints)
	transf_pts := p.TransformTo(o, points)
	result := CalculateBBox(transf_pts)
	return result
}

func (p *BD09MCProj) transformFromWGS84(points []vec2d.T) []vec2d.T {
	ret := make([]vec2d.T, len(points))
	for i, p := range points {
		lat, lng := gcj02.WGS84toBDMC(p[1], p[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret
}

func (p *BD09MCProj) transformToWGS84(points []vec2d.T) ([]vec2d.T, *SRSProj4) {
	ret := make([]vec2d.T, len(points))
	for i, pt := range points {
		lat, lng := gcj02.BDMCtoWGS84(pt[1], pt[0])
		ret[i] = vec2d.T{lng, lat}
	}
	return ret, NewSRSProj4("EPSG:4326")
}

func (p *BD09MCProj) Eq(o Proj) bool {
	if _, ok := o.(*BD09MCProj); ok {
		return true
	}
	return false
}

func (p *BD09MCProj) AlignBBox(t vec2d.Rect) vec2d.Rect {
	var delta float64
	minx, miny, maxx, maxy := t.Min[0], t.Min[1], t.Max[0], t.Max[1]
	if math.Abs(miny - -90.0) < 1e-6 {
		miny = -90.0 + delta
	}
	if math.Abs(maxy-90.0) < 1e-6 {
		maxy = 90.0 - delta
	}
	bbox := vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	return bbox
}

func (p *BD09MCProj) GetSrsCode() string {
	return "EPSG:BDMC"
}

func (p *BD09MCProj) GetDef() string {
	return "EPSG:BDMC"
}

func (p *BD09MCProj) IsLatLong() bool {
	return false
}

func (p *BD09MCProj) ToString() string {
	return "BD09MC"
}

func (p *BD09MCProj) IsAxisOrderNE() bool {
	return true
}

package geo

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/go-tileproxy/utils"

	"github.com/flywave/go-proj"
)

var (
	WEBMERCATOR_EPSG = []string{
		"EPSG:900913", "EPSG:3857",
		"EPSG:102100", "EPSG:102113",
	}
	AXIS_ORDER_NE = []string{"EPSG:4326", "EPSG:4258", "EPSG:31466", "EPSG:31467", "EPSG:31468"}
	AXIS_ORDER_EN = []string{"CRS:84", "EPSG:900913", "EPSG:25831", "EPSG:25832", "EPSG:25833"}
)

type SRS string

var (
	ProjInit = map[string]SRS{
		"EPSG:4326": "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs +over",
		"CRS:84":    "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs +over",
	}
	_private_4326 Proj = nil
)

func getCurrentDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

func init() {
	dir := getCurrentDir()
	proj.SetFinder([]string{filepath.Join(dir, "../proj_data")})
	for _, s := range WEBMERCATOR_EPSG {
		ProjInit[s] = "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +no_defs +over"
	}
}

type SRSProj4 struct {
	Proj
	SrsCode string
	proj    *proj.Proj
}

func GetEpsgNum(srsCode string) int {
	if strings.ContainsRune(srsCode, ':') {
		epscode, err := strconv.Atoi(strings.Split(srsCode, ":")[1])
		if err == nil {
			return epscode
		}
	}
	return -1
}

func newSRSProj4(srsCode string) *SRSProj4 {
	p := &SRSProj4{SrsCode: srsCode}
	var err error
	if srs, ok := ProjInit[srsCode]; ok {
		p.proj, err = proj.NewProj(string(srs))
		if err != nil {
			return nil
		}
	} else {
		epsg_num := GetEpsgNum(srsCode)
		if epsg_num < 0 {
			return nil
		}
		srs := fmt.Sprintf("+init=epsg:%d +no_defs", epsg_num)
		p.proj, err = proj.NewProj(srs)
		if err != nil {
			return nil
		}
	}
	return p
}

func (p *SRSProj4) TransformTo(o Proj, points []vec2d.T) []vec2d.T {
	if _private_4326 == nil {
		_private_4326 = newSRSProj4("EPSG:4326")
	}

	switch prj := o.(type) {
	case *GCJ02Proj:
		points = p.TransformTo(_private_4326, points)
		return prj.transformFromWGS84(points)
	case *BD09Proj:
		points = p.TransformTo(_private_4326, points)
		return prj.transformFromWGS84(points)
	case *BD09MCProj:
		points = p.TransformTo(_private_4326, points)
		return prj.transformFromWGS84(points)
	case *GCJ02MCProj:
		points = p.TransformTo(_private_4326, points)
		return prj.transformFromWGS84(points)
	case *SRSProj4:
		if p.Eq(prj) {
			return points
		}
		ret := make([]vec2d.T, len(points))
		for i, v := range points {
			if p.IsLatLong() {
				x, y, err := proj.Transform2(p.proj, prj.proj, v[0]*Deg2Rad, v[1]*Deg2Rad)
				if err != nil {
					continue
				}
				if prj.IsLatLong() {
					x, y = x*Rad2Deg, y*Rad2Deg
				}
				ret[i] = vec2d.T{x, y}
			} else {
				x, y, err := proj.Transform2(p.proj, prj.proj, v[0], v[1])
				if err != nil {
					continue
				}
				if prj.IsLatLong() {
					x, y = x*Rad2Deg, y*Rad2Deg
				}
				ret[i] = vec2d.T{x, y}
			}
		}
		return ret
	}
	return nil
}

func GenerateEnvelopePoints(bbox vec2d.Rect, n int) []vec2d.T {
	minx, miny, maxx, maxy := bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]
	if n <= 4 {
		n = 0
	} else {
		n = int(math.Ceil((float64(n) - 4) / 4.0))
	}

	width := maxx - minx
	height := maxy - miny

	minx, maxx = math.Min(minx, maxx), math.Max(minx, maxx)
	miny, maxy = math.Min(miny, maxy), math.Max(miny, maxy)

	n += 1
	xstep := width / float64(n)
	ystep := height / float64(n)
	result := make([]vec2d.T, 0)
	for i := 0; i < n+1; i++ {
		result = append(result, vec2d.T{minx + float64(i)*xstep, miny})
	}
	for i := 1; i < n; i++ {
		result = append(result, vec2d.T{maxx, miny + float64(i)*ystep})
	}
	for i := n; i > -1; i-- {
		result = append(result, vec2d.T{minx + float64(i)*xstep, maxy})
	}
	for i := n - 1; i > 0; i-- {
		result = append(result, vec2d.T{minx, miny + float64(i)*ystep})
	}
	return result
}

func CalculateBBox(points []vec2d.T) vec2d.Rect {
	minx, miny, maxx, maxy := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, p := range points {
		if p[0] != math.Inf(1) {
			minx = math.Min(minx, p[0])
			maxx = math.Max(maxx, p[0])
		}
		if p[1] != math.Inf(1) {
			miny = math.Min(miny, p[1])
			maxy = math.Max(maxy, p[1])
		}
	}
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

func MergeBBox(bbox1 vec2d.Rect, bbox2 vec2d.Rect) vec2d.Rect {
	minx := math.Min(bbox1.Min[0], bbox2.Min[0])
	miny := math.Min(bbox1.Min[1], bbox2.Min[1])
	maxx := math.Max(bbox1.Max[0], bbox2.Max[0])
	maxy := math.Max(bbox1.Max[1], bbox2.Max[1])
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

func BBoxEquals(src_bbox vec2d.Rect, dst_bbox vec2d.Rect, x_delta float64, y_delta float64) bool {
	if x_delta == math.Inf(1) {
		x_delta = math.Abs(src_bbox.Min[0]-src_bbox.Max[0]) / 1000000.0
	}
	if y_delta == math.Inf(1) {
		y_delta = x_delta
	}
	return (math.Abs(src_bbox.Min[0]-dst_bbox.Min[0]) < x_delta &&
		math.Abs(src_bbox.Min[1]-dst_bbox.Min[1]) < x_delta &&
		math.Abs(src_bbox.Max[0]-dst_bbox.Max[0]) < y_delta &&
		math.Abs(src_bbox.Max[1]-dst_bbox.Max[1]) < y_delta)
}

func (p *SRSProj4) TransformRectTo(o Proj, rect vec2d.Rect, withPoints int) vec2d.Rect {
	if p.Eq(o) {
		return rect
	}
	bbox := p.AlignBBox(rect)
	points := GenerateEnvelopePoints(bbox, withPoints)
	transf_pts := p.TransformTo(o, points)
	result := CalculateBBox(transf_pts)

	if osrs, ok := o.(*SRSProj4); p.SrsCode == "EPSG:4326" && ok && (osrs.SrsCode == "EPSG:3857" || osrs.SrsCode == "EPSG:900913") {
		minx, miny, maxx, maxy := result.Min[0], result.Min[1], result.Max[0], result.Max[1]
		if bbox.Min[0] <= -180.0 {
			minx = -20037508.342789244
		}
		if bbox.Min[1] <= -85.06 {
			miny = -20037508.342789244
		}
		if bbox.Max[0] >= 180.0 {
			maxx = 20037508.342789244
		}
		if bbox.Max[1] >= 85.06 {
			maxy = 20037508.
		}
		result = vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	}

	return result
}

func (p *SRSProj4) AlignBBox(t vec2d.Rect) vec2d.Rect {
	var delta float64
	if p.SrsCode == "EPSG:4326" {
		delta = 0.00000001
	}
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

func (p *SRSProj4) IsLatLong() bool {
	return p.proj.IsLatLong()
}

func (p *SRSProj4) IsAxisOrderNE() bool {
	if utils.ContainsString(AXIS_ORDER_NE, p.SrsCode) {
		return true
	}
	if utils.ContainsString(AXIS_ORDER_EN, p.SrsCode) {
		return false
	}
	if p.IsLatLong() {
		return true
	}
	return false
}

func (p *SRSProj4) Eq(o Proj) bool {
	if pj, ok := o.(*SRSProj4); ok {
		return p.proj.GetDef() == pj.proj.GetDef()
	}
	return false
}

func (p *SRSProj4) GetSrsCode() string {
	return p.SrsCode
}

func (p *SRSProj4) GetDef() string {
	return p.proj.GetDef()
}

func (p *SRSProj4) ToString() string {
	return fmt.Sprintf("SRS %s ('%s')", p.SrsCode, p.proj.GetDef())
}

func MakeLinTransf(src_bbox, dst_bbox vec2d.Rect) func([]float64) []float64 {
	f := func(x_y []float64) []float64 {
		return []float64{dst_bbox.Min[0] + (x_y[0]-src_bbox.Min[0])*
			(dst_bbox.Max[0]-dst_bbox.Min[0])/(src_bbox.Max[0]-src_bbox.Min[0]),
			dst_bbox.Min[1] + (src_bbox.Max[1]-x_y[1])*
				(dst_bbox.Max[1]-dst_bbox.Min[1])/(src_bbox.Max[1]-src_bbox.Min[1])}
	}
	return f
}

type PreferredSrcSRS map[string][]Proj

func (m PreferredSrcSRS) Add(target string, prefered_srs []Proj) {
	m[target] = prefered_srs
}

func ContainsSrs(target string, srcs []Proj) (bool, Proj) {
	for _, p := range srcs {
		if p.GetSrsCode() == target {
			return true, p
		}
	}
	return false, nil
}

func (m PreferredSrcSRS) PreferredSrc(target Proj, available_src []Proj) (Proj, error) {
	if available_src == nil {
		return nil, errors.New("no available src srs")
	}

	if ok, p := ContainsSrs(target.GetSrsCode(), available_src); ok {
		return p, nil
	}

	if tv, ok := m[target.GetSrsCode()]; ok {
		for _, preferred := range tv {
			if ok, p := ContainsSrs(preferred.GetSrsCode(), available_src); ok {
				return p, nil
			}
		}
	}

	for _, avail := range available_src {
		if avail.IsLatLong() == target.IsLatLong() {
			return avail, nil
		}
	}
	return available_src[0], nil
}

type SupportedSRS struct {
	Srs       []Proj
	Preferred PreferredSrcSRS
}

func (s *SupportedSRS) BestSrs(target Proj) (Proj, error) {
	return s.Preferred.PreferredSrc(target, s.Srs)
}

func (s *SupportedSRS) Eq(o *SupportedSRS) bool {
	if len(s.Srs) != len(o.Srs) {
		return false
	}
	for _, t := range s.Srs {
		if ok, _ := ContainsSrs(t.GetDef(), o.Srs); !ok {
			return false
		}
	}
	return true
}

type GeoReference struct {
	bbox vec2d.Rect
	srs  Proj
}

func NewGeoReference(bbox vec2d.Rect, srs Proj) *GeoReference {
	return &GeoReference{bbox: bbox, srs: srs}
}

func (g *GeoReference) GetOrigin() vec2d.T {
	return g.bbox.Min
}

func (g *GeoReference) GetBBox() vec2d.Rect {
	return g.bbox
}

func (g *GeoReference) GetSrs() Proj {
	return g.srs
}

func (g *GeoReference) TilePoints() [6]float64 {
	return [6]float64{
		0.0, 0.0, 0.0,
		g.bbox.Min[0], g.bbox.Max[1], 0.0,
	}
}

func (g *GeoReference) PixelScale(img_size [2]int) vec3d.T {
	width := g.bbox.Max[0] - g.bbox.Min[0]
	height := g.bbox.Max[1] - g.bbox.Min[1]
	return vec3d.T{width / float64(img_size[0]), height / float64(img_size[1]), 0.0}
}

func SrcInProj(srscode string, projs []Proj) bool {
	for i := range projs {
		if projs[i].GetSrsCode() == srscode {
			return true
		}
	}
	return false
}

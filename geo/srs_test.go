package geo

import (
	"fmt"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-proj"
)

func TestSrs4326(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")
	//dst := NewSRSProj4("EPSG:900913")

	if !srs.IsLatLong() {
		t.FailNow()
	}

	srsStr := srs.GetDef()

	if len(srsStr) == 0 {
		t.FailNow()
	}

	t.Log(srsStr)

	//pts := srs.TransformTo(dst, []vec2d.T{{8.3, 53.2}})

	srs4326, _ := proj.NewProj("+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs ")
	srs900913, _ := proj.NewProj("+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs")
	x, y, _ := proj.Transform2(srs4326, srs900913, 121.5272106*Deg2Rad, 31.1774276*Deg2Rad)

	t.Log(fmt.Sprintf("%v %v", x, y))

	x = 13528347.201518921
	y = 3655812.9954997827

	lx, ly, _ := proj.Transform2(srs900913, srs4326, x, y)

	t.Log(fmt.Sprintf("%v %v", lx, ly))
}

func TestSrs(t *testing.T) {
	srs := NewSRSProj4("EPSG:900913")

	if srs.IsLatLong() {
		t.FailNow()
	}

	srsStr := srs.GetDef()

	if len(srsStr) == 0 {
		t.FailNow()
	}

	t.Log(srsStr)
}

func TestSrs26592(t *testing.T) {
	srs := NewSRSProj4("EPSG:26592")

	if srs.IsLatLong() {
		t.FailNow()
	}

	srsStr := srs.GetDef()

	if len(srsStr) == 0 {
		t.FailNow()
	}

	t.Log(srsStr)
}

func TestSrsTranRect(t *testing.T) {
	src := NewSRSProj4("EPSG:4326")
	dest := NewSRSProj4("EPSG:3857")

	rect := src.TransformRectTo(dest, vec2d.Rect{Min: vec2d.T{-180.0, -90.0}, Max: vec2d.T{180.0, 90.0}}, 16)
	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", rect.Min[0], rect.Min[1], rect.Max[0], rect.Max[1]))

	rect = src.TransformRectTo(dest, vec2d.Rect{Min: vec2d.T{53.1, 8.2}, Max: vec2d.T{53.2, 8.3}}, 16)
	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", rect.Min[0], rect.Min[1], rect.Max[0], rect.Max[1]))
}

func TestGenerateEnvelopePoints(t *testing.T) {
	pts := GenerateEnvelopePoints(vec2d.Rect{Min: vec2d.T{10.0, 5.0}, Max: vec2d.T{20.0, 15.0}}, 8)
	for i := range pts {
		t.Log(fmt.Sprintf("%v--%v/n", pts[i][0], pts[i][1]))
	}
}

func TestBBox(t *testing.T) {
	rect := CalculateBBox([]vec2d.T{{-5, 20}, {3, 8}, {99, 0}})
	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", rect.Min[0], rect.Min[1], rect.Max[0], rect.Max[1]))

	rect = MergeBBox(vec2d.Rect{Min: vec2d.T{-10, 20}, Max: vec2d.T{0, 30}}, vec2d.Rect{Min: vec2d.T{30, -20}, Max: vec2d.T{90, 10}})

	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", rect.Min[0], rect.Min[1], rect.Max[0], rect.Max[1]))

	src_bbox := vec2d.Rect{Min: vec2d.T{939258.20356824622, 6887893.4928338043},
		Max: vec2d.T{1095801.2374962866, 7044436.5267618448}}
	dst_bbox := vec2d.Rect{Min: vec2d.T{939258.20260000182, 6887893.4908000007},
		Max: vec2d.T{1095801.2365000017, 7044436.5247000009}}

	if !BBoxEquals(src_bbox, dst_bbox, 61.1, 61.1) {
		t.FailNow()
	}

	if BBoxEquals(src_bbox, dst_bbox, 0.0001, 0.0001) {
		t.FailNow()
	}
}

func TestMakeLinTransf(t *testing.T) {
	transf := MakeLinTransf(vec2d.Rect{Min: vec2d.T{7, 50},
		Max: vec2d.T{8, 51}},
		vec2d.Rect{Min: vec2d.T{0, 0},
			Max: vec2d.T{500, 400}})

	pt := transf([]float64{7.5, 50.5})

	t.Log(fmt.Sprintf("%v--%v/n", pt[0], pt[1]))

	pt = transf([]float64{7.0, 50.0})

	t.Log(fmt.Sprintf("%v--%v/n", pt[0], pt[1]))
}

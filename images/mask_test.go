package images

import (
	"os"
	"testing"

	"github.com/flywave/go-geos"
	"github.com/flywave/go-tileproxy/maths"
	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/imaging"
	"github.com/fogleman/gg"
)

func TestMaskImage(t *testing.T) {
	geom := "POLYGON((2 2, 2 8, 8 8, 8 2, 2 2), (4 4, 4 6, 6 6, 6 4, 4 4))"
	img, _ := imaging.Open("./flowers.png")
	img, mask := maskImage(img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, maths.NewSRSProj4("EPSG:4326"), maths.NewGeomCoverage(geos.CreateFromWKT(geom), maths.NewSRSProj4("EPSG:4326"), false))

	gc := gg.NewContext(600, 400)
	gc.SetMask(mask)
	gc.DrawImage(img, 0, 0)

	gc.SavePNG("./test.png")

	defer os.Remove("./test.png")
}

func coverage(geom *geos.Geometry, srs string) *maths.GeomCoverage {
	return maths.NewGeomCoverage(geom, maths.NewSRSProj4(srs), false)
}

func TestMaskOutsideOfImageTransparent(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = newBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, maths.NewSRSProj4("EPSG:4326"), maths.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{20, 20}, Max: vec2d.T{30, 30}}, maths.NewSRSProj4("EPSG:4326"), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}
}

func TestWKTMask(t *testing.T) {
	geom := "POLYGON((2 2, 2 8, 8 8, 8 2, 2 2), (4 4, 4 6, 6 6, 6 4, 4 4))"

	img_opts := *PNG_FORMAT
	img_opts.Transparent = newBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, maths.NewSRSProj4("EPSG:4326"), maths.NewGeomCoverage(geos.CreateFromWKT(geom), maths.NewSRSProj4("EPSG:4326"), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}

	imaging.Save(result.GetImage(), "./test.png")
}

func TestGeosMask(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = newBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	geom := geos.CreatePolygon([]geos.Coord{{X: 0, Y: 0}, {X: 222000, Y: 0}, {X: 222000, Y: 222000}, {X: 0, Y: 222000}, {X: 0, Y: 0}})

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, maths.NewSRSProj4("EPSG:4326"), maths.NewGeomCoverage(geom, maths.NewSRSProj4("EPSG:4326"), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}
}

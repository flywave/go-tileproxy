package imagery

import (
	"image"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"

	"github.com/flywave/go-geos"

	"github.com/flywave/gg"
	"github.com/flywave/imaging"
)

func TestMaskImage(t *testing.T) {
	geom := "POLYGON((2 2, 2 8, 8 8, 8 2, 2 2), (4 4, 4 6, 6 6, 6 4, 4 4))"
	img, _ := imaging.Open("../data/flowers.png")
	img, mask := maskImage(img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), geo.NewGeosCoverage(geos.CreateFromWKT(geom), geo.NewProj(4326), false))

	gc := gg.NewContext(600, 400)
	gc.SetMask(mask)
	gc.DrawImage(img, 0, 0)

	gc.SavePNG("./test.png")

	defer os.Remove("./test.png")
}

func TestMaskOutsideOfImageTransparent(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), geo.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{20, 20}, Max: vec2d.T{30, 30}}, geo.NewProj(4326), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}
}

func TestWKTMask(t *testing.T) {
	geom := "POLYGON((2 2, 2 8, 8 8, 8 2, 2 2), (4 4, 4 6, 6 6, 6 4, 4 4))"

	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), geo.NewGeosCoverage(geos.CreateFromWKT(geom), geo.NewProj(4326), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}

	imaging.Save(result.GetTile().(image.Image), "./test.png")
}

func TestGeosMask(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	geom := geos.CreatePolygon([]geos.Coord{{X: 0, Y: 0}, {X: 222000, Y: 0}, {X: 222000, Y: 222000}, {X: 0, Y: 222000}, {X: 0, Y: 0}})

	result := MaskImageSourceFromCoverage(
		img, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), geo.NewGeosCoverage(geom, geo.NewProj(4326), false), &img_opts)

	if result.GetSize()[0] != 100 || result.GetSize()[1] != 100 {
		t.FailNow()
	}
}

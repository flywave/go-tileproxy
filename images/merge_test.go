package images

import (
	"image"
	"image/color"
	"testing"

	"github.com/flywave/go-geos"
	"github.com/flywave/go-tileproxy/maths"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestMergeSingleCoverage(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = newBool(true)
	img := CreateImageSource([2]uint32{100, 100}, &img_opts)

	nimg := img.GetImage().(*image.NRGBA)

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			nimg.Set(x, y, color.NRGBA{128, 0, 128, 255})
		}
	}

	geom := geos.CreatePolygon([]geos.Coord{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}})

	coverage1 := maths.NewGeomCoverage(geom, maths.NewSRSProj4("EPSG:3857"), true)

	merger := &LayerMerger{}
	merger.Add(img, coverage1)

	result := merger.Merge(&img_opts, nil, vec2d.Rect{Min: vec2d.T{5, 0}, Max: vec2d.T{15, 10}}, maths.NewSRSProj4("EPSG:3857"), nil)

	ri := result.GetImage()

	c := ri.At(6, 0)
	_, _, _, A := c.RGBA()
	if A != 0 {
		t.FailNow()
	}
}

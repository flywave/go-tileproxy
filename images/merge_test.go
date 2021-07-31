package images

import (
	"image"
	"image/color"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"

	"github.com/flywave/go-geos"
)

func TestMergeSingleCoverage(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img_opts.BgColor = color.Transparent
	img := CreateImageSource([2]uint32{10, 10}, &img_opts)

	nimg := img.GetImage().(*image.NRGBA)

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			nimg.Set(x, y, color.NRGBA{128, 128, 255, 255})
		}
	}

	geom := geos.CreatePolygon([]geos.Coord{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}})

	coverage1 := geo.NewGeomCoverage(geom, geo.NewSRSProj4("EPSG:3857"), true)

	merger := &LayerMerger{}
	merger.AddSource(img, coverage1)

	result := merger.Merge(&img_opts, nil, vec2d.Rect{Min: vec2d.T{5, 0}, Max: vec2d.T{15, 10}}, geo.NewSRSProj4("EPSG:3857"), nil)

	ri := result.GetTile().(image.Image)
	c := ri.At(6, 0)
	_, _, _, A := c.RGBA()
	if A != 0 {
		t.FailNow()
	}
}

package images

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/fogleman/gg"

	"github.com/flywave/go-tileproxy/geo"

	"github.com/flywave/imaging"
)

func TestImageTransform(t *testing.T) {
	img, _ := imaging.Open("../data/flowers.png")
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(false)

	src_img := CreateImageSourceFromImage(img, &img_opts)
	src_srs := geo.NewSRSProj4("EPSG:31467")
	dst_size := [2]uint32{100, 150}
	dst_srs := geo.NewSRSProj4("EPSG:4326")
	dst_bbox := vec2d.Rect{Min: vec2d.T{0.2, 45.1}, Max: vec2d.T{8.3, 53.2}}
	src_bbox := dst_srs.TransformRectTo(src_srs, dst_bbox, 16)

	transformer := &ImageTransformer{SrcSRS: src_srs, DstSRS: dst_srs}

	img_opts = *PNG_FORMAT
	img_opts.Resampling = "nearest"

	result := transformer.Transform(
		src_img,
		src_bbox,
		dst_size,
		dst_bbox,
		&img_opts)

	imaging.Save(result.GetTile().(image.Image), "./transform.png")
	errs := []float64{0.2, 0.5, 1, 2, 4, 6, 8, 12, 16}
	for _, err := range errs {
		transformer := &ImageTransformer{SrcSRS: src_srs, DstSRS: dst_srs, MaxPxErr: err}
		result = transformer.Transform(src_img, src_bbox, dst_size, dst_bbox, &img_opts)
		imaging.Save(result.GetTile().(image.Image), fmt.Sprintf("./transform_%d.png", int(err*10)))
	}
}

func TestMergeTransform(t *testing.T) {
	pgcj02 := geo.NewGCJ02Proj(true)
	srs900913 := geo.NewSRSProj4("EPSG:900913")
	srs4326 := geo.NewSRSProj4("EPSG:4326")

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox1 := grid.TileBBox([3]int{53958, 24829, 16}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox1, 16)
	bbox := srs4326.TransformRectTo(pgcj02, bbox2, 16)

	bbox3 := srs4326.TransformRectTo(srs900913, bbox, 16)

	rect, grids, tiles, _ := grid.GetAffectedTiles(bbox, [2]uint32{256, 256}, srs4326)

	tilesCoord := [][3]int{}
	minx, miny := 0, 0
	for {
		x, y, z, done := tiles.Next()

		if minx == 0 || x < minx {
			minx = x
		}

		if miny == 0 || y < miny {
			miny = y
		}

		tilesCoord = append(tilesCoord, [3]int{x, y, z})

		if done {
			break
		}
	}

	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img_opts.BgColor = color.Transparent

	result := CreateImage([2]uint32{uint32(grids[0] * 256), uint32(grids[1] * 256)}, &img_opts)

	dc := gg.NewContextForImage(result)

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		source := CreateImageSource([2]uint32{256, 256}, &img_opts)

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.png", z, x, y))

		img := source.GetImage()

		dc.DrawImage(img, (x-minx)*256, (y-miny)*256)
	}

	img := dc.Image()

	facx := (bbox3.Min[0] - rect.Min[0]) / (rect.Max[0] - rect.Min[0])
	facy := (bbox3.Min[1] - rect.Min[1]) / (rect.Max[1] - rect.Min[1])

	offx := int(facx * 512)
	offy := int(facy * 512)

	subimg := imaging.Crop(img, image.Rect(offx, offy, offx+256, offy+256))

	imaging.Save(subimg, "./result.png")

	if result != nil || rect.Max[0] == 0 {
		t.FailNow()
	}
}

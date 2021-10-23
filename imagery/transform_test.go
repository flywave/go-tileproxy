package imagery

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"

	"github.com/flywave/imaging"
)

func TestImageTransform(t *testing.T) {
	img, _ := imaging.Open("../data/flowers.png")
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(false)

	src_img := CreateImageSourceFromImage(img, &img_opts)
	src_srs := geo.NewProj(31467)
	dst_size := [2]uint32{100, 150}
	dst_srs := geo.NewProj(4326)
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
	pgcj02 := geo.NewProj("EPSG:GCJ02")
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox1 := grid.TileBBox([3]int{53958, 24829, 16}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox1, 16)
	bbox := srs4326.TransformRectTo(pgcj02, bbox2, 16)

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
	var sources []tile.Source

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		source := CreateImageSource([2]uint32{256, 256}, &img_opts)

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.png", z, x, y))

		sources = append(sources, source)
	}

	m := NewTileMerger(grids, [2]uint32{256, 256})
	rr := m.Merge(sources, PNG_FORMAT)
	img := rr.GetTile().(image.Image)

	imaging.Save(img, "./all.png")

	os.Remove("./all.png")

	sp := NewTileSplitter(rr, PNG_FORMAT)

	off := imageTileOffset(rect, srs900913, [2]uint32{uint32(grids[0] * 256), uint32(grids[1] * 256)}, bbox, srs4326)

	subimg := sp.GetTile(off, [2]uint32{256, 256}).GetImage()

	imaging.Save(subimg, "./result.png")

	os.Remove("./result.png")

	if result != nil || rect.Max[0] == 0 {
		t.FailNow()
	}
}

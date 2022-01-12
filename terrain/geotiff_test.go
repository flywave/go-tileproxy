package terrain

import (
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestGetGeotiff(t *testing.T) {
	bbox := vec2d.Rect{
		Min: vec2d.T{117.8265, 36.832349},
		Max: vec2d.T{117.8788, 36.842198},
	}

	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	r, _, _ := grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{512, 512}, srs4326)

	cbox, grids, it, err := grid.GetAffectedLevelTiles(r, 13)

	sbox := srs900913.TransformRectTo(srs4326, cbox, 16)

	if err != nil || sbox.Min[0] == 0 {
		t.FailNow()
	}

	tilesCoord := [][3]int{}
	minx, miny := 0, 0
	for {
		x, y, z, done := it.Next()

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

	if len(tilesCoord) == 0 {
		t.FailNow()
	}

	var sources []tile.Source
	opts := &RasterOptions{Format: tile.TileFormat("webp"), Mode: BORDER_BILATERAL}

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		source := NewDemRasterSource(ModeMapbox, opts)

		b := grid.TileBBox([3]int{x, y, z}, false)

		source.georef = geo.NewGeoReference(b, srs900913)

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.webp", z, x, y))

		sources = append(sources, source)
	}

	m := NewRasterMerger(grids, [2]uint32{512, 512})
	rr := m.Merge(sources, opts)

	tiledata := rr.GetTile().(*TileData)

	io := &DemIO{Mode: ModeMapbox, Format: tile.TileFormat("webp")}

	raw, _ := io.Encode(tiledata)

	f, _ := os.Create("./zhoucun.webp")
	f.Write(raw)
	f.Close()

	rect := image.Rect(0, 0, int(tiledata.Size[0]), int(tiledata.Size[1]))

	src := cog.NewSource(tiledata.Datas, &rect, cog.CTLZW)

	cog.WriteTile("./zhoucun.tif", src, sbox, srs4326, tiledata.Size, nil)
}

func TestGeoTIFF(t *testing.T) {
	f, _ := os.Open("../data/14_13733_6366.webp")
	data, _ := LoadDEM(f, ModeMapbox)
	f.Close()

	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_BILATERAL)
	for x := 0; x < data.Dim; x++ {
		for y := 0; y < data.Dim; y++ {
			if x > 0 && y > 0 && x < data.Dim-1 && y < data.Dim-1 {
				tiledata.Set(x-1, y-1, data.Get(x, y))
			}

			if x == 0 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_LEFT, y-1, data.Get(x, y))
			}

			if x == data.Dim-1 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_RIGHT, y-1, data.Get(x, y))
			}

			if y == 0 {
				tiledata.FillBorder(BORDER_TOP, x, data.Get(x, y))
			}

			if y == data.Dim-1 {
				tiledata.FillBorder(BORDER_BOTTOM, x, data.Get(x, y))
			}
		}
	}

	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{514, 514}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)

	tiledata.Box = bbox2
	tiledata.Boxsrs = srs4326

	tiffio := &GeoTIFFIO{Mode: BORDER_BILATERAL}

	tiff, _ := tiffio.Encode(tiledata)

	f, _ = os.Create("./data.tiff")
	f.Write(tiff)
	f.Close()

	tiffio2 := &GeoTIFFIO{Mode: BORDER_BILATERAL}

	f, _ = os.Open("./data.tiff")
	tiledata2, _ := tiffio2.Decode(f)

	if tiledata2 == nil {
		t.FailNow()
	}
}

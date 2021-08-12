package raster

import (
	"os"
	"testing"

	"github.com/flywave/go-tileproxy/geo"
)

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

	srs900913 := geo.NewSRSProj4("EPSG:900913")
	srs4326 := geo.NewSRSProj4("EPSG:4326")

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

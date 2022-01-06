package terrain

import (
	"fmt"
	"os"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
)

func TestDem(t *testing.T) {
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

	io := &DemIO{Mode: ModeMapbox, Format: tile.TileFormat("webp")}

	raw, _ := io.Encode(tiledata)

	f, _ = os.Create("./data.webp")
	f.Write(raw)
	f.Close()
}

func convertGeoTIFF(x, y, z int) {
	webp := fmt.Sprintf("../data/%d_%d_%d.webp", z, x, y)
	f, _ := os.Open(webp)
	data, _ := LoadDEM(f, ModeMapbox)
	f.Close()

	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_NONE)
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
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{x, y, z}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)

	tiledata.Box = bbox2
	tiledata.Boxsrs = srs4326

	tiffio := &GeoTIFFIO{Mode: BORDER_NONE}

	tiff, _ := tiffio.Encode(tiledata)

	geotiff := fmt.Sprintf("../data/%d_%d_%d.tif", z, x, y)

	f, _ = os.Create(geotiff)
	f.Write(tiff)
	f.Close()
}

func TestTiff(t *testing.T) {
	tiles := [][3]int{
		{13733, 6366, 14},
		{13733, 6367, 14},
		{13734, 6366, 14},
		{13734, 6367, 14},
	}
	tiffs := []string{}
	for i := range tiles {
		convertGeoTIFF(tiles[i][0], tiles[i][1], tiles[i][2])
		tiffs = append(tiffs, fmt.Sprintf("../data/%d_%d_%d.tif", tiles[i][2], tiles[i][0], tiles[i][1]))
	}

}

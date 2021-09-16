package terrain

import (
	"os"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
)

func TestTerrainSource(t *testing.T) {
	opts := &RasterOptions{Format: tile.TileFormat("terrain"), Mode: BORDER_BILATERAL}

	source := NewTerrainSource(opts)

	source.SetSource("../data/323.terrain")
	t1 := source.GetTile()

	if t1 == nil {
		t.FailNow()
	}
}

func TestGenTerrainSourceFromDem(t *testing.T) {
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

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)

	tiledata.Box = bbox
	tiledata.Boxsrs = srs900913

	opts := &RasterOptions{Format: tile.TileFormat("terrain"), MaxError: 2}

	source, err := GenTerrainSource(tiledata, opts)

	t1 := source.GetTile()

	buff := source.GetBuffer(nil, nil)

	f, _ = os.Create("./data.terrain")
	f.Write(buff)
	f.Close()

	if t1 == nil && err != nil {
		t.FailNow()
	}
}

func TestGenTerrainSourceFromLerc(t *testing.T) {
	lercio := &LercIO{Mode: BORDER_UNILATERAL}

	f, _ := os.Open("../data/title_13_3252_6773.atm")
	tiledata, _ := lercio.Decode(f)
	f.Close()

	if tiledata == nil {
		t.FailNow()
	}

	srs900913 := geo.NewProj(900913)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{3252, 6773, 13}, false)

	tiledata.Box = bbox
	tiledata.Boxsrs = srs900913

	opts := &RasterOptions{Format: tile.TileFormat("terrain"), MaxError: 2}

	source, err := GenTerrainSource(tiledata, opts)

	t1 := source.GetTile()

	buff := source.GetBuffer(nil, nil)

	f, _ = os.Create("./data.terrain")
	f.Write(buff)
	f.Close()

	if t1 == nil && err != nil {
		t.FailNow()
	}
}

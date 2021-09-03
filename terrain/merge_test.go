package terrain

import (
	"fmt"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

func TestRasterMerger(t *testing.T) {
	pgcj02 := geo.NewProj("EPSG:GCJ02")
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox1 := grid.TileBBox([3]int{13733, 6367, 14}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox1, 16)
	bbox := srs4326.TransformRectTo(pgcj02, bbox2, 16)

	rect, grids, tiles, _ := grid.GetAffectedTiles(bbox, [2]uint32{512, 512}, srs4326)

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

	if rr != nil && rect.Min[0] == 0 {
		t.FailNow()
	}

	tiledata := rr.GetTile().(*TileData)

	io := &DemIO{Mode: ModeMapbox, Format: tile.TileFormat("webp")}

	raw, _ := io.Encode(tiledata)

	f, _ := os.Create("./data.webp")
	f.Write(raw)
	f.Close()

	os.Remove("./data.webp")

	rminx := rect.Min[0] + ((rect.Max[0] - rect.Min[0]) * 0.25)
	rminy := rect.Min[1] + ((rect.Max[1] - rect.Min[1]) * 0.25)

	rmaxx := rect.Min[0] + ((rect.Max[0] - rect.Min[0]) * 0.75)
	rmaxy := rect.Min[1] + ((rect.Max[1] - rect.Min[1]) * 0.75)

	newbox := vec2d.Rect{Min: vec2d.T{rminx, rminy}, Max: vec2d.T{rmaxx, rmaxy}}

	georef := geo.NewGeoReference(newbox, srs900913)

	Grid := CaclulateGrid(512, 512, BORDER_BILATERAL, georef)

	rsource := rr.(*DemRasterSource)

	rsource.Resample(nil, Grid)

	smtd := Grid.GetTileDate(BORDER_BILATERAL)

	raw, _ = io.Encode(smtd)

	f, _ = os.Create("./smtd.webp")
	f.Write(raw)
	f.Close()

	os.Remove("./smtd.webp")
}

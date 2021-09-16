package terrain

import (
	"testing"

	"github.com/flywave/go-geo"
)

func TestGrid(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)

	georef := geo.NewGeoReference(bbox2, srs4326)

	Grid := CaclulateGrid(512, 512, BORDER_BILATERAL, georef)

	if Grid == nil {
		t.FailNow()
	}
}

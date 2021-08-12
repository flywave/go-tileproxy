package raster

import (
	"testing"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

func TestRaster(t *testing.T) {
	opts := &RasterOptions{Format: tile.TileFormat("webp"), Mode: BORDER_BILATERAL}

	source := NewDemRasterSource(ModeMapbox, opts)

	source.SetSource("../data/14_13733_6366.webp")
	t1 := source.GetTile()

	if t1 == nil {
		t.FailNow()
	}

	srs900913 := geo.NewSRSProj4("EPSG:900913")
	srs4326 := geo.NewSRSProj4("EPSG:4326")

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)

	georef := geo.NewGeoReference(bbox2, srs4326)

	if georef == nil {
		t.FailNow()
	}
}

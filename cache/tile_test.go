package cache

import (
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestTransformCoord(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-20037508.342789244, -20037508.342789244}, Max: vec2d.T{20037508.342789244, 20037508.342789244}}
	grid := geo.NewTileGrid(opts)

	opts2 := geo.DefaultTileGridOptions()
	opts2[geo.TILEGRID_SRS] = "EPSG:3857"
	opts2[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts2[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-20037508.342789244, -20037508.342789244}, Max: vec2d.T{20037508.342789244, 20037508.342789244}}
	grid2 := geo.NewTileGrid(opts2)

	target, err := TransformCoord([3]int{1, 1, 1}, grid, grid2)

	if err != nil || target[0] == 0 {
		t.FailNow()
	}
}

func TestTransformCoord2(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	opts2 := geo.DefaultTileGridOptions()
	opts2[geo.TILEGRID_SRS] = "EPSG:4326"
	opts2[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts2[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid2 := geo.NewTileGrid(opts2)

	target, err := TransformCoord([3]int{0, 0, 2}, grid, grid2)

	if err != nil || target != [3]int{0, 1, 2} {
		t.FailNow()
	}

	grid3 := geo.NewTileGrid(opts)

	target, err = TransformCoord([3]int{0, 0, 2}, grid, grid3)

	if err != nil || target != [3]int{0, 0, 2} {
		t.FailNow()
	}
}

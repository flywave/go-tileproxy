package client

import (
	"testing"

	"github.com/flywave/go-tileproxy/geo"
)

func TestTileURLTemplate(t *testing.T) {
	ut := NewURLTemplate("/key={{ .quadkey }}&format={{ .format }}", "png")
	url := ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/{{ .tc_path }}.png", "")

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/x={{ .x }}&y={{ .y }}&z={{ .z }}&format={{ .format }}", "png")

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/{{ .arcgiscache_path }}.png", "")

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/service?BBOX={{ .bbox }}", "")
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	grid := geo.NewTileGrid(opts)

	url = ut.substitute([3]int{0, 1, 2}, nil, grid)

	if url == "" {
		t.FailNow()
	}
}

func TestTileClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	ut := NewURLTemplate("/key={{ .quadkey }}&format={{ .format }}", "png")
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	grid := geo.NewTileGrid(opts)

	client := NewTileClient(grid, ut, mock)

	ret := client.GetTile([3]int{5, 13, 9}, nil)

	if mock.url == "" || ret != nil {
		t.FailNow()
	}
}

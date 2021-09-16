package client

import (
	"testing"

	"github.com/flywave/go-geo"
)

func TestTileURLTemplate(t *testing.T) {
	ut := NewURLTemplate("/key={quadkey}&format={format}", "png", nil)
	url := ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/{tc_path}.png", "", nil)

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/x={x}&y={y}&z={z}&format={format}", "png", nil)

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/{arcgiscache_path}.png", "", nil)

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("{subdomains}/{arcgiscache_path}.png", "", []string{"a", "b", "c", "d"})

	url = ut.substitute([3]int{5, 13, 9}, nil, nil)

	if url == "" {
		t.FailNow()
	}

	ut = NewURLTemplate("/service?BBOX={bbox}", "", nil)
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
	ctx := &mockContext{c: mock}

	ut := NewURLTemplate("/key={quadkey}&format={format}", "png", nil)
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	grid := geo.NewTileGrid(opts)

	client := NewTileClient(grid, ut, ctx)

	ret := client.GetTile([3]int{5, 13, 9}, nil)

	if mock.url == "" || ret == nil {
		t.FailNow()
	}
}

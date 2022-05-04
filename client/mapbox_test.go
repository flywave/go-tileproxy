package client

import (
	"testing"
)

func TestMapboxTileClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	client := NewMapboxTileClient("http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8/{z}/{x}/{y}.vector.pbf",
		"http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8", "101XxiLvoFYxL",
		"pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja3pjOXRqcWkybWY3MnVwaGxkbTgzcXAwIn0._tCv9fpOyCT4O_Tdpl6h0w", "access_token", ctx)

	client.GetTile([3]int{1171, 1566, 12})

	if mock.url == "" {
		t.FailNow()
	}
}

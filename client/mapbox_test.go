package client

import (
	"testing"
)

func TestMapboxTileClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	client := NewMapboxTileClient("https://api.mapbox.com/v4/mapbox.mapbox-streets-v8.json",
		"https://api.mapbox.com/tilestats/v1/mapbox/mapbox.mapbox-streets-v8", "101XxiLvoFYxL",
		"pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja3pjOXRqcWkybWY3MnVwaGxkbTgzcXAwIn0._tCv9fpOyCT4O_Tdpl6h0w", "access_token", ctx)

	client.GetTileStats()

	if mock.url == "" {
		t.FailNow()
	}
}

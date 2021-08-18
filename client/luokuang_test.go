package client

import (
	"testing"

	"github.com/flywave/go-tileproxy/layer"
)

func TestLuokuangTileClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	query := &layer.LuoKuangTileQuery{X: 1171, Y: 1566, Zoom: 12, Width: 256, Height: 256, Format: "pbf", Style: "main"}

	client := NewLuoKuangTileClient("https://api.luokuang.com", "DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8", "map", ctx)

	client.GetTile(query)

	if mock.url == "" {
		t.FailNow()
	}
}

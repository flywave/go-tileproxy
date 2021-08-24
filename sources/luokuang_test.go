package sources

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

func TestLuoKuangTileSource(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	creater := &dummyCreater{}

	client := client.NewLuoKuangTileClient("https://api.luokuang.com", "DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8", "map", ctx)

	source := &LuoKuangTileSource{Grid: grid, Client: client, SourceCreater: creater}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

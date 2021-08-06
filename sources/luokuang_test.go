package sources

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestLuoKuangTileSource(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	creater := func(size [2]uint32, opts tile.TileOptions, data interface{}) tile.Source {
		if data != nil {
			reader := data.(io.Reader)
			buf, _ := ioutil.ReadAll(reader)
			return &tile.DummyTileSource{Data: string(buf)}
		}
		return nil
	}

	client := client.NewLuoKuangTileClient("https://api.luokuang.com", "DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8", "map", mock)

	source := &LuoKuangTileSource{Grid: grid, Client: client, SourceCreater: creater}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}
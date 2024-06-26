package sources

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type dummyCreater struct {
}

func (c *dummyCreater) GetExtension() string {
	return "png"
}

func (c *dummyCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return nil
}

func (c *dummyCreater) Create(data []byte, t [3]int) tile.Source {
	if data != nil {
		return &tile.DummyTileSource{Data: string(data)}
	}
	return nil
}

func TestMapboxTileSource(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	creater := &dummyCreater{}

	client := client.NewMapboxTileClient("https://api.mapbox.com/v4/mapbox.mapbox-streets-v8.json", "https://api.mapbox.com/tilestats/v1/mapbox/mapbox.mapbox-streets-v8", "", "{token}", "access_token", ctx)

	source := &MapboxTileSource{Grid: grid, Client: client, SourceCreater: creater}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewProj(4326), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

package sources

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type mockClient struct {
	client.HttpClient
	data []byte
	url  string
	body []byte
	code int
}

func (c *mockClient) Open(url string, data []byte) (statusCode int, body []byte) {
	c.data = data
	c.url = url
	return c.code, c.body
}

type mockContext struct {
	client.Context
	c *mockClient
}

func (c *mockContext) Client() client.HttpClient {
	return c.c
}

func (c *mockContext) Sync() {
}

func TestTileSource(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{tms_path}.png", "", nil)

	client := client.NewTileClient(grid, urlTemplate, ctx)

	creater := &dummyCreater{}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	source := &TileSource{Grid: grid, Client: client, SourceCreater: creater}

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewProj(4326), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

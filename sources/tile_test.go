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

type MockClient struct {
	client.HttpClient
	data []byte
	url  string
	body []byte
	code int
}

func (c *MockClient) Open(url string, data []byte) (statusCode int, body []byte) {
	c.data = data
	c.url = url
	return c.code, c.body
}

func TestTileSource(t *testing.T) {
	mock := &MockClient{code: 200, body: []byte{0}}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{{ .tms_path }}.png", "")

	client := client.NewTileClient(grid, urlTemplate, mock)

	creater := func(size [2]uint32, opts tile.TileOptions, data interface{}) tile.Source {
		if data != nil {
			reader := data.(io.Reader)
			buf, _ := ioutil.ReadAll(reader)
			return &tile.DummyTileSource{Data: string(buf)}
		}
		return nil
	}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	source := &TileSource{Grid: grid, Client: client, SourceCreater: creater}

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

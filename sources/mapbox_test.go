package sources

import (
	"io"
	"io/ioutil"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

func TestMapboxTileSource(t *testing.T) {
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

	client := client.NewMapboxTileClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", "mapbox.mapbox-streets-v8", mock)

	source := &MapboxTileSource{Grid: grid, Client: client, SourceCreater: creater}

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{256, 256}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

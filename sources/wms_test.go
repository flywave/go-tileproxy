package sources

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

func TestWMSSource(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, nil, nil, ctx)

	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	source := NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	box := grid.TileBBox([3]int{0, 0, 1}, false)

	query := &layer.MapQuery{BBox: box, Size: [2]uint32{512, 512}, Srs: geo.NewProj(900913), Format: tile.TileFormat("png")}

	resp, err := source.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

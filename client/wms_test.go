package client

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type mockClient struct {
	HttpClient
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
	Context
	c *mockClient
}

func (c *mockContext) Client() HttpClient {
	return c.c
}

func (c *mockContext) Sync() {
}

func TestWMSClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	param := http.Header{
		"layers":      []string{"foo"},
		"transparent": []string{"true"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)
	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}, Size: [2]uint32{512, 512}, Srs: geo.NewSRSProj4("EPSG:900913"), Format: tile.TileFormat("png")}

	client := NewWMSClient(req, ctx)
	format := tile.TileFormat("png")
	client.Retrieve(query, &format)

	if mock.url == "" {
		t.FailNow()
	}
}

func TestWMSInfoClient(t *testing.T) {
	mockF := "text"
	mock := &mockClient{code: 200, body: []byte(mockF)}
	ctx := &mockContext{c: mock}

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSFeatureInfoRequest(param, "/service?map=foo", false, nil, false)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewSRSProj4("EPSG:25832")}}

	client := NewWMSInfoClient(req, srs, ctx)

	query := &layer.InfoQuery{BBox: vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}, Size: [2]uint32{512, 512}, Srs: geo.NewSRSProj4("EPSG:4326"), Pos: [2]float64{128, 64}, Format: "text/plain"}

	feature := client.GetInfo(query)

	if mock.url == "" || feature.ToString() != "text" {
		t.FailNow()
	}
}

func TestWMSLegendClient(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}
	ctx := &mockContext{c: mock}

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSLegendGraphicRequest(param, "/service?map=foo", false, nil, false)

	client := NewWMSLegendClient(req, ctx)

	query := &layer.LegendQuery{Scale: 2}

	src := client.GetLegend(query)

	if mock.url == "" && src != nil {
		t.FailNow()
	}
}

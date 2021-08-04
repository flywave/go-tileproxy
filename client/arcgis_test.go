package client

import (
	"net/http"
	"testing"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestArcGISClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewArcGISRequest(param, "/MapServer/export?map=foo", false, nil)
	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}, Size: [2]uint32{512, 512}, Srs: geo.NewSRSProj4("EPSG:900913"), Format: tile.TileFormat("png")}

	client := NewArcGISClient(req, mock)
	format := tile.TileFormat("png")
	client.Retrieve(query, &format)

	if mock.url == "" {
		t.FailNow()
	}
}

func TestArcGISInfoClient(t *testing.T) {
	mockF := "text"
	mock := &mockClient{code: 200, body: []byte(mockF)}

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewArcGISIdentifyRequest(param, "/MapServer/export?map=foo", false, nil)
	query := &layer.InfoQuery{BBox: vec2d.Rect{Min: vec2d.T{8, 50}, Max: vec2d.T{9, 51}}, Size: [2]uint32{512, 512}, Srs: geo.NewSRSProj4("EPSG:4326"), Pos: [2]float64{128, 64}, Format: "text/plain"}

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewSRSProj4("EPSG:4326")}}

	client := NewArcGISInfoClient(req, srs, mock, false, 5)

	feature := client.GetInfo(query)

	if mock.url == "" || feature.ToString() != "text" {
		t.FailNow()
	}
}

package client

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/flywave/go-tileproxy/layer"
)

func TestMapboxTileClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	query := &layer.TileQuery{X: 1171, Y: 1566, Zoom: 12, Width: 256, Height: 256, Format: "application/vnd.mapbox-vector-tile", Retina: true}

	client := NewMapboxTileClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", "mapbox.mapbox-streets-v8", mock)

	client.GetTile(query)

	if mock.url == "" {
		t.FailNow()
	}
}

func TestMapboxSpriteClient(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	query := &layer.SpriteQuery{StyleQuery: layer.StyleQuery{StyleID: "testStylteId"}, SpriteID: "testSprite", Retina: true}

	client := NewMapboxSpriteClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", mock)

	src := client.GetSprite(query)

	if mock.url == "" && src != nil {
		t.FailNow()
	}
}

func TestMapboxStyleClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	query := &layer.StyleQuery{StyleID: "testStylteId"}

	client := NewMapboxStyleClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", mock)

	client.GetStyle(query)

	if mock.url == "" {
		t.FailNow()
	}
}

func TestMapboxGlyphsClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}

	query := &layer.GlyphsQuery{Font: "Arial Unicode MS", Start: 0, End: 255}

	client := NewMapboxGlyphsClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", mock)

	f := client.GetGlyphs(query)

	if mock.url == "" && f != nil {
		t.FailNow()
	}
}

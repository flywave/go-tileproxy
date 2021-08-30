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
	ctx := &mockContext{c: mock}

	client := NewMapboxTileClient("http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8/{z}/{x}/{y}.vector.pbf", "http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8", "{token}", "mapbox.mapbox-streets-v8", ctx)

	client.GetTile([3]int{1171, 1566, 12})

	if mock.url == "" {
		t.FailNow()
	}
}

func TestMapboxSpriteClient(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}
	ctx := &mockContext{c: mock}

	client := NewMapboxStyleClient("http://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7", "{token}", "access_token", ctx)

	src := client.GetSprite()

	if mock.url == "" && src != nil {
		t.FailNow()
	}
}

func TestMapboxStyleClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	client := NewMapboxStyleClient("http://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7", "{token}", "access_token", ctx)

	client.GetStyle()

	if mock.url == "" {
		t.FailNow()
	}
}

func TestMapboxGlyphsClient(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	query := &layer.GlyphsQuery{Font: "Arial Unicode MS", Start: 0, End: 255}

	client := NewMapboxStyleClient("http://api.mapbox.com/fonts/v1/examples", "{token}", "access_token", ctx)

	f := client.GetGlyphs(query)

	if mock.url == "" && f != nil {
		t.FailNow()
	}
}

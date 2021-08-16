package tile

import "testing"

func TestFormat(t *testing.T) {
	f := TileFormat("application/vnd.mapbox-vector-tile")
	if f.MimeType() != "application/vnd.mapbox-vector-tile" {
		t.FailNow()
	}
	if f.Extension() != "mvt" {
		t.FailNow()
	}

	f = TileFormat("image/webp")
	if f.MimeType() != "image/webp" {
		t.FailNow()
	}
	if f.Extension() != "webp" {
		t.FailNow()
	}
}

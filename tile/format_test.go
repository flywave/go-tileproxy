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

	f = TileFormat("image/vnd.wap.wbmp")
	if f.MimeType() != "image/vnd.wap.wbmp" {
		t.FailNow()
	}
	if f.Extension() != "wbmp" {
		t.FailNow()
	}
}

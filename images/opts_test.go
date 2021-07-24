package images

import (
	"testing"

	"github.com/flywave/go-tileproxy/tile"
)

func TestImageFormat(t *testing.T) {
	png1 := tile.TileFormat("image/png")
	if png1.Extension() != "png" {
		t.FailNow()
	}
	png2 := tile.TileFormat("png")
	if png2.Extension() != "png" {
		t.FailNow()
	}
	if png2.MimeType() != "image/png" {
		t.FailNow()
	}
}

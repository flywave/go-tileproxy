package images

import "testing"

func TestImageFormat(t *testing.T) {
	png1 := ImageFormat("image/png")
	if png1.Extension() != "png" {
		t.FailNow()
	}
	png2 := ImageFormat("png")
	if png2.Extension() != "png" {
		t.FailNow()
	}
	if png2.MimeType() != "image/png" {
		t.FailNow()
	}
}

package imagery

import (
	"image/color"
	"testing"

	"github.com/flywave/imaging"
	"github.com/fogleman/gg"
)

func TestTextDraw(t *testing.T) {
	tx := NewTextDraw("test", loadDefaultFontFace(), color.Black, color.White, "lr", 1, 0)
	gtx := gg.NewContext(600, 400)
	tx.draw(gtx, []uint32{600, 400})
}

func TestExceptionImage(t *testing.T) {
	img, _ := imaging.Open("../data/flowers.png")

	imgs := CreateImageSourceFromImage(img, PNG_FORMAT)

	ei := NewExceptionImage("message", PNG_FORMAT)

	_, err := ei.Draw(imgs, nil, false)

	if err != nil {
		t.FailNow()
	}
}

package cache

import (
	"image/color"

	"github.com/flywave/go-tileproxy/imagery"
)

type Watermark struct {
	Filter
	text      string
	opacity   *float64
	spacing   *string
	fontSize  *int
	fontColor *color.Color
}

func NewWatermark(text string, opacity *float64, spacing *string, font_size *int, font_color *color.Color) *Watermark {
	return &Watermark{text: text, opacity: opacity, spacing: spacing, fontSize: font_size, fontColor: font_color}
}

func tileWatermarkPlacement(coord [3]int, double_spacing bool) string {
	if !double_spacing {
		if coord[1]%2 == 0 {
			return "c"
		} else {
			return "b"
		}
	}

	if coord[1]%2 != coord[0]%2 {
		return "c"
	}

	return ""
}

func (w *Watermark) Apply(tile *Tile) *Tile {
	double_spacing := false
	if w.spacing != nil && *w.spacing == "wide" {
		double_spacing = true
	}
	placement := tileWatermarkPlacement(tile.Coord, double_spacing)
	wimg := imagery.NewWatermarkImage(w.text, tile.Source.GetTileOptions().(*imagery.ImageOptions),
		placement, w.opacity, w.fontColor, w.fontSize)
	tile.Source, _ = wimg.Draw(tile.Source, nil, false)
	return tile
}

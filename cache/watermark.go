package cache

import (
	"image/color"

	"github.com/flywave/go-tileproxy/images"
)

type Watermark struct {
	Filter
	text       string
	opacity    float64
	spacing    string
	font_size  int
	font_color color.Color
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
	placement := tileWatermarkPlacement(tile.Coord, w.spacing == "wide")
	wimg := images.NewWatermarkImage(w.text, tile.Source.GetTileOptions().(*images.ImageOptions),
		placement, &w.opacity, &w.font_color, &w.font_size)
	tile.Source, _ = wimg.Draw(tile.Source, nil, false)
	return tile
}

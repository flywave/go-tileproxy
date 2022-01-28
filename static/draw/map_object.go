package draw

import (
	"github.com/flywave/gg"
	"github.com/flywave/go-geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MapObject interface {
	Bounds() vec2d.Rect
	SrsProj() geo.Proj
	ExtraMarginPixels() (float64, float64, float64, float64)
	Draw(dc *gg.Context, trans *Transformer)
}

func CanDisplay(pos vec2d.T) bool {
	const minLatitude float64 = -85.0
	const maxLatitude float64 = 85.0
	return (minLatitude <= pos[0]) && (pos[0] <= maxLatitude)
}

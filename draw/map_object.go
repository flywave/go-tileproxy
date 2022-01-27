package draw

import (
	"github.com/flywave/gg"
	"github.com/golang/geo/s2"
)

type MapObject interface {
	Bounds() s2.Rect
	ExtraMarginPixels() (float64, float64, float64, float64)
	Draw(dc *gg.Context, trans *Transformer)
}

func CanDisplay(pos s2.LatLng) bool {
	const minLatitude float64 = -85.0
	const maxLatitude float64 = 85.0
	return (minLatitude <= pos.Lat.Degrees()) && (pos.Lat.Degrees() <= maxLatitude)
}

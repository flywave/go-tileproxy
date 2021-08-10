package vector

import (
	"github.com/flywave/go-tileproxy/geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type VectorTransformer struct {
	SrcSRS  geo.Proj
	DstSRS  geo.Proj
	DstBBox *vec2d.Rect
	DstSize *vec2d.Rect
}

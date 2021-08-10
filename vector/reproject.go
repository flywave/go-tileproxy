package vector

import (
	"github.com/flywave/go-tileproxy/geo"
)

type VectorTransformer struct {
	SrcSRS geo.Proj
	DstSRS geo.Proj
}

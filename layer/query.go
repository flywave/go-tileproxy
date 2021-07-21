package layer

import (
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/utils"
)

type Query interface{}

type MapQuery struct {
	Query
	BBox        vec2d.Rect
	Size        [2]uint32
	Srs         geo.Proj
	Format      images.ImageFormat
	Transparent bool
	TiledOnly   bool
	Dimensions  map[string]string
}

func (q *MapQuery) DimensionsForParams(params map[string]string) map[string]string {
	keys := []string{}
	for k := range params {
		keys = append(keys, strings.ToUpper(k))
	}
	result := make(map[string]string)
	for k, v := range q.Dimensions {
		if utils.ContainsString(keys, k) {
			result[k] = v
		}
	}
	return result
}

func EqualsParams(params1, params2 map[string]string) bool {
	if len(params1) != len(params2) {
		return false
	}
	for k, v := range params1 {
		if ov, ok := params2[k]; ok {
			if v != ov {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

type InfoQuery struct {
	Query
	BBox         vec2d.Rect
	Size         [2]uint32
	Srs          geo.Proj
	Pos          [2]float64
	InfoFormat   string
	Format       string
	FeatureCount int
}

func (i *InfoQuery) GetCoord() []float64 {
	return geo.MakeLinTransf(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(i.Size[0]), float64(i.Size[1])}}, i.BBox)(i.Pos[:])
}

type LegendQuery struct {
	Query
	Format string
	Scale  int
}

type StyleQuery struct {
	Query
}

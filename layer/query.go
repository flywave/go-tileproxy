package layer

import (
	"github.com/flywave/go-tileproxy/geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type Query interface{}

type MapQuery struct {
	Query
	BBox        vec2d.Rect
	Size        [2]uint32
	Srs         geo.Proj
	Format      string
	Transparent bool
	TiledOnly   bool
	Dimensions  int
}

func (q *MapQuery) DimensionsForParams(params map[string]string) map[string]string {
	return nil
}

func EqualsParams(params1, params2 map[string]string) bool {
	return false
}

type InfoQuery struct {
	Query
	BBox         vec2d.Rect
	Size         [2]uint32
	Srs          geo.Proj
	Pos          [2]int
	InfoFormat   string
	Format       string
	FeatureCount int
}

type LegendQuery struct {
	Query
	Format string
	Scale  int
}

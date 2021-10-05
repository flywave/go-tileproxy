package layer

import (
	"fmt"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type Query interface{}

type MapQuery struct {
	Query
	BBox        vec2d.Rect
	Size        [2]uint32
	Srs         geo.Proj
	Format      tile.TileFormat
	Transparent bool
	TiledOnly   bool
	Dimensions  utils.Dimensions
}

func (q *MapQuery) Eq(o *MapQuery) bool {
	if q.BBox.Max != o.BBox.Max || q.BBox.Min != o.BBox.Min {
		return false
	}
	if q.Size[0] != o.Size[0] || q.Size[1] != o.Size[1] {
		return false
	}
	if !q.Srs.Eq(o.Srs) {
		return false
	}
	if q.Format != o.Format {
		return false
	}
	if q.Transparent != o.Transparent {
		return false
	}
	if q.TiledOnly != o.TiledOnly {
		return false
	}
	if !q.Dimensions.Eq(o.Dimensions) {
		return false
	}
	return true
}

func (q *MapQuery) DimensionsForParams(params map[string]string) map[string]string {
	keys := []string{}
	for k := range params {
		keys = append(keys, strings.ToLower(k))
	}
	result := make(map[string]string)
	for k, v := range q.Dimensions {
		if utils.ContainsString(keys, k) {
			result[k] = utils.ValueToString(v.GetFirstValue())
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
	FeatureCount *int
}

func (q *InfoQuery) Eq(o *InfoQuery) bool {
	if q.BBox.Max != o.BBox.Max || q.BBox.Min != o.BBox.Min {
		return false
	}
	if q.Size[0] != o.Size[0] || q.Size[1] != o.Size[1] {
		return false
	}
	if !q.Srs.Eq(o.Srs) {
		return false
	}
	if q.Pos[0] != o.Pos[0] || q.Pos[1] != o.Pos[1] {
		return false
	}
	if q.Format != o.Format {
		return false
	}
	if q.InfoFormat != o.InfoFormat {
		return false
	}
	if q.FeatureCount != o.FeatureCount {
		return false
	}
	return true
}

func (i *InfoQuery) GetCoord() []float64 {
	return geo.MakeLinTransf(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(i.Size[0]), float64(i.Size[1])}}, i.BBox)(i.Pos[:])
}

type LegendQuery struct {
	Query
	Format string
	Scale  int
}

func (q *LegendQuery) Eq(o *LegendQuery) bool {
	if q.Format != o.Format {
		return false
	}
	if q.Scale != o.Scale {
		return false
	}
	return true
}

type GlyphsQuery struct {
	Query
	Font  string
	Start int
	End   int
}

func (req *GlyphsQuery) GetID() string {
	return fmt.Sprintf("%s-%d-%d", req.Font, req.Start, req.End)
}

func (q *GlyphsQuery) Eq(o *GlyphsQuery) bool {
	if q.Start != o.Start {
		return false
	}
	if q.End != o.End {
		return false
	}
	return true
}

package layer

import (
	"fmt"
	"net/url"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
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

func (q *MapQuery) EQ(o *MapQuery) bool {
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
	if !q.Dimensions.EQ(o.Dimensions) {
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
			result[k] = v.GetFirstValue()
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

func (q *InfoQuery) EQ(o *InfoQuery) bool {
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

func (q *LegendQuery) EQ(o *LegendQuery) bool {
	if q.Format != o.Format {
		return false
	}
	if q.Scale != o.Scale {
		return false
	}
	return true
}

type StyleQuery struct {
	Query
	StyleID string
}

func (req *StyleQuery) GetID() string {
	return req.StyleID
}

func (q *StyleQuery) EQ(o *StyleQuery) bool {
	if q.StyleID != o.StyleID {
		return false
	}
	return true
}

func (req *StyleQuery) BuildURL(URL string, username string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/styles/v1/%s/%s", URL, username, req.StyleID)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type TileQuery struct {
	Query
	Y       int
	X       int
	Zoom    int
	Width   int
	Height  int
	Format  string
	Retina  bool
	Markers []*Marker
}

func (q *TileQuery) EQ(o *TileQuery) bool {
	if q.Y != o.Y {
		return false
	}
	if q.X != o.X {
		return false
	}
	if q.Zoom != o.Zoom {
		return false
	}
	if q.Width != o.Width {
		return false
	}
	if q.Height != o.Height {
		return false
	}
	if q.Format != o.Format {
		return false
	}
	if q.Retina != o.Retina {
		return false
	}
	return true
}

func (req *TileQuery) BuildURL(URL string, accessToken string, mapid string) (string, error) {
	urls := fmt.Sprintf("%s/v4/%s", URL, url.QueryEscape(mapid))
	if len(req.Markers) > 0 {
		s := ""
		for i, marker := range req.Markers {
			if i > 0 {
				s += "/"
			}
			s += marker.String()
		}
		urls += s
	}
	urls += fmt.Sprintf("/%d/%d/%d", req.Zoom, req.X, req.Y)
	if req.Retina {
		urls += "@2x"
	}
	urls += fmt.Sprintf(".%s", req.Format)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type SpriteQuery struct {
	StyleQuery
	SpriteID string
	Retina   bool
	Format   *string
}

func (q *SpriteQuery) EQ(o *SpriteQuery) bool {
	if !q.StyleQuery.EQ(&o.StyleQuery) {
		return false
	}
	if q.SpriteID != o.SpriteID {
		return false
	}
	return true
}

func (req *SpriteQuery) GetID() string {
	return fmt.Sprintf("%s-%s", req.StyleID, req.SpriteID)
}

func (req *SpriteQuery) BuildURL(URL string, username string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/styles/v1/%s/%s/%s/sprite", URL, username, req.StyleID, req.SpriteID)
	if req.Retina {
		urls += "@2x"
	}
	if req.Format != nil {
		urls += fmt.Sprintf(".%s", *req.Format)
	}
	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
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

func (q *GlyphsQuery) EQ(o *GlyphsQuery) bool {
	if q.Font != o.Font {
		return false
	}
	if q.Start != o.Start {
		return false
	}
	if q.End != o.End {
		return false
	}
	return true
}

func (req *GlyphsQuery) BuildURL(URL string, username string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/fonts/v1/%s/%s/%d-%d.pbf", URL, username, req.Font, req.Start, req.End)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

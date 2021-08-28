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

func (req *StyleQuery) BuildURL(URL string, version string, username string, styleid string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/styles/%s/%s/%s", URL, version, username, styleid)

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
	Y         int
	X         int
	Zoom      int
	Width     int
	Height    int
	Format    string
	Retina    *int
	Markers   []*Marker
	TilesetID string
	Layer     *string
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
	if q.TilesetID != o.TilesetID {
		return false
	}
	return true
}

func (req *TileQuery) BuildURL(URL string, version string, accessToken string, tilesetid string) (string, error) {
	var urls string
	if req.Layer != nil {
		urls = fmt.Sprintf("%s/%s/%s/%s", URL, *req.Layer, version, url.QueryEscape(tilesetid))
	} else {
		urls = fmt.Sprintf("%s/%s/%s", URL, version, url.QueryEscape(tilesetid))
	}
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
	if req.Retina != nil {
		urls += fmt.Sprintf("@%dx", *req.Retina)
	}
	if req.Format != "" {
		if strings.Contains(req.Format, "/") {
			f := tile.TileFormat(req.Format)
			req.Format = f.Extension()
		}
		urls += fmt.Sprintf(".%s", req.Format)
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

type SpriteQuery struct {
	StyleQuery
	Retina *int
	Format *tile.TileFormat
}

func (q *SpriteQuery) EQ(o *SpriteQuery) bool {
	if !q.StyleQuery.EQ(&o.StyleQuery) {
		return false
	}
	return true
}

func (req *SpriteQuery) GetID() string {
	return req.StyleID
}

func (req *SpriteQuery) BuildURL(URL string, version string, username string, styleid string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/styles/%s/%s/%s/sprite", URL, version, username, styleid)
	if req.Retina != nil {
		urls += fmt.Sprintf("@%dx", *req.Retina)
	}
	if req.Format != nil {
		urls += fmt.Sprintf(".%s", req.Format.Extension())
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

func (req *GlyphsQuery) BuildURL(URL string, version string, username string, font string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/fonts/%s/%s/%s/%d-%d.pbf", URL, version, username, font, req.Start, req.End)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type TileJSONQuery struct {
	Query
	TilesetID string
}

func (req *TileJSONQuery) GetID() string {
	return fmt.Sprintf("%s", req.TilesetID)
}

func (q *TileJSONQuery) EQ(o *TileJSONQuery) bool {
	if q.TilesetID != o.TilesetID {
		return false
	}
	return true
}

func (req *TileJSONQuery) BuildURL(URL string, version string, username string, tilesetid string, accessToken string) (string, error) {
	urls := fmt.Sprintf("%s/%s/%s.json", URL, version, tilesetid)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type LuoKuangTileQuery struct {
	Query
	Y      int
	X      int
	Zoom   int
	Width  int
	Height int
	Format string
	Style  string
}

func (q *LuoKuangTileQuery) EQ(o *LuoKuangTileQuery) bool {
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
	if q.Style != o.Style {
		return false
	}
	return true
}

func (req *LuoKuangTileQuery) BuildURL(URL string, version string, accessToken string, styleid string) (string, error) {
	urls := fmt.Sprintf("%s/emg/%s/%s/tile", URL, version, styleid)

	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}

	q := u.Query()
	if strings.Contains(req.Format, "/") {
		f := tile.TileFormat(req.Format)
		req.Format = f.Extension()
	}
	q.Set("format", req.Format)
	q.Set("style", req.Style)
	q.Set("zoom", fmt.Sprintf("%d", req.Zoom))
	q.Set("x", fmt.Sprintf("%d", req.X))
	q.Set("y", fmt.Sprintf("%d", req.Y))
	q.Set("ak", accessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

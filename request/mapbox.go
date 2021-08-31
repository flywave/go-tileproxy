package request

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

var (
	MapbpxFormats = []string{"mvt", "vector.pbf", "grid.json", "png", "png32", "png64", "png128", "png256", "jpg70", "jpg80", "jpg90"}
)

type MapboxRequest struct {
	BaseRequest
	RequestHandlerName string
	AccessToken        string
	ReqRegex           *regexp.Regexp
	Version            string
}

func (r *MapboxRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

type MapboxTileJSONRequest struct {
	MapboxRequest
	TilesetID string
	Secure    bool
}

func NewMapboxTileJSONRequest(hreq *http.Request, validate bool) *MapboxTileJSONRequest {
	req := &MapboxTileJSONRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *MapboxTileJSONRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tilejson"
	r.Version = "v4"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.Secure = false
	r.ReqRegex = regexp.MustCompile(`^/(?P<version>[^/]+)/(?P<tileset_id>[^/]+).json`)
	r.initRequest()
}

func (r *MapboxTileJSONRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	if len(match) == 0 {
		return errors.New("url error")
	}
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["tileset_id"]; ok {
		r.TilesetID = v
	}

	return nil
}

type MapboxTileRequest struct {
	MapboxRequest
	TilesetID string
	Tile      []int
	Format    *tile.TileFormat
	Retina    *int
	Origin    string
}

func (r *MapboxTileRequest) GetFormat() *tile.TileFormat {
	return r.Format
}

func (r *MapboxTileRequest) GetTile() [3]int {
	return [3]int{r.Tile[0], r.Tile[1], r.Tile[2]}
}

func (r *MapboxTileRequest) GetOriginString() string {
	return r.Origin
}

func (r *MapboxTileRequest) GetOrigin() geo.OriginType {
	return geo.OriginFromString(r.Origin)
}

func NewMapboxTileRequest(hreq *http.Request, validate bool) *MapboxTileRequest {
	req := &MapboxTileRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *MapboxTileRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.Version = "v4"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.Retina = nil
	r.ReqRegex = regexp.MustCompile(`^/(?P<version>[^/]+)/(?P<tileset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)(@(?P<retina>[^/]+)x)?\.(?P<format>\w+)`)
	r.initRequest()
}

func (r *MapboxTileRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	if len(match) == 0 {
		return errors.New("url error")
	}
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	if len(match) == 0 {
		return errors.New("error")
	}
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["tileset_id"]; ok {
		r.TilesetID = v
	}
	if v, ok := result["retina"]; ok {
		rt, _ := strconv.ParseInt(v, 10, 64)
		r.Retina = geo.NewInt(int(rt))
	}
	if r.Tile == nil {
		x, _ := strconv.ParseInt(result["x"], 10, 64)
		y, _ := strconv.ParseInt(result["y"], 10, 64)
		z, _ := strconv.ParseInt(result["zoom"], 10, 64)
		r.Tile = []int{int(x), int(y), int(z)}
	}
	if v, ok := result["format"]; ok {
		tf := tile.TileFormat(v)
		r.Format = &tf
	}
	return nil
}

type MapboxStyleRequest struct {
	MapboxRequest
	Username string
	StyleID  string
}

func NewMapboxStyleRequest(hreq *http.Request, validate bool) *MapboxStyleRequest {
	req := &MapboxStyleRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *MapboxStyleRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "style"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/styles/(?P<version>[^/]+)/((?P<username>[^/]+)/)?(?P<style_id>[^/]+)`)
	r.initRequest()
}

func (r *MapboxStyleRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	if len(match) == 0 {
		return errors.New("url error")
	}
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		r.Username = v
	}
	if v, ok := result["style_id"]; ok {
		r.StyleID = v
	}
	return nil
}

type MapboxSpriteRequest struct {
	MapboxRequest
	Username string
	StyleID  string
	Format   *tile.TileFormat
	Retina   *int
}

func NewMapboxSpriteRequest(hreq *http.Request, validate bool) *MapboxSpriteRequest {
	req := &MapboxSpriteRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *MapboxSpriteRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "sprite"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/styles/(?P<version>[^/]+)/((?P<username>[^/]+)/)?(?P<style_id>[^/]+)/sprite(@(?P<retina>[^/]+)x)?\.?(?P<format>\w+)?`)
	r.initRequest()
}

func (r *MapboxSpriteRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	if len(match) == 0 {
		return errors.New("url error")
	}
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		result[name] = match[i]
	}

	if len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		r.Username = v
	}
	if v, ok := result["style_id"]; ok {
		r.StyleID = v
	}
	if v, ok := result["retina"]; ok {
		if v != "" {
			rt, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				r.Retina = geo.NewInt(int(rt))
			}
		}
	}
	if v, ok := result["format"]; ok {
		f := tile.TileFormat(v)
		r.Format = &f
	}
	return nil
}

type MapboxGlyphsRequest struct {
	MapboxRequest
	Username string
	Font     string
	Start    int
	End      int
}

func NewMapboxGlyphsRequest(hreq *http.Request, validate bool) *MapboxGlyphsRequest {
	req := &MapboxGlyphsRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *MapboxGlyphsRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "glyphs"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/fonts/(?P<version>[^/]+)/((?P<username>[^/]+)/)?(?P<font>[^/]+)/(?P<start>-?\d+)-(?P<end>-?\d+)\.(?P<format>\w+)`)
	r.initRequest()
}

func (r *MapboxGlyphsRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	if len(match) == 0 {
		return errors.New("url error")
	}
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["format"]; !ok || v != "pbf" {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		str, _ := url.PathUnescape(v)
		r.Username = str
	}
	if v, ok := result["font"]; ok {
		r.Font = v
	}

	if v, ok := result["start"]; ok {
		r.Start, _ = strconv.Atoi(v)
	}

	if v, ok := result["end"]; ok {
		r.End, _ = strconv.Atoi(v)
	}

	return nil
}

func MakeMapboxRequest(req *http.Request, validate bool) Request {
	url := req.URL.String()
	if strings.Contains(url, "/styles/") {
		if strings.Contains(url, "/sprite") {
			return NewMapboxSpriteRequest(req, validate)
		} else {
			return NewMapboxStyleRequest(req, validate)
		}
	} else if strings.Contains(url, "/fonts/") {
		return NewMapboxGlyphsRequest(req, validate)
	} else {
		if strings.Contains(url, ".json") {
			return NewMapboxTileJSONRequest(req, validate)
		} else {
			return NewMapboxTileRequest(req, validate)
		}
	}
}

package request

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/flywave/go-geo"
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
	FileName  string
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
	r.ReqRegex = regexp.MustCompile(`^/(?P<tileset_id>[^/]+)/(?P<file_name>[^/]+).json`) //`^/(?P<version>[^/]+)/(?P<tileset_id>[^/]+).json`
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

	if len(match) == 0 {
		return fmt.Errorf("invalid request (%s)", r.Http.URL.Path)
	}

	if v, ok := result["tileset_id"]; ok {
		r.TilesetID = v
	}

	if v, ok := result["file_name"]; ok {
		r.FileName = v
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
	r.ReqRegex = regexp.MustCompile(`^/(?P<tileset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)(@(?P<retina>[^/]+)x)?\.(?P<format>\w+)`) //`^/(?P<version>[^/]+)/(?P<tileset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)(@(?P<retina>[^/]+)x)?\.(?P<format>\w+)`
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

	if len(match) == 0 {
		return fmt.Errorf("invalid request (%s)", r.Http.URL.Path)
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

func MakeMapboxRequest(req *http.Request, validate bool) Request {
	url := req.URL.String()
	if strings.Contains(url, ".json") {
		return NewMapboxTileJSONRequest(req, validate)
	} else {
		return NewMapboxTileRequest(req, validate)
	}
}

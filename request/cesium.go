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

type CesiumRequest struct {
	BaseRequest
	RequestHandlerName string
	ReqRegex           *regexp.Regexp
	Version            string
}

func (r *CesiumRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

type CesiumLayerJSONRequest struct {
	CesiumRequest
	AssetID string
}

func NewCesiumLayerJSONRequest(hreq *http.Request, validate bool) *CesiumLayerJSONRequest {
	req := &CesiumLayerJSONRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq)
	return req
}

func (r *CesiumLayerJSONRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "layer.json"
	r.Version = "1.2.0"
	r.ReqRegex = regexp.MustCompile(`^/(?P<asset_id>[^/]+)/layer.json`)
	r.initRequest()
}

func (r *CesiumLayerJSONRequest) initRequest() error {
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

	if v, ok := result["asset_id"]; ok {
		r.AssetID = v
	}

	return nil
}

type CesiumTileRequest struct {
	CesiumRequest
	AssetID    string
	Tile       []int
	Extensions []string
	Format     *tile.TileFormat
}

func (r *CesiumTileRequest) GetOriginString() string {
	return "ul"
}

func (r *CesiumTileRequest) GetTile() [3]int {
	return [3]int{r.Tile[0], r.Tile[1], r.Tile[2]}
}

func (r *CesiumTileRequest) GetOrigin() geo.OriginType {
	return geo.ORIGIN_UL
}

func (r *CesiumTileRequest) GetFormat() *tile.TileFormat {
	return r.Format
}

func (r *CesiumTileRequest) GetExtensions() []string {
	return r.Extensions
}

func NewCesiumTileRequest(hreq *http.Request, validate bool) *CesiumTileRequest {
	req := &CesiumTileRequest{}
	req.init(hreq.Header, hreq.URL.Path, validate, hreq, hreq.URL.Query())
	return req
}

func (r *CesiumTileRequest) init(param interface{}, url string, validate bool, http *http.Request, query map[string][]string) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.Version = "1.2.0"
	r.ReqRegex = regexp.MustCompile(`^/(?P<asset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	r.initRequest(query)
}

func (r *CesiumTileRequest) initRequest(query map[string][]string) error {
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

	if v, ok := result["asset_id"]; ok {
		r.AssetID = v
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
	for k, v := range query {
		if k == "v" {
			r.Version = v[0]
		} else if k == "extensions" {
			r.Extensions = strings.Split(v[0], "-")
		}
	}

	return nil
}

func MakeCesiumRequest(req *http.Request, validate bool) Request {
	url := req.URL.String()
	if strings.Contains(url, "layer.json") {
		return NewCesiumLayerJSONRequest(req, validate)
	} else {
		return NewCesiumTileRequest(req, validate)
	}
}

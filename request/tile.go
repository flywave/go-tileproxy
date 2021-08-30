package request

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type TiledRequest interface {
	Request
	GetFormat() *tile.TileFormat
	GetTile() [3]int
	GetOriginString() string
	GetOrigin() geo.OriginType
}

type TileRequest struct {
	TiledRequest
	TileReqRegex       *regexp.Regexp
	RequestHandlerName string
	UseProfiles        bool
	RequestPrefix      string
	Origin             string
	Dimensions         map[string][]string
	Tile               []int
	Layer              string
	Format             *tile.TileFormat
	Http               *http.Request
}

func NewTileRequest(req *http.Request) *TileRequest {
	r := &TileRequest{Http: req}
	r.init()
	return r
}

func (r *TileRequest) GetFormat() *tile.TileFormat {
	return r.Format
}

func (r *TileRequest) GetTile() [3]int {
	return [3]int{r.Tile[0], r.Tile[1], r.Tile[2]}
}

func (r *TileRequest) GetOriginString() string {
	return r.Origin
}

func (r *TileRequest) GetOrigin() geo.OriginType {
	return geo.OriginFromString(r.Origin)
}

func (r *TileRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

func (r *TileRequest) init() {
	r.RequestHandlerName = "map"
	r.TileReqRegex = regexp.MustCompile(`(?P<begin>/[^/]+)/((?P<version>1\.0\.0)/)?(?P<layer>[^/]+)/((?P<layer_spec>[^/]+)/)?(?P<z>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	r.UseProfiles = false
	r.RequestPrefix = "/tiles"
	r.Dimensions = make(map[string][]string)
	r.initRequest()
}

func (r *TileRequest) initRequest() error {
	url := r.Http.URL.Path
	match := r.TileReqRegex.FindStringSubmatch(url)
	groupNames := r.TileReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 || result["begin"] != r.RequestPrefix {
		return errors.New(fmt.Sprintf("invalid request (%s)", url))
	}

	r.Layer = result["layer"]
	r.Dimensions = make(map[string][]string)
	if _, ok := result["layer_spec"]; ok {
		r.Dimensions["_layer_spec"] = []string{result["layer_spec"]}
	}
	if r.Tile == nil {
		x, _ := strconv.ParseInt(result["x"], 10, 64)
		y, _ := strconv.ParseInt(result["y"], 10, 64)
		z, _ := strconv.ParseInt(result["z"], 10, 64)
		r.Tile = []int{int(x), int(y), int(z)}
	}
	if format, ok := result["format"]; r.Format == nil && ok {
		tf := tile.TileFormat(format)
		r.Format = &tf
	}
	return nil
}

type TMSRequest struct {
	TileRequest
	CapabilitiesRegex *regexp.Regexp
	RootRequestRegex  *regexp.Regexp
}

func NewTMSRequest(req *http.Request) *TMSRequest {
	r := &TMSRequest{TileRequest: TileRequest{Http: req}}
	r.init()
	return r
}

func (r *TMSRequest) init() {
	r.TileRequest.init()
	r.CapabilitiesRegex = regexp.MustCompile(`^.*/1\.0\.0/?(/(?P<layer>[^/]+))?(/(?P<layer_spec>[^/]+))?$`)
	r.RootRequestRegex = regexp.MustCompile(`/tms/?$`)
	r.UseProfiles = true
	r.RequestPrefix = "/tms"
	r.Dimensions = make(map[string][]string)
	r.Origin = "sw"

	match := r.CapabilitiesRegex.FindStringSubmatch(r.Http.URL.Path)
	cap_match := make(map[string]string)

	if len(match) > 0 {
		groupNames := r.CapabilitiesRegex.SubexpNames()
		for i, name := range groupNames {
			if name != "" && match[i] != "" {
				cap_match[name] = match[i]
			}
		}
	}

	root_match := r.RootRequestRegex.MatchString(r.Http.URL.Path)

	if len(match) > 0 {
		if layer, ok := cap_match["layer"]; ok {
			r.Layer = layer
			if layer_spec, ok := cap_match["layer_spec"]; ok {
				r.Dimensions["_layer_spec"] = []string{layer_spec}
			}
		}
		r.RequestHandlerName = "tms_capabilities"
	} else if root_match {
		r.RequestHandlerName = "tms_root_resource"
	} else {
		r.RequestHandlerName = "map"
		r.initRequest()
	}
}

func MakeTileRequest(req *http.Request, validate bool) Request {
	if strings.Contains(req.URL.String(), "/tms") {
		return NewTMSRequest(req)
	} else {
		return NewTileRequest(req)
	}
}

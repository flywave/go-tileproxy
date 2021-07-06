package request

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/flywave/go-tileproxy/images"
)

type TileRequest struct {
	Request
	TileReqRegex       *regexp.Regexp
	RequestHandlerName string
	UseProfiles        bool
	RequestPrefix      string
	Origin             string
	Dimensions         map[string][]string
	Tile               []int
	Layer              string
	Format             *images.ImageFormat
	Http               *http.Request
}

func (r *TileRequest) init() {
	r.RequestHandlerName = "map"
	r.TileReqRegex = regexp.MustCompile(`^(?P<begin>/[^/]+)/
            ((?P<version>1\.0\.0)/)?
            (?P<layer>[^/]+)/
            ((?P<layer_spec>[^/]+)/)?
            (?P<z>-?\d+)/
            (?P<x>-?\d+)/
            (?P<y>-?\d+)\.(?P<format>\w+)`)
	r.UseProfiles = false
	r.RequestPrefix = "/tiles"
	r.Dimensions = make(map[string][]string)
	r.initRequest()
}

func (r *TileRequest) initRequest() error {
	match := r.TileReqRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.TileReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		result[name] = match[i]
	}

	if match == nil || len(match) == 0 || result["begin"] != r.RequestPrefix {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	r.Layer = result["layer"]
	r.Dimensions = make(map[string][]string)
	if _, ok := result["layer_spec"]; ok {
		r.Dimensions["_layer_spec"] = []string{result["layer_spec"]}
	}
	if r.Tile != nil {
		x, _ := strconv.ParseInt(result["x"], 10, 64)
		y, _ := strconv.ParseInt(result["y"], 10, 64)
		z, _ := strconv.ParseInt(result["z"], 10, 64)
		r.Tile = []int{int(x), int(y), int(z)}
	}
	if r.Format != nil {
		*r.Format = images.ImageFormat(result["format"])
	}
	return nil
}

type TMSRequest struct {
	TileRequest
	CapabilitiesRegex *regexp.Regexp
	RootRequestRegex  *regexp.Regexp
}

func (r *TMSRequest) init() {
	r.RequestHandlerName = "map"
	r.CapabilitiesRegex = regexp.MustCompile(`^.*/1\.0\.0/?
	(/(?P<layer>[^/]+))?
	(/(?P<layer_spec>[^/]+))?
	$`)
	r.RootRequestRegex = regexp.MustCompile(`/tms/?$`)
	r.UseProfiles = true
	r.RequestPrefix = "/tms"
	r.Dimensions = make(map[string][]string)
	r.Origin = "sw"

	match := r.CapabilitiesRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.CapabilitiesRegex.SubexpNames()
	cap_match := make(map[string]string)
	for i, name := range groupNames {
		cap_match[name] = match[i]
	}

	root_match := r.CapabilitiesRegex.FindString(r.Http.URL.Path)

	if len(cap_match) > 0 {
		if layer, ok := cap_match["layer"]; ok {
			r.Layer = layer
			if layer_spec, ok := cap_match["layer_spec"]; ok {
				r.Dimensions["_layer_spec"] = []string{layer_spec}
			}
		}
		r.RequestHandlerName = "tms_capabilities"
	} else if root_match != "" {
		r.RequestHandlerName = "tms_root_resource"
	} else {
		r.initRequest()
	}
}

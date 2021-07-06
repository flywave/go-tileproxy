package request

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	mapset "github.com/deckarep/golang-set"

	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/utils"
)

type WMTSTileRequestParams struct {
	RequestParams
}

func (r *WMTSTileRequestParams) GetLayer() string {
	val, ok := r.Get("layer")
	if ok {
		return val[0]
	}
	return ""
}

func (r *WMTSTileRequestParams) SetLayer(l string) {
	r.Set("layer", []string{l})
}

func (r *WMTSTileRequestParams) GetCoord() [3]int {
	x, err := strconv.Atoi(r.GetOne("tilecol", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	y, err := strconv.Atoi(r.GetOne("tilerow", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	z, err := strconv.Atoi(r.GetOne("tilematrix", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	return [3]int{x, y, z}
}

func (r *WMTSTileRequestParams) SetCoord(c [3]int) {
	r.Set("tilecol", []string{strconv.Itoa(c[0])})
	r.Set("tilerow", []string{strconv.Itoa(c[1])})
	r.Set("tilematrix", []string{strconv.Itoa(c[2])})
}

func (r *WMTSTileRequestParams) GetFormat() images.ImageFormat {
	strs := SplitMimeType(r.GetOne("format", ""))
	return images.ImageFormat(strs[1])
}

func (r *WMTSTileRequestParams) SetFormat(fmrt images.ImageFormat) {
	r.Set("tilematrix", []string{fmrt.MimeType()})
}

func (r *WMTSTileRequestParams) GetFormatMimeType() string {
	return r.GetOne("format", "")
}

func (r *WMTSTileRequestParams) GetDimensions() map[string][]string {
	expected_param := mapset.NewSet("version", "request", "layer", "style", "tilematrixset",
		"tilematrix", "tilerow", "tilecol", "format", "service")
	dimensions := make(map[string][]string)
	for key, value := range r.RequestParams {
		if !expected_param.Contains(key) {
			dimensions[strings.ToLower(key)] = value
		}
	}
	return dimensions
}

type WMTSRequest struct {
	BaseRequest
	RequestHandlerName string
	FixedParams        map[string]string
	ExpectedParam      []string
	NonStrict          bool
	NonStrictParams    mapset.Set
}

type WMTS100TileRequest struct {
	WMTSRequest
}

func (r *WMTS100TileRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.FixedParams = map[string]string{"request": "GetTile", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "layer", "style", "tilematrixset",
		"tilematrix", "tilerow", "tilecol", "format"}
}

func (r *WMTS100TileRequest) MakeRequest() map[string]interface{} {
	params := (*WMTSTileRequestParams)(unsafe.Pointer(&r.Params))
	req := make(map[string]interface{})
	req["layer"] = params.GetOne("layer", "")
	req["tilematrixset"] = params.GetOne("tilematrixset", "")
	req["format"] = images.ImageFormat(params.GetOne("format", ""))
	req["tile"] = params.GetCoord()
	req["origin"] = "nw"
	req["dimensions"] = params.GetDimensions()
	return req
}

func (r *WMTS100TileRequest) Validate() error {
	missing_param := []string{}
	for _, param := range r.ExpectedParam {
		if r.NonStrict && r.NonStrictParams.Contains(param) {
			continue
		}
		if !r.NonStrictParams.Contains(param) {
			missing_param = append(missing_param, param)
		}
	}

	if len(missing_param) > 0 {
		if utils.ContainsString(missing_param, "format") {
			r.Params.Set("format", []string{"image/png"})
		}
		return errors.New("missing parameters" + strings.Join(missing_param, ","))
	}

	return nil
}

type WMTSFeatureInfoRequestParams struct {
	WMTSTileRequestParams
}

func (r *WMTSFeatureInfoRequestParams) GetPos() [2]int {
	i, err := strconv.Atoi(r.GetOne("i", "-1"))
	if err != nil {
		return [2]int{-1, -1}
	}
	j, err := strconv.Atoi(r.GetOne("j", "-1"))
	if err != nil {
		return [2]int{-1, -1}
	}
	return [2]int{i, j}
}

func (r *WMTSFeatureInfoRequestParams) SetPos(pos [2]int) {
	r.Set("i", []string{strconv.Itoa(pos[0])})
	r.Set("j", []string{strconv.Itoa(pos[1])})
}

type WMTS100FeatureInfoRequest struct {
	WMTS100TileRequest
}

func (r *WMTS100FeatureInfoRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "featureinfo"
	r.FixedParams = map[string]string{"request": "GetFeatureInfo", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "layer", "style", "tilematrixset",
		"tilematrix", "tilerow", "tilecol", "format", "infoformat", "i", "j"}
	r.NonStrictParams = mapset.NewSet("format", "styles")
}

func (r *WMTS100FeatureInfoRequest) MakeRequest() map[string]interface{} {
	ret := r.WMTS100TileRequest.MakeRequest()
	params := (*WMTSFeatureInfoRequestParams)(unsafe.Pointer(&r.Params))

	ret["infoformat"] = params.GetOne("infoformat", "")
	ret["pos"] = params.GetPos()

	return ret
}

type WMTS100CapabilitiesRequest struct {
	WMTSRequest
	CapabilitiesTemplate string
	MimeType             string
}

func (r *WMTS100CapabilitiesRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "capabilities"
	r.CapabilitiesTemplate = "wmts100capabilities.xml"
	r.MimeType = "text/xml"
	r.FixedParams = map[string]string{}
}

func parseRequestType(req *http.Request) string {
	if _, ok := req.Header["request"]; ok {
		request_type := strings.ToLower(req.Header["request"][0])
		if strings.HasPrefix(request_type, "get") {
			request_type = request_type[3:]
			if utils.ContainsString([]string{"tile", "featureinfo", "capabilities"}, request_type) {
				return request_type
			}
		}
	}
	return ""
}

func MakeWMTSRequest(req *http.Request, validate bool) Request {
	req_type := parseRequestType(req)
	switch req_type {
	case "featureinfo":
		r := &WMTS100FeatureInfoRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	case "tile":
		r := &WMTS100TileRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	case "capabilities":
		r := &WMTS100CapabilitiesRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	}
	return nil
}

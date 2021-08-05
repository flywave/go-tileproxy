package request

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	mapset "github.com/deckarep/golang-set"
	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type WMTSTileRequestParams struct {
	params RequestParams
}

func NewWMTSTileRequestParams(params RequestParams) WMTSTileRequestParams {
	return WMTSTileRequestParams{params: params}
}

func (r *WMTSTileRequestParams) GetBBox() vec2d.Rect {
	if v, ok := r.params.Get("bbox"); !ok {
		return vec2d.Rect{}
	} else {
		if len(v) == 4 {
			bbox := [4]float64{}
			for i := range v {
				v, err := strconv.ParseFloat(v[i], 64)
				if err != nil {
					return vec2d.Rect{}
				}
				bbox[i] = v
			}
			return vec2d.Rect{Min: vec2d.T{bbox[0], bbox[1]}, Max: vec2d.T{bbox[2], bbox[3]}}
		} else if len(v) == 1 {
			bstr := strings.Split(v[0], ",")
			if len(bstr) == 4 {
				bbox := [4]float64{}
				for i := range bstr {
					v, err := strconv.ParseFloat(bstr[i], 64)
					if err != nil {
						return vec2d.Rect{}
					}
					bbox[i] = v
				}
				return vec2d.Rect{Min: vec2d.T{bbox[0], bbox[1]}, Max: vec2d.T{bbox[2], bbox[3]}}
			}
		}
	}
	return vec2d.Rect{}
}

func (r *WMTSTileRequestParams) SetBBox(bbox vec2d.Rect) {
	minx := strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'f', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'f', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *WMTSTileRequestParams) GetSize() [2]int {
	w, okw := r.params.Get("width")
	h, okh := r.params.Get("height")
	if !okw || !okh {
		return [2]int{-1, -1}
	} else {
		ws, err := strconv.ParseInt(w[0], 10, 64)
		if err != nil {
			return [2]int{-1, -1}
		}
		hs, err := strconv.ParseInt(h[0], 10, 64)
		if err != nil {
			return [2]int{-1, -1}
		}
		return [2]int{int(ws), int(hs)}
	}
}

func (r *WMTSTileRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("width", []string{width})
	r.params.Set("height", []string{height})
}

func (r *WMTSTileRequestParams) GetLayer() string {
	val, ok := r.params.Get("layer")
	if ok {
		return val[0]
	}
	return ""
}

func (r *WMTSTileRequestParams) SetLayer(l string) {
	r.params.Set("layer", []string{l})
}

func (r *WMTSTileRequestParams) SetLayers(ls []string) {
	r.params.Set("layer", ls)
}

func (r *WMTSTileRequestParams) GetTileMatrixSet() string {
	return r.params.GetOne("tilematrixset", "")
}

func (r *WMTSTileRequestParams) GetCoord() [3]int {
	x, err := strconv.Atoi(r.params.GetOne("tilecol", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	y, err := strconv.Atoi(r.params.GetOne("tilerow", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	z, err := strconv.Atoi(r.params.GetOne("tilematrix", "-1"))
	if err != nil {
		return [3]int{-1, -1, -1}
	}
	return [3]int{x, y, z}
}

func (r *WMTSTileRequestParams) SetCoord(c [3]int) {
	r.params.Set("tilecol", []string{strconv.Itoa(c[0])})
	r.params.Set("tilerow", []string{strconv.Itoa(c[1])})
	r.params.Set("tilematrix", []string{strconv.Itoa(c[2])})
}

func (r *WMTSTileRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *WMTSTileRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMTSTileRequestParams) GetFormatMimeType() string {
	return r.params.GetOne("format", "")
}

func (r *WMTSTileRequestParams) GetDimensions() map[string][]string {
	expected_param := mapset.NewSet("version", "request", "layer", "style", "tilematrixset",
		"tilematrix", "tilerow", "tilecol", "format", "service")
	dimensions := make(map[string][]string)
	for key, value := range r.params {
		if !expected_param.Contains(key) {
			dimensions[strings.ToLower(key)] = value
		}
	}
	return dimensions
}

func (r *WMTSTileRequestParams) Update(params map[string]string) {
	for key, value := range params {
		if _, ok := r.params[key]; !ok {
			r.params[key] = []string{value}
		} else {
			r.params[key] = append(r.params[key], value)
		}
	}
}

func (r *WMTSTileRequestParams) GetSrs() string {
	return r.params.GetOne("srs", "EPSG:4326")
}

func (r *WMTSTileRequestParams) SetSrs(srs string) {
	r.params.Set("srs", []string{srs})
}

type WMTSRequest struct {
	BaseRequest
	RequestHandlerName string
	ExpectedParam      []string
	NonStrict          bool
	NonStrictParams    mapset.Set
	UseProfiles        bool
}

type WMTS100TileRequest struct {
	WMTSRequest
}

func (r *WMTSRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

func NewWMTS100TileRequest(param interface{}, url string, validate bool, ht *http.Request) *WMTS100TileRequest {
	req := &WMTS100TileRequest{}
	req.init(param, url, validate, ht)
	return req
}

func (r *WMTS100TileRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.FixedParams = map[string]string{"request": "GetTile", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "layer", "style", "tilematrixset", "tilematrix", "tilerow", "tilecol", "format"}
}

func (r *WMTS100TileRequest) MakeRequest() map[string]interface{} {
	params := &WMTSTileRequestParams{params: r.Params}
	req := make(map[string]interface{})
	req["layer"] = params.params.GetOne("layer", "")
	req["tilematrixset"] = params.params.GetOne("tilematrixset", "")
	req["format"] = tile.TileFormat(params.params.GetOne("format", ""))
	req["tile"] = params.GetCoord()
	req["origin"] = geo.ORIGIN_NW
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

func NewWMTSFeatureInfoRequestParams(params RequestParams) WMTSFeatureInfoRequestParams {
	return WMTSFeatureInfoRequestParams{WMTSTileRequestParams: WMTSTileRequestParams{params: params}}
}

func (r *WMTSFeatureInfoRequestParams) GetInfoformat() string {
	return r.params.GetOne("infoformat", "")
}

func (r *WMTSFeatureInfoRequestParams) SetInfoformat(info string) {
	r.params.Set("infoformat", []string{info})
}

func (r *WMTSFeatureInfoRequestParams) GetPos() [2]float64 {
	i, err := strconv.ParseFloat(r.params.GetOne("i", "-1"), 64)
	if err != nil {
		return [2]float64{-1, -1}
	}
	j, err := strconv.ParseFloat(r.params.GetOne("j", "-1"), 64)
	if err != nil {
		return [2]float64{-1, -1}
	}
	return [2]float64{i, j}
}

func (r *WMTSFeatureInfoRequestParams) SetPos(pos [2]float64) {
	r.params.Set("i", []string{strconv.FormatFloat(pos[0], 'f', -1, 64)})
	r.params.Set("j", []string{strconv.FormatFloat(pos[1], 'f', -1, 64)})
}

type WMTS100FeatureInfoRequest struct {
	WMTS100TileRequest
}

func NewWMTS100FeatureInfoRequest(param interface{}, url string, validate bool, ht *http.Request) *WMTS100FeatureInfoRequest {
	req := &WMTS100FeatureInfoRequest{}
	req.init(param, url, validate, ht)
	return req
}

func (r *WMTS100FeatureInfoRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "featureinfo"
	r.FixedParams = map[string]string{"request": "GetFeatureInfo", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "layer", "style", "tilematrixset", "tilematrix", "tilerow", "tilecol", "format", "infoformat", "i", "j"}
	r.NonStrictParams = mapset.NewSet("format", "styles")
}

func (r *WMTS100FeatureInfoRequest) MakeRequest() map[string]interface{} {
	ret := r.WMTS100TileRequest.MakeRequest()
	params := (*WMTSFeatureInfoRequestParams)(unsafe.Pointer(&r.Params))

	ret["infoformat"], _ = params.params.Get("infoformat")
	ret["pos"] = params.GetPos()
	return ret
}

type WMTS100CapabilitiesRequest struct {
	WMTSRequest
	CapabilitiesTemplate string
	MimeType             string
}

func NewWMTS100CapabilitiesRequest(param interface{}, url string, validate bool, ht *http.Request) *WMTS100CapabilitiesRequest {
	req := &WMTS100CapabilitiesRequest{}
	req.init(param, url, validate, ht)
	return req
}

func (r *WMTS100CapabilitiesRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "capabilities"
	r.CapabilitiesTemplate = "wmts100capabilities.xml"
	r.MimeType = "text/xml"
	r.FixedParams = map[string]string{}
}

func parseWMTSRequestType(req *http.Request) string {
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
	req_type := parseWMTSRequestType(req)
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

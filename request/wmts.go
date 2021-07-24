package request

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	mapset "github.com/deckarep/golang-set"
	vec2d "github.com/flywave/go3d/float64/vec2"

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
	minx := strconv.FormatFloat(bbox.Min[0], 'E', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'E', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'E', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'E', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *WMTSTileRequestParams) GetSize() [2]int {
	if v, ok := r.params.Get("size"); !ok {
		return [2]int{-1, -1}
	} else {
		if len(v) == 2 {
			si := [2]int{}
			for i := range v {
				v, err := strconv.ParseInt(v[i], 10, 64)
				if err != nil {
					return si
				}
				si[i] = int(v)
			}
			return si
		} else if len(v) == 1 {
			bstr := strings.Split(v[0], ",")
			if len(bstr) == 2 {
				si := [2]int{}
				for i := range bstr {
					v, err := strconv.ParseInt(v[i], 10, 64)
					if err != nil {
						return si
					}
					si[i] = int(v)
				}
				return si
			}
		}
	}
	return [2]int{-1, -1}
}

func (r *WMTSTileRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("size", []string{width, height})
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
	return r.params.GetOne("bboxSR", "EPSG:4326")
}

func (r *WMTSTileRequestParams) SetSrs(srs string) {
	r.params.Set("bboxSR", []string{srs})
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

func (r *WMTSRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

func (r *WMTS100TileRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.FixedParams = map[string]string{"request": "GetTile", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "layer", "style", "tilematrixset",
		"tilematrix", "tilerow", "tilecol", "format"}
}

func (r *WMTS100TileRequest) MakeRequest() map[string]interface{} {
	params := &WMTSTileRequestParams{params: r.Params}
	req := make(map[string]interface{})
	req["layer"] = params.params.GetOne("layer", "")
	req["tilematrixset"] = params.params.GetOne("tilematrixset", "")
	req["format"] = tile.TileFormat(params.params.GetOne("format", ""))
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

func NewWMTSFeatureInfoRequestParams(params RequestParams) WMTSFeatureInfoRequestParams {
	return WMTSFeatureInfoRequestParams{WMTSTileRequestParams: WMTSTileRequestParams{params: params}}
}

func (r *WMTSFeatureInfoRequestParams) GetPos() [2]int {
	i, err := strconv.Atoi(r.params.GetOne("i", "-1"))
	if err != nil {
		return [2]int{-1, -1}
	}
	j, err := strconv.Atoi(r.params.GetOne("j", "-1"))
	if err != nil {
		return [2]int{-1, -1}
	}
	return [2]int{i, j}
}

func (r *WMTSFeatureInfoRequestParams) SetPos(pos [2]float64) {
	r.params.Set("i", []string{strconv.FormatFloat(pos[0], 'E', -1, 64)})
	r.params.Set("j", []string{strconv.FormatFloat(pos[1], 'E', -1, 64)})
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

	ret["infoformat"] = params.params.GetOne("infoformat", "")
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

type WMTSLegendRequestParams struct {
	params RequestParams
}

func NewWMTSLegendRequestParams(params RequestParams) WMTSLegendRequestParams {
	return WMTSLegendRequestParams{params: params}
}

func (r *WMTSLegendRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *WMTSLegendRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMTSLegendRequestParams) SetScale(si int) {
	scale := strconv.FormatInt(int64(si), 10)
	r.params.Set("scale", []string{scale})
}

func (r *WMTSLegendRequestParams) GetScale() int {
	if v, ok := r.params.Get("scale"); !ok {
		return -1
	} else {
		vv, err := strconv.ParseInt(v[0], 10, 64)
		if err != nil {
			return -1
		}
		return int(vv)
	}
}

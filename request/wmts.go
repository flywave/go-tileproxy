package request

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	mapset "github.com/deckarep/golang-set"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type WMTSTileRequestParams struct {
	params RequestParams
}

func NewWMTSTileRequestParams(params RequestParams) WMTSTileRequestParams {
	return WMTSTileRequestParams{params: params}
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

func (r *WMTSTileRequestParams) SetTileMatrixSet(tms string) {
	r.params.Set("tilematrixset", []string{tms})
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
			dimensions[strings.ToUpper(key)] = value
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

func (r *WMTS100TileRequest) GetFormat() *tile.TileFormat {
	p := r.GetRequestParams()
	f := p.GetFormat()
	return &f
}

func (r *WMTS100TileRequest) GetTile() [3]int {
	p := r.GetRequestParams()
	return p.GetCoord()
}

func (r *WMTS100TileRequest) GetOriginString() string {
	return "nw"
}

func (r *WMTS100TileRequest) GetOrigin() geo.OriginType {
	return geo.ORIGIN_NW
}

func (r *WMTSRequest) GetRequestParams() WMTSTileRequestParams {
	return WMTSTileRequestParams{params: r.Params}
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

func (r *WMTS100FeatureInfoRequest) GetRequestParams() WMTSFeatureInfoRequestParams {
	return WMTSFeatureInfoRequestParams{WMTSTileRequestParams: WMTSTileRequestParams{params: r.Params}}
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

const (
	UpdateSequenceNone   = "none"
	UpdateSequenceAny    = "any"
	UpdateSequenceEqual  = "equal"
	UpdateSequenceLower  = "lower"
	UpdateSequenceHigher = "higher"
)

type WMTSCapabilitiesRequestParams struct {
	params RequestParams
}

func NewWMTSCapabilitiesRequestParams(params RequestParams) WMTSCapabilitiesRequestParams {
	return WMTSCapabilitiesRequestParams{params: params}
}

func (r *WMTSCapabilitiesRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *WMTSCapabilitiesRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMTSCapabilitiesRequestParams) GetFormatMimeType() string {
	return r.params.GetOne("format", "")
}

func (r *WMTSCapabilitiesRequestParams) SetUpdateSequence(s string) {
	r.params.Set("updatesequence", []string{s})
}

func (r *WMTSCapabilitiesRequestParams) GetUpdateSequence() string {
	return r.params.GetOne("updatesequence", "")
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

func (r *WMTS100CapabilitiesRequest) GetRequestParams() WMTSCapabilitiesRequestParams {
	return WMTSCapabilitiesRequestParams{params: r.Params}
}

func (r *WMTS100CapabilitiesRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "capabilities"
	r.FixedParams = map[string]string{"request": "GetCapabilities", "version": "1.0.0", "service": "WMTS"}
	r.ExpectedParam = []string{"version", "request", "format", "updateSequence"}
	r.CapabilitiesTemplate = "wmts100capabilities.xml"
	r.MimeType = "text/xml"
}

func parseWMTSRequestType(req *http.Request) (string, RequestParams) {
	values, _ := url.ParseQuery(req.URL.RawQuery)
	if _, ok := values["request"]; ok {
		request_type := strings.ToLower(values["request"][0])
		if strings.HasPrefix(request_type, "get") {
			request_type = request_type[3:]
			if utils.ContainsString([]string{"tile", "featureinfo", "capabilities"}, request_type) {
				return request_type, NewRequestParams(values)
			}
		}
	}
	return "", nil
}

func MakeWMTSRequest(req *http.Request, validate bool) Request {
	req_type, values := parseWMTSRequestType(req)

	switch req_type {
	case "featureinfo":
		r := &WMTS100FeatureInfoRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	case "tile":
		r := &WMTS100TileRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	case "capabilities":
		r := &WMTS100CapabilitiesRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	}
	return nil
}

type URLTemplateConverter struct {
	Template   string
	varregexp  *regexp.Regexp
	variables  map[string]string
	required   []string
	found      []string
	dimensions []string
	_regexp    *regexp.Regexp
}

func NewURLTemplateConverter(template string) *URLTemplateConverter {
	tpl := &URLTemplateConverter{Template: template, _regexp: nil}
	tpl.varregexp = regexp.MustCompile(`(?:\\{)?\\{(\w+)\\}(?:\\})?`)
	tpl.variables = map[string]string{
		"TileMatrixSet": `[\w_.:-]+`,
		"TileMatrix":    `\d+`,
		"TileRow":       `-?\d+`,
		"TileCol":       `-?\d+`,
		"I":             `\d+`,
		"J":             `\d+`,
		"Style":         `[\w_.:-]+`,
		"Layer":         `[\w_.:-]+`,
		"Format":        `\w+`,
		"InfoFormat":    `\w+`,
	}
	tpl.required = []string{
		"TileCol", "TileRow", "TileMatrix", "TileMatrixSet", "Layer",
	}
	tpl.regexp()
	return tpl
}

func (t *URLTemplateConverter) GetDimensions() []string {
	return t.dimensions
}

func (t *URLTemplateConverter) substitute_var(match string) string {
	var var_type_re string
	if var_, ok := t.variables[match]; ok {
		var_type_re = var_
	} else {
		t.dimensions = append(t.dimensions, var_)
		var_type_re = `[\w_.,:-]+`
	}

	t.found = append(t.found, match)
	return fmt.Sprintf("(?P<%s>%s)", strings.ToLower(match), var_type_re)
}

func (t *URLTemplateConverter) regexp() *regexp.Regexp {
	if t._regexp != nil {
		return t._regexp
	}
	converted_re := []string{}
	match := t.varregexp.FindStringSubmatch(t.Template)
	for i := range match {
		if match[i] != "" {
			converted_re = append(converted_re, t.substitute_var(match[i]))
		}
	}

	wmts_re := regexp.MustCompile("/wmts" + strings.Join(converted_re, ""))

	t._regexp = wmts_re
	return wmts_re
}

func NewFeatureInfoURLTemplateConverter(template string) *URLTemplateConverter {
	tpl := &URLTemplateConverter{Template: template, _regexp: nil}
	tpl.varregexp = regexp.MustCompile(`(?:\\{)?\\{(\w+)\\}(?:\\})?`)
	tpl.variables = map[string]string{
		"TileMatrixSet": `[\w_.:-]+`,
		"TileMatrix":    `\d+`,
		"TileRow":       `-?\d+`,
		"TileCol":       `-?\d+`,
		"I":             `\d+`,
		"J":             `\d+`,
		"Style":         `[\w_.:-]+`,
		"Layer":         `[\w_.:-]+`,
		"Format":        `\w+`,
		"InfoFormat":    `\w+`,
	}
	tpl.required = []string{
		"TileCol", "TileRow", "TileMatrix", "TileMatrixSet", "Layer", "I", "J",
	}
	tpl.regexp()
	return tpl
}

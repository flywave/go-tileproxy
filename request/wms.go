package request

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/lucasb-eyer/go-colorful"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

var axisOrderNECache sync.Map

// srsAxisOrderNE reports whether the SRS uses northing/easting axis order.
// Results are cached because building a proj definition is expensive.
func srsAxisOrderNE(srs string) bool {
	if srs == "" {
		return false
	}
	if v, ok := axisOrderNECache.Load(srs); ok {
		return v.(bool)
	}
	ne := false
	if prj := geo.NewProj(srs); prj != nil {
		if sp, ok := prj.(*geo.SRSProj4); !ok || sp != nil {
			ne = prj.IsAxisOrderNE()
		}
	}
	axisOrderNECache.Store(srs, ne)
	return ne
}

func SwitchBBoxEpsgAxisOrder(rect vec2d.Rect, srs string) vec2d.Rect {
	bbox := *rect.Array()
	if bbox != [4]float64{0, 0, 0, 0} && srs != "" && srsAxisOrderNE(srs) {
		return vec2d.Rect{Min: vec2d.T{bbox[1], bbox[0]}, Max: vec2d.T{bbox[3], bbox[2]}}
	}
	return rect
}

// splitLayerList splits a comma separated WMS parameter value such as
// LAYERS=a,b into individual names.
func splitLayerList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// normalizeWMSParams converts wire parameters into the internal form.
// WMS 1.3.0 clients pass the SRS via "crs" and send the bbox in the axis
// order defined by the CRS (northing first for EPSG:4326); older versions
// always send easting first. The bbox is normalized to easting/northing.
func normalizeWMSParams(params RequestParams) {
	if _, ok := params.Get("crs"); !ok {
		return
	}
	crs := params.GetOne("crs", "")
	if crs == "" {
		return
	}
	p := WMSMapRequestParams{params: params}
	p.SetBBox(SwitchBBoxEpsgAxisOrder(p.GetBBox(), crs))
}

type WMSMapRequestParams struct {
	params RequestParams
}

func NewWMSMapRequestParams(params RequestParams) WMSMapRequestParams {
	return WMSMapRequestParams{params: params}
}

func (r *WMSMapRequestParams) Update(params map[string]string) {
	for key, value := range params {
		if _, ok := r.params[key]; !ok {
			r.params[key] = []string{value}
		} else {
			r.params[key] = append(r.params[key], value)
		}
	}
}

func (r *WMSMapRequestParams) GetBBox() vec2d.Rect {
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

func (r *WMSMapRequestParams) SetBBox(bbox vec2d.Rect) {
	minx := strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'f', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'f', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *WMSMapRequestParams) GetLayers() []string {
	l, ok := r.params.Get("layers")
	if ok {
		return splitLayerList(l)
	}
	return []string{}
}

func (r *WMSMapRequestParams) AddLayer(layer string) {
	l, ok := r.params.Get("layers")
	if ok {
		l = append(l, layer)
		r.params.Set("layers", l)
		return
	}
	r.params.Set("layers", []string{layer})
}

func (r *WMSMapRequestParams) AddLayers(layers []string) {
	r.params.Set("layers", layers)
}

func (r *WMSMapRequestParams) GetSize() [2]uint32 {
	w, okw := r.params.Get("width")
	h, okh := r.params.Get("height")
	if !okw || !okh {
		return [2]uint32{0, 0}
	} else {
		ws, err := strconv.ParseInt(w[0], 10, 64)
		if err != nil {
			return [2]uint32{0, 0}
		}
		hs, err := strconv.ParseInt(h[0], 10, 64)
		if err != nil {
			return [2]uint32{0, 0}
		}
		return [2]uint32{uint32(ws), uint32(hs)}
	}
}

func (r *WMSMapRequestParams) GetMetaSize() [2]uint32 {
	w, okw := r.params.Get("metax")
	h, okh := r.params.Get("metay")
	if !okw || !okh {
		return [2]uint32{1, 1}
	} else {
		ws, err := strconv.ParseInt(w[0], 10, 64)
		if err != nil {
			return [2]uint32{1, 1}
		}
		hs, err := strconv.ParseInt(h[0], 10, 64)
		if err != nil {
			return [2]uint32{1, 1}
		}
		return [2]uint32{uint32(ws), uint32(hs)}
	}
}

func (r *WMSMapRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("width", []string{width})
	r.params.Set("height", []string{height})
}

func (r *WMSMapRequestParams) GetCrs() string {
	if v, ok := r.params.Get("crs"); ok && len(v) > 0 && v[0] != "" {
		return v[0]
	}
	// tolerate WMS 1.1.1 style requests that pass the SRS via "srs"
	if v, ok := r.params.Get("srs"); ok && len(v) > 0 {
		return v[0]
	}
	return "EPSG:4326"
}

func (r *WMSMapRequestParams) SetCrs(srs string) {
	r.params.Set("crs", []string{srs})
}

func (r *WMSMapRequestParams) GetSrs() string {
	return r.params.GetOne("srs", "")
}

func (r *WMSMapRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *WMSMapRequestParams) GetFormatString() string {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return strs[1]
}

func (r *WMSMapRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMSMapRequestParams) GetFormatMimeType() string {
	return r.params.GetOne("format", "")
}

func (r *WMSMapRequestParams) GetTransparent() bool {
	str := r.params.GetOne("transparent", "false")
	return strings.ToLower(str) == "true"
}

func (r *WMSMapRequestParams) SetTransparent(transparent bool) {
	if transparent {
		r.params.Set("transparent", []string{"true"})
	} else {
		r.params.Set("transparent", []string{"false"})
	}
}

func (r *WMSMapRequestParams) GetBGColor() color.Color {
	str := r.params.GetOne("bgcolor", "#ffffff")
	c := utils.HexColor(str)
	return &c
}

func (r *WMSMapRequestParams) SetBGColor(c color.Color) {
	cc, _ := colorful.MakeColor(c)
	r.params.Set("bgcolor", []string{cc.Hex()})
}

type WMSRequest struct {
	BaseRequest
	RequestHandlerName string
	ExpectedParam      []string
	NonStrict          bool
	NonStrictParams    mapset.Set
	v                  *Version
}

func (r *WMSRequest) GetRequestHandler() string {
	return r.RequestHandlerName
}

type Version struct {
	version [3]int
}

func NewVersion(ver string) *Version {
	vers := strings.Split(ver, ".")
	if len(vers) != 3 {
		return nil
	}
	v1, err1 := strconv.Atoi(vers[0])
	v2, err2 := strconv.Atoi(vers[1])
	v3, err3 := strconv.Atoi(vers[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}
	return &Version{version: [3]int{v1, v2, v3}}
}

func (v1 *Version) Less(v2 *Version) bool {
	if v1.version[0] > v2.version[0] {
		return false
	} else if v1.version[0] < v2.version[0] {
		return true
	}
	if v1.version[1] > v2.version[1] {
		return false
	} else if v1.version[1] < v2.version[1] {
		return true
	}
	if v1.version[2] >= v2.version[2] {
		return false
	}
	return true
}

type WMSMapRequest struct {
	WMSRequest
}

func NewWMSMapRequest(param interface{}, url string, validate bool, ht *http.Request, nonStrict bool) *WMSMapRequest {
	v := NewVersion("1.3.0")
	req := &WMSMapRequest{WMSRequest{NonStrict: nonStrict, v: v}}
	req.init(param, url, validate, ht)
	return req
}

func (r *WMSMapRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "map"
	r.FixedParams = make(map[string]string)
	r.FixedParams["request"] = "GetMap"
	r.FixedParams["version"] = "1.3.0"
	r.FixedParams["service"] = "WMS"
	r.FixedParams["styles"] = ""
	r.ExpectedParam = []string{"version", "request", "layers", "styles", "srs", "bbox",
		"width", "height", "format"}
}

func (r *WMSMapRequest) AdaptToWMS111() {
	// WMS 1.1.1 passes the SRS via "srs" and always uses easting/northing
	// order, which is the internal order already, so only the name changes.
	if _, ok := r.Params["CRS"]; ok {
		r.Params["SRS"] = r.Params["CRS"]
		delete(r.Params, "CRS")
	}
}

func (r *WMSMapRequest) AdaptParamsToVersion() {
	// convert to the WMS 1.3.0 wire format: rename "srs" to "crs" and switch
	// the bbox to the axis order defined by the CRS
	if _, ok := r.Params["SRS"]; ok {
		r.Params["CRS"] = r.Params["SRS"]
		delete(r.Params, "SRS")
	}
	p := r.GetRequestParams()
	p.SetBBox(SwitchBBoxEpsgAxisOrder(p.GetBBox(), p.GetCrs()))
}

func (r *WMSMapRequest) GetRequestParams() *WMSMapRequestParams {
	return &WMSMapRequestParams{params: r.Params}
}

func (s *WMSMapRequest) Validate() error {
	if err := s.ValidateParam(); err != nil {
		return err
	}
	if err := s.ValidateBBox(); err != nil {
		return err
	}
	if err := s.ValidateStyles(); err != nil {
		return err
	}
	return nil
}

func (s *WMSMapRequest) ValidateParam() error {
	missing_param := []string{}
	for _, param := range s.ExpectedParam {
		if s.NonStrict && s.NonStrictParams.Contains(param) {
			continue
		}
		if _, ok := s.Params[param]; !ok {
			missing_param = append(missing_param, param)
		}
	}
	if len(missing_param) > 0 {
		if utils.ContainsString(missing_param, "format") {
			s.Params["format"] = []string{"image/png"}
		}
		return errors.New("missing parameters " + strings.Join(missing_param, ","))
	}
	return nil
}

func (s *WMSMapRequest) ValidateBBox() error {
	params := &WMSMapRequestParams{params: s.Params}
	bbox := params.GetBBox()
	x0, y0, x1, y1 := bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]
	if x0 >= x1 || y0 >= y1 {
		return fmt.Errorf("invalid bbox [%f %f %f %f]", x0, y0, x1, y1)
	}
	return nil
}

func (s *WMSMapRequest) ValidateStyles() error {
	if _, ok := s.Params["styles"]; ok {
		styles := s.Params["styles"][0]
		strs := strings.Split(styles, ",")
		set := mapset.NewSet()
		for _, sr := range strs {
			set.Add(sr)
		}
		subset := mapset.NewSet("default", "", "inspire_common:DEFAULT")
		if !set.IsSubset(subset) {
			return errors.New("unsupported styles: " + styles)
		}
	}
	return nil
}

func (s *WMSMapRequest) ValidateFormat(image_formats []string) error {
	params := &WMSMapRequestParams{params: s.Params}
	mimetype := params.GetFormatMimeType()
	if mimetype == "" {
		return errors.New("missing format parameter")
	}
	if utils.ContainsString(image_formats, mimetype) {
		return nil
	}
	// tolerate bare format subtypes such as "png" or "jpeg" by matching
	// the subtype against the configured formats
	requested := SplitMimeType(mimetype)
	for _, f := range image_formats {
		if strings.EqualFold(SplitMimeType(f)[1], requested[1]) {
			return nil
		}
	}
	return errors.New("unsupported image format: " + mimetype)
}

func (s *WMSMapRequest) ValidateSrs(srs []string) error {
	params := &WMSMapRequestParams{params: s.Params}
	ss := strings.ToUpper(params.GetCrs())
	if !utils.ContainsString(srs, ss) {
		return errors.New("unsupported srs: " + ss)
	}
	return nil
}

type WMSLegendGraphicRequestParams struct {
	WMSMapRequestParams
}

func NewWMSLegendGraphicRequestParams(params RequestParams) WMSLegendGraphicRequestParams {
	ret := WMSLegendGraphicRequestParams{WMSMapRequestParams: WMSMapRequestParams{params: params}}
	return ret
}

func (r *WMSLegendGraphicRequestParams) GetLayer() string {
	val, ok := r.params.Get("layer")
	if ok {
		return val[0]
	}
	return ""
}

func (r *WMSLegendGraphicRequestParams) SetLayer(l string) {
	r.params.Set("layer", []string{l})
}

func (r *WMSLegendGraphicRequestParams) GetScale() int {
	val, ok := r.params.Get("scale")
	if ok {
		f, err := strconv.Atoi(val[0])
		if err != nil {
			return 1
		}
		return int(f)
	}
	return 1
}

func (r *WMSLegendGraphicRequestParams) SetScale(l int) {
	r.params.Set("scale", []string{strconv.Itoa(l)})
}

type WMSFeatureInfoRequestParams struct {
	WMSMapRequestParams
}

func NewWMSFeatureInfoRequestParams(params RequestParams) WMSFeatureInfoRequestParams {
	return WMSFeatureInfoRequestParams{WMSMapRequestParams: WMSMapRequestParams{params: params}}
}

func (r *WMSFeatureInfoRequestParams) GetQueryLayers() []string {
	if l, ok := r.params.Get("query_layers"); ok && len(l) > 0 {
		return splitLayerList(l)
	}
	return r.GetLayers()
}

func (r *WMSFeatureInfoRequestParams) GetPos() [2]float64 {
	// WMS 1.3.0 uses I/J, older versions use X/Y
	i, err := strconv.ParseFloat(r.params.GetOne("i", r.params.GetOne("x", "-1")), 64)
	if err != nil {
		return [2]float64{-1, -1}
	}
	j, err := strconv.ParseFloat(r.params.GetOne("j", r.params.GetOne("y", "-1")), 64)
	if err != nil {
		return [2]float64{-1, -1}
	}
	return [2]float64{i, j}
}

func (r *WMSFeatureInfoRequestParams) SetPos(pos [2]float64) {
	r.params.Set("i", []string{strconv.FormatFloat(pos[0], 'f', -1, 64)})
	r.params.Set("j", []string{strconv.FormatFloat(pos[1], 'f', -1, 64)})
}

func (r *WMSFeatureInfoRequestParams) GetPosCoords() []float64 {
	si := r.GetSize()
	bbox := r.GetBBox()
	pos := r.GetPos()
	return geo.MakeLinTransf(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(si[0]), float64(si[1])}}, bbox)([]float64{float64(pos[0]), float64(pos[1])})
}

type WMSLegendGraphicRequest struct {
	WMSRequest
}

func (r *WMSLegendGraphicRequest) GetRequestParams() *WMSLegendGraphicRequestParams {
	return &WMSLegendGraphicRequestParams{WMSMapRequestParams: WMSMapRequestParams{params: r.Params}}
}

func NewWMSLegendGraphicRequest(param interface{}, url string, validate bool, ht *http.Request, nonStrict bool) *WMSLegendGraphicRequest {
	v := NewVersion("1.3.0")
	req := &WMSLegendGraphicRequest{WMSRequest{RequestHandlerName: "legendgraphic", NonStrict: nonStrict, v: v}}
	req.RequestHandlerName = "legendgraphic"
	req.FixedParams = make(map[string]string)
	req.FixedParams["request"] = "GetLegendGraphic"
	req.FixedParams["sld_version"] = "1.3.0"
	req.FixedParams["service"] = "WMS"
	req.ExpectedParam = []string{"version", "request", "layer", "format", "sld_version"}
	req.init(param, url, validate, ht)
	return req
}

type WMSFeatureInfoRequest struct {
	WMSMapRequest
}

func NewWMSFeatureInfoRequest(param interface{}, url string, validate bool, ht *http.Request, nonStrict bool) *WMSFeatureInfoRequest {
	req := &WMSFeatureInfoRequest{WMSMapRequest: *NewWMSMapRequest(param, url, validate, ht, nonStrict)}
	req.RequestHandlerName = "featureinfo"
	req.FixedParams = make(map[string]string)
	req.FixedParams["request"] = "GetFeatureInfo"
	req.ExpectedParam = []string{"format", "styles", "query_layers", "i", "j"}
	req.NonStrictParams = mapset.NewSet("format", "styles")
	return req
}

func (r *WMSFeatureInfoRequest) AdaptToWMS111() {
	r.WMSMapRequest.AdaptToWMS111()
	if _, ok := r.Params["I"]; ok {
		r.Params["X"] = r.Params["I"]
		delete(r.Params, "I")
	}
	if _, ok := r.Params["J"]; ok {
		r.Params["Y"] = r.Params["J"]
		delete(r.Params, "J")
	}
}

func (r *WMSFeatureInfoRequest) AdaptParamsToVersion() {
	r.WMSMapRequest.AdaptParamsToVersion()
	if _, ok := r.Params["X"]; ok {
		r.Params["I"] = r.Params["X"]
		delete(r.Params, "X")
	}
	if _, ok := r.Params["Y"]; ok {
		r.Params["J"] = r.Params["Y"]
		delete(r.Params, "Y")
	}
}

func (r *WMSFeatureInfoRequest) GetRequestParams() *WMSFeatureInfoRequestParams {
	return &WMSFeatureInfoRequestParams{WMSMapRequestParams: WMSMapRequestParams{params: r.Params}}
}

func (s *WMSFeatureInfoRequest) ValidateFormat(image_formats []string) bool {
	return false
}

const (
	CAPABILITIES_MIME_TYPE_XML  = "text/xml"
	CAPABILITIES_MIME_TYPE_OGC  = "application/vnd.ogc.wms_xml"
	CAPABILITIES_MIME_TYPE_TEXT = "text/plain"
	CAPABILITIES_MIME_TYPE_HTML = "text/html"
)

type WMSCapabilitiesRequestParams struct {
	params RequestParams
}

func NewWMSCapabilitiesRequestParams(params RequestParams) WMSCapabilitiesRequestParams {
	ret := WMSCapabilitiesRequestParams{params: params}
	return ret
}

func (r *WMSCapabilitiesRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *WMSCapabilitiesRequestParams) GetFormatString() string {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return strs[1]
}

func (r *WMSCapabilitiesRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMSCapabilitiesRequestParams) GetFormatMimeType() string {
	return r.params.GetOne("format", "")
}

type WMSCapabilitiesRequest struct {
	WMSRequest
	MimeType string
}

func (r *WMSCapabilitiesRequest) GetRequestParams() *WMSCapabilitiesRequestParams {
	return &WMSCapabilitiesRequestParams{params: r.Params}
}

func NewWMSCapabilitiesRequest(param interface{}, url string, validate bool, ht *http.Request) *WMSCapabilitiesRequest {
	v := NewVersion("1.3.0")
	req := &WMSCapabilitiesRequest{WMSRequest: WMSRequest{NonStrict: false, v: v}, MimeType: "text/xml"}
	req.FixedParams = make(map[string]string)
	req.RequestHandlerName = "capabilities"
	req.FixedParams["request"] = "GetCapabilities"
	req.FixedParams["version"] = "1.3.0"
	req.FixedParams["service"] = "WMS"
	req.ExpectedParam = []string{"format", "namespace", "rootLayer"}
	req.init(param, url, validate, ht)
	return req
}

func parseWMSRequestType(req *http.Request) (string, RequestParams) {
	values, _ := url.ParseQuery(strings.ToLower(req.URL.RawQuery))
	if _, ok := values["request"]; ok {
		request_type := strings.ToLower(values["request"][0])
		values, _ = url.ParseQuery(req.URL.RawQuery)
		if utils.ContainsString([]string{"getmap", "map"}, request_type) {
			return "map", NewRequestParams(values)
		} else if utils.ContainsString([]string{"getfeatureinfo", "feature_info"}, request_type) {
			return "featureinfo", NewRequestParams(values)
		} else if utils.ContainsString([]string{"getcapabilities", "capabilities"}, request_type) {
			return "capabilities", NewRequestParams(values)
		} else if request_type == "getlegendgraphic" {
			return "legendgraphic", NewRequestParams(values)
		} else {
			return request_type, NewRequestParams(values)
		}
	} else {
		return "", nil
	}
}

func MakeWMSRequest(req *http.Request, validate bool) Request {
	req_type, values := parseWMSRequestType(req)
	if values == nil {
		return nil
	}
	normalizeWMSParams(values)
	switch req_type {
	case "featureinfo":
		r := &WMSFeatureInfoRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	case "map":
		r := &WMSMapRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	case "capabilities":
		r := &WMSCapabilitiesRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	case "legendgraphic":
		r := &WMSLegendGraphicRequest{}
		r.init(values, req.URL.String(), validate, req)
		return r
	}
	return nil
}

var (
	info_formats = map[string]string{
		"text": "text/plain",
		"html": "text/html",
		"xml":  "text/xml",
		"json": "application/json",
	}
)

func InfotypeFromMimetype(mime_type string) string {
	for t, m := range info_formats {
		if m == mime_type {
			return t
		}
	}
	return "text"
}

func MimetypeFromInfotype(info_type string) string {
	for t, m := range info_formats {
		if t == info_type {
			return m
		}
	}
	return "text/plain"
}

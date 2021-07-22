package request

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	mapset "github.com/deckarep/golang-set"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/utils"
)

func switchBBoxEpsgAxisOrder(rect vec2d.Rect, srs string) vec2d.Rect {
	bbox := *rect.Array()
	if bbox != [4]float64{0, 0, 0, 0} && srs != "" {
		prj := geo.NewSRSProj4(srs)
		if prj.IsAxisOrderNE() {
			return vec2d.Rect{Min: vec2d.T{bbox[1], bbox[0]}, Max: vec2d.T{bbox[3], bbox[2]}}
		}
	}
	return rect
}

type WMSMapRequestParams struct {
	params RequestParams
}

func NewWMSMapRequestParams(params RequestParams) WMSMapRequestParams {
	ret := WMSMapRequestParams{params: params}
	ret.switchBBox()
	return ret
}

func (r *WMSMapRequestParams) switchBBox() {
	r.SetBBox(switchBBoxEpsgAxisOrder(r.GetBBox(), r.GetSrs()))
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
	minx := strconv.FormatFloat(bbox.Min[0], 'E', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'E', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'E', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'E', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *WMSMapRequestParams) GetLayers() []string {
	val, ok := r.params.Get("layer")
	if ok {
		layers := []string{}
		for i := range val {
			tmps := strings.Split(layers[i], ",")
			for _, l := range tmps {
				layers = append(layers, l)
			}
		}
		return layers
	}
	return nil
}

func (r *WMSMapRequestParams) SetLayers(l []string) {
	r.params.Set("layer", []string{strings.Join(l, ",")})
}

func (r *WMSMapRequestParams) GetSize() [2]int {
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

func (r *WMSMapRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("size", []string{width, height})
}

func (r *WMSMapRequestParams) GetSrs() string {
	return r.params.GetOne("bboxSR", "EPSG:4326")
}

func (r *WMSMapRequestParams) SetSrs(srs string) {
	r.params.Set("bboxSR", []string{srs})
}

func (r *WMSMapRequestParams) GetFormat() images.ImageFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return images.ImageFormat(strs[1])
}

func (r *WMSMapRequestParams) SetFormat(fmrt images.ImageFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *WMSMapRequestParams) GetFormatMimeType() string {
	return r.params.GetOne("format", "")
}

func (r *WMSMapRequestParams) GetTransparent() bool {
	str := r.params.GetOne("format", "false")
	if strings.ToLower(str) == "true" {
		return true
	}
	return false
}

func (r *WMSMapRequestParams) SetTransparent(transparent bool) {
	if transparent {
		r.params.Set("format", []string{"true"})
	} else {
		r.params.Set("format", []string{"false"})
	}
}

func (r *WMSMapRequestParams) GetBGColor() color.Color {
	str := r.params.GetOne("bgcolor", "#ffffff")
	return utils.ColorRGBFromHex(str)
}

type WMSRequest struct {
	BaseRequest
	RequestHandlerName string
	FixedParams        map[string]string
	ExpectedParam      []string
	NonStrict          bool
	NonStrictParams    mapset.Set
	v                  *Version
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

func NewWMSMapRequest(nonStrict bool) *WMSMapRequest {
	v := NewVersion("1.3.0")
	req := &WMSMapRequest{WMSRequest{RequestHandlerName: "map", NonStrict: nonStrict, v: v}}
	req.FixedParams["request"] = "GetMap"
	req.FixedParams["version"] = "1.3.0"
	req.FixedParams["service"] = "WMS"
	req.ExpectedParam = []string{"version", "request", "layers", "styles", "srs", "bbox",
		"width", "height", "format"}
	return req
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
		return errors.New(fmt.Sprintf("invalid bbox [%f %f %f %f]", x0, y0, x1, y1))
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
	format := params.GetFormat()
	if utils.ContainsString(image_formats, string(format)) {
		params.SetFormat("image/png")
		return errors.New("unsupported image format: " + string(format))
	}
	return nil
}

func (s *WMSMapRequest) ValidateSrs(srs []string) error {
	params := &WMSMapRequestParams{params: s.Params}
	ss := strings.ToUpper(params.GetSrs())
	if utils.ContainsString(srs, ss) {
		return errors.New("unsupported srs: " + ss)
	}
	return nil
}

type WMSLegendGraphicRequestParams struct {
	WMSMapRequestParams
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

func (r *WMSLegendGraphicRequestParams) GetScale() float64 {
	val, ok := r.params.Get("scale")
	if ok {
		f, err := strconv.ParseFloat(val[0], 64)
		if err != nil {
			return 1
		}
		return f
	}
	return 1
}

func (r *WMSLegendGraphicRequestParams) SetScale(l float64) {
	r.params.Set("scale", []string{strconv.FormatFloat(l, 'g', 1, 64)})
}

type WMSFeatureInfoRequestParams struct {
	WMSMapRequestParams
}

func NewWMSFeatureInfoRequestParams(params RequestParams) WMSFeatureInfoRequestParams {
	return WMSFeatureInfoRequestParams{WMSMapRequestParams: WMSMapRequestParams{params: params}}
}

func (r *WMSFeatureInfoRequestParams) GetPos() [2]int {
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

func (r *WMSFeatureInfoRequestParams) SetPos(pos [2]float64) {
	r.params.Set("i", []string{strconv.FormatFloat(pos[0], 'E', -1, 64)})
	r.params.Set("j", []string{strconv.FormatFloat(pos[1], 'E', -1, 64)})
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

func NewWMSLegendGraphicRequest(nonStrict bool) *WMSLegendGraphicRequest {
	v := NewVersion("1.3.0")
	req := &WMSLegendGraphicRequest{WMSRequest{RequestHandlerName: "legendgraphic", NonStrict: nonStrict, v: v}}
	req.FixedParams["request"] = "GetLegendGraphic"
	req.FixedParams["sld_version"] = "1.3.0"
	req.FixedParams["service"] = "WMS"
	req.ExpectedParam = []string{"version", "request", "layer", "format", "sld_version"}
	return req
}

type WMSFeatureInfoRequest struct {
	WMSMapRequest
}

func NewWMSFeatureInfoRequest(nonStrict bool) *WMSFeatureInfoRequest {
	req := &WMSFeatureInfoRequest{WMSMapRequest: *NewWMSMapRequest(nonStrict)}
	req.RequestHandlerName = "featureinfo"
	req.FixedParams["request"] = "GetFeatureInfo"
	req.ExpectedParam = []string{"format", "styles", "query_layers", "i", "j"}
	req.NonStrictParams = mapset.NewSet("format", "styles")
	return req
}

func (s *WMSFeatureInfoRequest) ValidateFormat(image_formats []string) bool {
	return false
}

type WMSCapabilitiesRequest struct {
	WMSRequest
	MimeType string
}

func NewWMSCapabilitiesRequest() *WMSCapabilitiesRequest {
	v := NewVersion("1.3.0")
	req := &WMSCapabilitiesRequest{WMSRequest: WMSRequest{RequestHandlerName: "capabilities", NonStrict: false, v: v}, MimeType: "text/xml"}
	req.FixedParams["request"] = "GetCapabilities"
	req.FixedParams["version"] = "1.3.0"
	req.FixedParams["service"] = "WMS"
	return req
}

func parseWMSRequestType(req *http.Request) string {
	if _, ok := req.Header["request"]; ok {
		request_type := strings.ToLower(req.Header["request"][0])
		if utils.ContainsString([]string{"getmap", "map"}, request_type) {
			return "map"
		} else if utils.ContainsString([]string{"getfeatureinfo", "feature_info"}, request_type) {
			return "featureinfo"
		} else if utils.ContainsString([]string{"getcapabilities", "capabilities"}, request_type) {
			return "capabilities"
		} else if request_type == "getlegendgraphic" {
			return "legendgraphic"
		} else {
			return request_type
		}
	} else {
		return ""
	}
}

func MakeWMSRequest(req *http.Request, validate bool) Request {
	req_type := parseWMSRequestType(req)
	switch req_type {
	case "featureinfo":
		r := &WMSFeatureInfoRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	case "map":
		r := &WMSMapRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	case "capabilities":
		r := &WMSCapabilitiesRequest{}
		r.init(req.Header, req.URL.String(), true, req)
		return r
	case "legendgraphic":
		r := &WMSLegendGraphicRequest{}
		r.init(req.Header, req.URL.String(), true, req)
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

func infotype_from_mimetype(mime_type string) string {
	for t, m := range info_formats {
		if m == mime_type {
			return t
		}
	}
	return "text"
}

func mimetype_from_infotype(info_type string) string {
	for t, m := range info_formats {
		if t == info_type {
			return m
		}
	}
	return "text/plain"
}

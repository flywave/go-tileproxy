package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type ArcGISExportRequestParams struct {
	params RequestParams
}

func NewArcGISExportRequestParams(params RequestParams) ArcGISExportRequestParams {
	return ArcGISExportRequestParams{params: params}
}

func (r *ArcGISExportRequestParams) GetFormat() string {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return strs[1]
}

func (r *ArcGISExportRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *ArcGISExportRequestParams) GetBBox() vec2d.Rect {
	if v, ok := r.params.Get("bbox"); !ok {
		return vec2d.Rect{}
	} else {
		if len(v) == 4 {
			bbox := [4]float64{}
			for i := range v {
				v_, err := strconv.ParseFloat(v[i], 64)
				if err != nil {
					return vec2d.Rect{}
				}
				bbox[i] = v_
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

func (r *ArcGISExportRequestParams) SetBBox(bbox vec2d.Rect) {
	minx := strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'f', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'f', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *ArcGISExportRequestParams) GetSize() [2]uint32 {
	if v, ok := r.params.Get("size"); !ok {
		return [2]uint32{0, 0}
	} else {
		if len(v) >= 2 {
			si := [2]uint32{0, 0}
			for i := range v[:2] {
				v, err := strconv.ParseInt(v[i], 10, 64)
				if err != nil {
					return si
				}
				si[i] = uint32(v)
			}
			return si
		} else if len(v) == 1 {
			bstr := strings.Split(v[0], ",")
			if len(bstr) == 2 {
				si := [2]uint32{0, 0}
				for i := range bstr {
					v, err := strconv.ParseInt(v[i], 10, 64)
					if err != nil {
						return si
					}
					si[i] = uint32(v)
				}
				return si
			}
		}
	}
	return [2]uint32{0, 0}
}

func (r *ArcGISExportRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("size", []string{width, height})
}

func (r *ArcGISExportRequestParams) GetLayers() []string {
	l, ok := r.params.Get("layers")
	if ok {
		return l
	}
	return []string{}
}

func (r *ArcGISExportRequestParams) AddLayer(layer string) {
	l, ok := r.params.Get("layers")
	if ok {
		l = append(l, layer)
		r.params.Set("layers", l)
		return
	}
	r.params.Set("layers", []string{layer})
}

func (r *ArcGISExportRequestParams) AddLayers(layers []string) {
	r.params.Set("layers", layers)
}

func (r *ArcGISExportRequestParams) GetBBOxSrs() string {
	return fmt.Sprintf("EPSG:%s", r.params.GetOne("bboxSR", "4326"))
}

func (r *ArcGISExportRequestParams) SetBBoxSrs(srs string) {
	epsg := geo.GetEpsgNum(srs)
	r.params.Set("bboxSR", []string{strconv.FormatInt(int64(epsg), 10)})
}

func (r *ArcGISExportRequestParams) GetImageSrs() string {
	return fmt.Sprintf("EPSG:%s", r.params.GetOne("imageSR", "4326"))
}

func (r *ArcGISExportRequestParams) SetImageSrs(srs string) {
	epsg := geo.GetEpsgNum(srs)
	r.params.Set("imageSR", []string{strconv.FormatInt(int64(epsg), 10)})
}

func (r *ArcGISExportRequestParams) GetTransparent() bool {
	str := r.params.GetOne("format", "false")
	if strings.ToLower(str) == "true" {
		return true
	}
	return false
}

func (r *ArcGISExportRequestParams) SetTransparent(b bool) {
	if b {
		r.params.Set("transparent", []string{"true"})
	} else {
		r.params.Set("transparent", []string{"false"})
	}
}

func (r *ArcGISExportRequestParams) GetDpi() int {
	str := r.params.GetOne("dpi", "-1")
	dpi, _ := strconv.Atoi(str)
	return dpi
}

func (r *ArcGISExportRequestParams) SetDpi(dpi int) {
	r.params.Set("dpi", []string{strconv.Itoa(dpi)})
}

type ArcGISIdentifyRequestParams struct {
	params RequestParams
}

func NewArcGISIdentifyRequestParams(params RequestParams) ArcGISIdentifyRequestParams {
	return ArcGISIdentifyRequestParams{params: params}
}

func (r *ArcGISIdentifyRequestParams) GetFormat() tile.TileFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return tile.TileFormat(strs[1])
}

func (r *ArcGISIdentifyRequestParams) SetFormat(fmrt tile.TileFormat) {
	r.params.Set("format", []string{fmrt.MimeType()})
}

func (r *ArcGISIdentifyRequestParams) GetBBox() vec2d.Rect {
	if v, ok := r.params.Get("mapExtent"); !ok {
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

func (r *ArcGISIdentifyRequestParams) SetBBox(bbox vec2d.Rect) {
	minx := strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'f', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'f', -1, 64)
	r.params.Set("mapExtent", []string{minx, miny, maxx, maxy})
}

func (r *ArcGISIdentifyRequestParams) GetSize() [2]uint32 {
	if v, ok := r.params.Get("imageDisplay"); !ok {
		return [2]uint32{0, 0}
	} else {
		if len(v) >= 2 {
			si := [2]uint32{0, 0}
			for i := range v[:2] {
				v, err := strconv.ParseInt(v[i], 10, 64)
				if err != nil {
					return si
				}
				si[i] = uint32(v)
			}
			return si
		} else if len(v) == 1 {
			bstr := strings.Split(v[0], ",")
			if len(bstr) == 2 {
				si := [2]uint32{0, 0}
				for i := range bstr {
					v, err := strconv.ParseInt(v[i], 10, 64)
					if err != nil {
						return si
					}
					si[i] = uint32(v)
				}
				return si
			}
		}
	}
	return [2]uint32{0, 0}
}

func (r *ArcGISIdentifyRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("imageDisplay", []string{width, height, ",96"})
}

func (r *ArcGISIdentifyRequestParams) GetPos() [2]float64 {
	if v, ok := r.params.Get("geometry"); !ok {
		return [2]float64{}
	} else {
		if len(v) == 2 {
			si := [2]float64{}
			for i := range v {
				v, err := strconv.ParseFloat(v[i], 64)
				if err != nil {
					return si
				}
				si[i] = v
			}
			return si
		} else if len(v) == 1 {
			bstr := strings.Split(v[0], ",")
			if len(bstr) == 2 {
				si := [2]float64{}
				for i := range bstr {
					v, err := strconv.ParseFloat(v[i], 64)
					if err != nil {
						return si
					}
					si[i] = v
				}
				return si
			}
		}
	}
	return [2]float64{}
}

func (r *ArcGISIdentifyRequestParams) SetPos(pos [2]float64) {
	posx := strconv.FormatFloat(pos[0], 'f', -1, 64)
	posy := strconv.FormatFloat(pos[1], 'f', -1, 64)
	r.params.Set("geometry", []string{posx, posy})
}

func (r *ArcGISIdentifyRequestParams) GetSrs() string {
	srs := r.params.GetOne("sr", "4326")
	return fmt.Sprintf("EPSG:%s", srs)
}

func (r *ArcGISIdentifyRequestParams) SetSrs(srs string) {
	if strings.Contains(srs, ":") {
		strs := strings.Split(srs, ":")
		r.params.Set("sr", []string{strs[1]})
	} else {
		r.params.Set("sr", []string{srs})
	}
}

func (r *ArcGISIdentifyRequestParams) GetTransparent() bool {
	str := r.params.GetOne("transparent", "false")
	if strings.ToLower(str) == "true" {
		return true
	}
	return false
}

func (r *ArcGISIdentifyRequestParams) SetTransparent(b bool) {
	if b {
		r.params.Set("transparent", []string{"true"})
	} else {
		r.params.Set("transparent", []string{"false"})
	}
}

func (r *ArcGISIdentifyRequestParams) GetLayers() []string {
	l, ok := r.params.Get("layers")
	if ok {
		return l
	}
	return []string{}
}

func (r *ArcGISIdentifyRequestParams) AddLayer(layer string) {
	l, ok := r.params.Get("layers")
	if ok {
		l = append(l, layer)
		r.params.Set("layers", l)
		return
	}
	r.params.Set("layers", []string{layer})
}

func (r *ArcGISIdentifyRequestParams) AddLayers(layers []string) {
	r.params.Set("layers", layers)
}

type ArcGISRequest struct {
	BaseRequest
	Uri *url.URL
}

func NewArcGISRequest(param interface{}, url string) *ArcGISRequest {
	req := &ArcGISRequest{}
	req.init(param, url, false, nil)
	return req
}

func (r *ArcGISRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.FixedParams = map[string]string{"f": "image"}
	r.Uri = rest_endpoint(url)
}

func (r *ArcGISRequest) QueryString() string {
	params := r.Params.copy()
	for key, value := range r.FixedParams {
		params[key] = []string{value}
	}
	return params.QueryString()
}

func (r *ArcGISRequest) GetRequestParams() *ArcGISExportRequestParams {
	return &ArcGISExportRequestParams{params: r.Params}
}

type ArcGISIdentifyRequest struct {
	BaseRequest
	Uri *url.URL
}

func NewArcGISIdentifyRequest(param interface{}, url string) *ArcGISIdentifyRequest {
	req := &ArcGISIdentifyRequest{}
	req.init(param, url, false, nil)
	return req
}

func (r *ArcGISIdentifyRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.FixedParams = map[string]string{"geometryType": "esriGeometryPoint"}
	r.Uri = rest_identify_endpoint(url)
}

func (r *ArcGISIdentifyRequest) QueryString() string {
	params := r.Params.copy()
	for key, value := range r.FixedParams {
		params[key] = []string{value}
	}
	return params.QueryString()
}

func (r *ArcGISIdentifyRequest) GetRequestParams() *ArcGISIdentifyRequestParams {
	return &ArcGISIdentifyRequestParams{params: r.Params}
}

func urlParse(uri string) *url.URL {
	u, _ := url.Parse(uri)
	return u
}

func rest_endpoint(url string) *url.URL {
	parts := urlParse(url)
	ps := strings.TrimPrefix(parts.Path, "/")
	path := strings.Split(ps, "/")

	if path[len(path)-1] == "export" || path[len(path)-1] == "exportImage" {
		if path[len(path)-2] == "MapServer" {
			path[len(path)-1] = "export"
		} else if path[len(path)-2] == "ImageServer" {
			path[len(path)-1] = "exportImage"
		}
	} else if path[len(path)-1] == "MapServer" {
		path = append(path, "export")
	} else if path[len(path)-1] == "ImageServer" {
		path = append(path, "exportImage")
	}
	parts.Path = strings.Join(path, "/")
	return parts
}

func rest_identify_endpoint(url string) *url.URL {
	parts := urlParse(url)
	ps := strings.TrimPrefix(parts.Path, "/")
	path := strings.Split(ps, "/")

	if path[len(path)-1] == "export" || path[len(path)-1] == "exportImage" {
		path[len(path)-1] = "identify"
	} else if path[len(path)-1] == "MapServer" || path[len(path)-1] == "ImageServer" {
		path = append(path, "identify")
	}

	parts.Path = strings.Join(path, "/")
	return parts
}

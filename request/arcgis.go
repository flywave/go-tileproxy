package request

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/images"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type ArcGISExportRequestParams struct {
	params RequestParams
}

func NewArcGISExportRequestParams(params RequestParams) ArcGISExportRequestParams {
	return ArcGISExportRequestParams{params: params}
}

func (r *ArcGISExportRequestParams) GetFormat() images.ImageFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return images.ImageFormat(strs[1])
}

func (r *ArcGISExportRequestParams) SetFormat(fmrt images.ImageFormat) {
	r.params.Set("tilematrix", []string{fmrt.MimeType()})
}

func (r *ArcGISExportRequestParams) GetBBox() vec2d.Rect {
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

func (r *ArcGISExportRequestParams) SetBBox(bbox vec2d.Rect) {
	minx := strconv.FormatFloat(bbox.Min[0], 'E', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'E', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'E', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'E', -1, 64)
	r.params.Set("bbox", []string{minx, miny, maxx, maxy})
}

func (r *ArcGISExportRequestParams) GetSize() [2]int {
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

func (r *ArcGISExportRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("size", []string{width, height})
}

func (r *ArcGISExportRequestParams) GetBBOxSrs() string {
	return r.params.GetOne("bboxSR", "EPSG:4326")
}

func (r *ArcGISExportRequestParams) SetBBOxSrs(srs string) {
	r.params.Set("bboxSR", []string{srs})
}

func (r *ArcGISExportRequestParams) GetImageSrs() string {
	return r.params.GetOne("imageSR", "EPSG:4326")
}

func (r *ArcGISExportRequestParams) SetImageSrs(srs string) {
	r.params.Set("imageSR", []string{srs})
}

func (r *ArcGISExportRequestParams) SetTransparent(b bool) {
	if b {
		r.params.Set("transparent", []string{"true"})
	} else {
		r.params.Set("transparent", []string{"false"})
	}
}

type ArcGISIdentifyRequestParams struct {
	params RequestParams
}

func NewArcGISIdentifyRequestParams(params RequestParams) ArcGISIdentifyRequestParams {
	return ArcGISIdentifyRequestParams{params: params}
}

func (r *ArcGISIdentifyRequestParams) GetFormat() images.ImageFormat {
	strs := SplitMimeType(r.params.GetOne("format", ""))
	return images.ImageFormat(strs[1])
}

func (r *ArcGISIdentifyRequestParams) SetFormat(fmrt images.ImageFormat) {
	r.params.Set("tilematrix", []string{fmrt.MimeType()})
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
	minx := strconv.FormatFloat(bbox.Min[0], 'E', -1, 64)
	miny := strconv.FormatFloat(bbox.Min[1], 'E', -1, 64)
	maxx := strconv.FormatFloat(bbox.Max[0], 'E', -1, 64)
	maxy := strconv.FormatFloat(bbox.Max[1], 'E', -1, 64)
	r.params.Set("mapExtent", []string{minx, miny, maxx, maxy})
}

func (r *ArcGISIdentifyRequestParams) GetSize() [2]int {
	if v, ok := r.params.Get("imageDisplay"); !ok {
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

func (r *ArcGISIdentifyRequestParams) SetSize(si [2]uint32) {
	width := strconv.FormatInt(int64(si[0]), 10)
	height := strconv.FormatInt(int64(si[1]), 10)
	r.params.Set("imageDisplay", []string{width, height})
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
	posx := strconv.FormatFloat(pos[0], 'E', -1, 64)
	posy := strconv.FormatFloat(pos[1], 'E', -1, 64)
	r.params.Set("geometry", []string{posx, posy})
}

func (r *ArcGISIdentifyRequestParams) GetSrs() string {
	srs := r.params.GetOne("sr", "4326")
	return fmt.Sprintf("%s", srs)
}

func (r *ArcGISIdentifyRequestParams) SetSrs(srs string) {
	if strings.Contains(srs, ":") {
		strs := strings.Split(srs, ":")
		r.params.Set("sr", []string{strs[1]})
	} else {
		r.params.Set("sr", []string{srs})
	}
}

func (r *ArcGISIdentifyRequestParams) SetTransparent(b bool) {
	if b {
		r.params.Set("transparent", []string{"true"})
	} else {
		r.params.Set("transparent", []string{"false"})
	}
}

type ArcGISRequest struct {
	BaseRequest
	FixedParams map[string]string
}

func (r *ArcGISRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.FixedParams = map[string]string{"f": "image"}
}

type ArcGISIdentifyRequest struct {
	BaseRequest
	FixedParams map[string]string
}

func (r *ArcGISIdentifyRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.FixedParams = map[string]string{"geometryType": "esriGeometryPoint"}
}

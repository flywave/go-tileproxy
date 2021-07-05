package sources

import (
	"image/color"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type errorInfo struct {
	Color     color.Color
	Cacheable bool
}

type HTTPSourceErrorHandler struct {
	ResponseErrorCodes map[string]errorInfo
}

func (h *HTTPSourceErrorHandler) AddHandler(http_code string, color color.Color, cacheable bool) {
	h.ResponseErrorCodes[http_code] = errorInfo{Color: color, Cacheable: cacheable}
}

func (h *HTTPSourceErrorHandler) Handle(status_code string, query *layer.MapQuery) images.Source {
	var info errorInfo
	if _, ok := h.ResponseErrorCodes[status_code]; ok {
		info = h.ResponseErrorCodes[status_code]
	} else if _, ok := h.ResponseErrorCodes["other"]; ok {
		info = h.ResponseErrorCodes["other"]
	} else {
		return nil
	}

	image_opts := &images.ImageOptions{BgColor: info.Color, Transparent: geo.NewBool(true)}
	img_source := images.NewBlankImageSource(query.Size, image_opts, info.Cacheable)
	return img_source
}

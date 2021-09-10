package sources

import (
	"image/color"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type errorInfo struct {
	Color     color.Color
	Cacheable *tile.CacheInfo
}

type HTTPSourceErrorHandler struct {
	ResponseErrorCodes map[string]errorInfo
}

func (h *HTTPSourceErrorHandler) AddHandler(http_code string, color color.Color, cacheable *tile.CacheInfo) {
	h.ResponseErrorCodes[http_code] = errorInfo{Color: color, Cacheable: cacheable}
}

func (h *HTTPSourceErrorHandler) Handle(status_code string, query *layer.MapQuery) tile.Source {
	var info errorInfo
	if _, ok := h.ResponseErrorCodes[status_code]; ok {
		info = h.ResponseErrorCodes[status_code]
	} else if _, ok := h.ResponseErrorCodes["other"]; ok {
		info = h.ResponseErrorCodes["other"]
	} else {
		return nil
	}

	imageOpts := &imagery.ImageOptions{
		BgColor:     info.Color,
		Transparent: geo.NewBool(true),
	}
	return imagery.NewBlankImageSource(query.Size, imageOpts, info.Cacheable)
}

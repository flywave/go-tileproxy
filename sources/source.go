package sources

import (
	"errors"

	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

var ErrInvalidBBOX = errors.New("Invalid BBOX")

var ErrInvalidSourceQuery = errors.New("Invalid source query")

type Source interface {
}

type ImagerySource interface {
	Source
	GetMap(query *layer.MapQuery) images.Source
}

type InfoSource interface {
	Source
	GetInfo(query *layer.MapQuery)
}

type LegendSource interface {
	Source
	GetLegend(query *layer.MapQuery)
}

type StyleSource interface {
	Source
	GetStyle(query *layer.StyleQuery)
}

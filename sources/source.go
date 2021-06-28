package sources

import (
	"errors"

	"github.com/flywave/go-tileproxy/layer"
)

var ErrInvalidBBOX = errors.New("Invalid BBOX ")

var ErrInvalidSourceQuery = errors.New("Invalid source query ")

type Source interface {
}

type MapSource interface {
	Source
	GetMap(query layer.MapQuery)
}

type InfoSource interface {
	Source
	GetInfo(query layer.MapQuery)
}

type LegendSource interface {
	Source
	GetLegend(query layer.MapQuery)
}

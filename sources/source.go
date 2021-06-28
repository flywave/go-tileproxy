package sources

import "errors"

var ErrInvalidBBOX = errors.New("Invalid BBOX ")

var ErrInvalidSourceQuery = errors.New("Invalid source query ")

type Source interface {
}

type MapSource interface {
	Source
	GetMap(query MapQuery)
}

type InfoSource interface {
	Source
	GetInfo(query MapQuery)
}

type LegendSource interface {
	Source
	GetLegend(query MapQuery)
}

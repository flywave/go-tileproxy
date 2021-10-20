package imports

import "github.com/flywave/go-geo"

type Provider interface {
	GetFormat() string
	GetFormatMimeType() string
	GetGrid() geo.Grid
	GetCoverage() geo.Coverage
}

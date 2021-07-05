package layer

import (
	"github.com/flywave/go-tileproxy/geo"
)

type MapLayer struct {
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
}

type DirectMapLayer struct {
}

type CacheMapLayer struct {
}

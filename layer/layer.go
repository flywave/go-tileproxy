package layer

import (
	"github.com/flywave/go-tileproxy/maths"
)

type MapLayer struct {
	SupportMetaTiles bool
	ResRange         maths.ResolutionRange
	Coverage         maths.Coverage
}

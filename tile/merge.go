package tile

import (
	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
)

type Merger interface {
	AddSource(src Source, cov geo.Coverage)
	Merge(opts TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) Source
}

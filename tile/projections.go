package tile

import "github.com/flywave/go-geom/generic"

const (
	WebMercator = 3857
	WGS84       = 4326
)

var (
	WebMercatorBounds = &generic.Extent{-20026376.39, -20048966.10, 20026376.39, 20048966.10}
	WGS84Bounds       = &generic.Extent{-180.0, -85.0511, 180.0, 85.0511}
)

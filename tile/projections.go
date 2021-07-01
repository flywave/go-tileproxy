package tile

import "github.com/flywave/go-geom/general"

const (
	WebMercator = 3857
	WGS84       = 4326
)

var (
	WebMercatorBounds = &general.Extent{-20026376.39, -20048966.10, 20026376.39, 20048966.10}
	WGS84Bounds       = &general.Extent{-180.0, -85.0511, 180.0, 85.0511}
)

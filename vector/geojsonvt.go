package vector

import (
	"github.com/flywave/go-geom"
	_ "github.com/flywave/go-mbgeom/geojsonvt"
)

type GeoJSONVT geom.FeatureCollection

type GeoJSONVTSource struct {
	VectorSource
}

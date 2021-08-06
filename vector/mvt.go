package vector

import (
	"github.com/flywave/go-geom"
	_ "github.com/flywave/go-mapbox"
	_ "github.com/flywave/go-mbgeom"
)

type PBFProto uint32

const (
	PBF_PTOTO_MAPBOX   = 0
	PBF_PTOTO_LUOKUANG = 1
)

type PBF map[string][]*geom.Feature

type MVTSource struct {
	VectorSource
	Proto PBFProto
}

package tileproxy

import (
	"net/http"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/flywave/go-tileproxy/sources"
)

type ServiceType uint32

const (
	MapboxService   ServiceType = 0
	WMSService      ServiceType = 1
	WMTSService     ServiceType = 2
	WMTSRestService ServiceType = 3
	TileService     ServiceType = 4
)

type DatasetService struct {
	Identifier string
	Service    service.Service
	Caches     map[string]layer.Layer
	Grids      map[string]geo.Grid
	Coverages  map[string]geo.Coverage
	Sources    map[string]sources.Source
}

type TileProxy struct {
	Datasets map[string]DatasetService
}

func NewTileProxy(proxy []*setting.ProxyDataset) *TileProxy {
	return nil
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

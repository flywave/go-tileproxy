package tileproxy

import (
	"net/http"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/flywave/go-tileproxy/sources"
)

type TileProxy struct {
	Caches    map[string]cache.Cache
	Grids     map[string]geo.Grid
	Coverages map[string]geo.Coverage
	Sources   map[string]sources.Source
	Services  map[string]service.Service
}

func NewTileProxy(proxy *setting.ProxyDataset) *TileProxy {
	return nil
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

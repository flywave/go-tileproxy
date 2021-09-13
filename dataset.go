package tileproxy

import (
	"net/http"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/setting"
)

type ServiceType uint32

const (
	MapboxService   ServiceType = 0
	WMSService      ServiceType = 1
	WMTSService     ServiceType = 2
	WMTSRestService ServiceType = 3
	TileService     ServiceType = 4
)

type Dataset struct {
	setting.ProxyInstance
	Identifier    string
	Type          ServiceType
	Service       service.Service
	Grids         map[string]geo.Grid
	Sources       map[string]layer.Layer
	InfoSources   map[string]layer.InfoLayer
	LegendSources map[string]layer.LegendLayer
	Caches        map[string]cache.Manager
}

func NewDataset(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) *Dataset {
	ret := &Dataset{
		Identifier:    dataset.Identifier,
		Grids:         make(map[string]geo.Grid),
		Sources:       make(map[string]layer.Layer),
		Caches:        make(map[string]cache.Manager),
		InfoSources:   make(map[string]layer.InfoLayer),
		LegendSources: make(map[string]layer.LegendLayer),
	}
	ret.load(dataset, basePath, globals)
	return ret
}

func (s *Dataset) load(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) {
	s.loadGrids(dataset, basePath, globals)
	s.loadSources(dataset, basePath, globals)
	s.loadCaches(dataset, basePath, globals)
	s.loadService(dataset, basePath, globals)
}

func (s *Dataset) loadGrids(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) {
	for k, g := range dataset.Grids {
		s.Grids[k] = setting.ConvertGridOpts(&g)
	}
}

func (s *Dataset) loadSources(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) {
	for k, src := range dataset.Sources {
		switch source := src.(type) {
		case *setting.WMSSource:
			if source.Opts.FeatureInfo != nil && *source.Opts.FeatureInfo {
				s.InfoSources[k] = setting.LoadWMSInfoSource(source, basePath, globals)
			} else if source.Opts.LegendGraphic != nil && *source.Opts.LegendGraphic {
				s.LegendSources[k] = setting.LoadWMSLegendsSource(source, globals)
			} else {
				s.Sources[k] = setting.LoadWMSMapSource(source, s, globals)
			}
		case *setting.TileSource:
			s.Sources[k] = setting.LoadTileSource(source, globals, s)
		case *setting.MapboxTileSource:
			s.Sources[k] = setting.LoadMapboxTileSource(source, globals, s)
		case *setting.ArcGISSource:
			if source.Opts.Featureinfo != nil && *source.Opts.Featureinfo {
				s.InfoSources[k] = setting.LoadArcGISInfoSource(source, globals)
			} else {
				s.Sources[k] = setting.LoadArcGISSource(source, s, globals)
			}
		}
	}
}

func (s *Dataset) loadCaches(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) {
	for k, c := range dataset.Caches {
		switch cache := c.(type) {
		case *setting.CacheSource:
			s.Caches[k] = setting.LoadCacheManager(cache, globals, s)
		}
	}
}

func (s *Dataset) loadService(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting) {
	switch srv := dataset.Service.(type) {
	case *setting.TMSService:
		s.Service = setting.LoadTMSService(srv, s)
	case *setting.WMSService:
		s.Service = setting.LoadWMSService(srv, s, globals, basePath)
	case *setting.MapboxService:
		s.Service = setting.LoadMapboxService(srv, globals, s)
	case *setting.WMTSService:
		if srv.Restful != nil && *srv.Restful {
			s.Service = setting.LoadWMTSRestfulService(srv, s)
		} else {
			s.Service = setting.LoadWMTSService(srv, s)
		}
	}
}

func (s *Dataset) Clean() {
}

func (s *Dataset) GetIdentifier() string {
	return s.Identifier
}

func (s *Dataset) GetServiceType() ServiceType {
	return s.Type
}

func (s *Dataset) GetGrid(name string) geo.Grid {
	if g, ok := s.Grids[name]; ok {
		return g
	}
	return nil
}

func (s *Dataset) GetSource(name string) layer.Layer {
	if g, ok := s.Sources[name]; ok {
		return g
	}
	return nil
}

func (s *Dataset) GetCache(name string) cache.Manager {
	if l, ok := s.Caches[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetCacheSource(name string) layer.Layer {
	manager := s.GetCache(name)
	if manager != nil {
		tile_grid := manager.GetGrid()
		sources := manager.GetSources()
		extent := layer.MergeLayerExtents(sources)
		if extent.IsDefault() {
			extent = geo.MapExtentFromGrid(tile_grid)
		}

		cache_extent := geo.MapExtentFromGrid(tile_grid)
		cache_extent = extent.Intersection(cache_extent)

		return cache.NewCacheSource(manager, cache_extent, manager.GetTileOptions(), nil, true)
	}
	return nil
}

func (s *Dataset) GetInfoSource(name string) layer.InfoLayer {
	if l, ok := s.InfoSources[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetLegendSource(name string) layer.LegendLayer {
	if l, ok := s.LegendSources[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetService() service.Service {
	return s.Service
}

func (s *Dataset) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Service != nil {
		s.Service.ServeHTTP(w, r)
	}
}

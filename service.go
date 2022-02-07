package tileproxy

import (
	"net/http"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
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

type Service struct {
	setting.ProxyInstance
	UUID          string
	Type          ServiceType
	Service       service.Service
	Grids         map[string]geo.Grid
	Sources       map[string]layer.Layer
	InfoSources   map[string]layer.InfoLayer
	LegendSources map[string]layer.LegendLayer
	Caches        map[string]cache.Manager
}

func NewService(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting, fac setting.CacheFactory) *Service {
	ret := &Service{
		UUID:          dataset.UUID,
		Grids:         make(map[string]geo.Grid),
		Sources:       make(map[string]layer.Layer),
		Caches:        make(map[string]cache.Manager),
		InfoSources:   make(map[string]layer.InfoLayer),
		LegendSources: make(map[string]layer.LegendLayer),
	}
	ret.load(dataset, basePath, globals, fac)
	return ret
}

func (s *Service) load(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	s.loadGrids(dataset, basePath, globals)
	s.loadSources(dataset, basePath, globals, fac)
	s.loadCaches(dataset, basePath, globals, fac)
	s.loadService(dataset, basePath, globals, fac)
}

func (s *Service) loadGrids(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting) {
	for k, g := range dataset.Grids {
		s.Grids[k] = setting.ConvertGridOpts(&g)
	}
}

func (s *Service) loadSources(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	for k, src := range dataset.Sources {
		switch source := src.(type) {
		case *setting.WMSSource:
			if source.Opts.FeatureInfo != nil && *source.Opts.FeatureInfo {
				s.InfoSources[k] = setting.LoadWMSInfoSource(source, basePath, globals)
			} else if source.Opts.LegendGraphic != nil && *source.Opts.LegendGraphic {
				s.LegendSources[k] = setting.LoadWMSLegendsSource(source, globals, fac)
			} else {
				s.Sources[k] = setting.LoadWMSMapSource(source, s, globals)
			}
		case *setting.TileSource:
			s.Sources[k] = setting.LoadTileSource(source, globals, s)
		case *setting.MapboxTileSource:
			s.Sources[k] = setting.LoadMapboxTileSource(source, globals, s, fac)
		case *setting.CesiumTileSource:
			s.Sources[k] = setting.LoadCesiumTileSource(source, globals, s, fac)
		case *setting.ArcGISSource:
			if source.Opts.Featureinfo != nil && *source.Opts.Featureinfo {
				s.InfoSources[k] = setting.LoadArcGISInfoSource(source, globals)
			} else {
				s.Sources[k] = setting.LoadArcGISSource(source, s, globals)
			}
		}
	}
}

func (s *Service) loadCaches(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	for k, c := range dataset.Caches {
		switch cache := c.(type) {
		case *setting.CacheSource:
			s.Caches[k] = setting.PreLoadCacheManager(cache, globals, s, fac)
		}
	}

	for k, c := range dataset.Caches {
		switch cache := c.(type) {
		case *setting.CacheSource:
			setting.LoadCacheManager(cache, globals, s, fac, s.Caches[k])
		}
	}
}

func (s *Service) loadService(dataset *setting.ProxyService, basePath string, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	switch srv := dataset.Service.(type) {
	case *setting.TMSService:
		s.Service = setting.LoadTMSService(srv, s)
	case *setting.WMSService:
		s.Service = setting.LoadWMSService(srv, s, globals, basePath)
	case *setting.MapboxService:
		s.Service = setting.LoadMapboxService(srv, globals, s, fac)
	case *setting.CesiumService:
		s.Service = setting.LoadCesiumService(srv, globals, s, fac)
	case *setting.WMTSService:
		if srv.Restful != nil && *srv.Restful {
			s.Service = setting.LoadWMTSRestfulService(srv, s)
		} else {
			s.Service = setting.LoadWMTSService(srv, s)
		}
	}
}

func (s *Service) Clean() {
}

func (s *Service) GetUUID() string {
	return s.UUID
}

func (s *Service) GetServiceType() ServiceType {
	return s.Type
}

func (s *Service) GetGrid(name string) geo.Grid {
	if g, ok := s.Grids[name]; ok {
		return g
	}
	return nil
}

func (s *Service) GetSource(name string) layer.Layer {
	if g, ok := s.Sources[name]; ok {
		return g
	}
	return nil
}

func (s *Service) GetCache(name string) cache.Manager {
	if l, ok := s.Caches[name]; ok {
		return l
	}
	return nil
}

func (s *Service) GetCacheSource(name string) layer.Layer {
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

		return cache.NewCacheSource(manager, cache_extent, manager.GetTileOptions(), nil, true, manager.GetQueryBuffer(), manager.GetReprojectSrcSrs(), manager.GetReprojectDstSrs())
	}
	return nil
}

func (s *Service) GetInfoSource(name string) layer.InfoLayer {
	if l, ok := s.InfoSources[name]; ok {
		return l
	}
	return nil
}

func (s *Service) GetLegendSource(name string) layer.LegendLayer {
	if l, ok := s.LegendSources[name]; ok {
		return l
	}
	return nil
}

func (s *Service) GetService() service.Service {
	return s.Service
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Service != nil {
		s.Service.ServeHTTP(w, r)
	}
}

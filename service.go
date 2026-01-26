package tileproxy

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/flywave/go-tileproxy/tile"
)

type ServiceType uint32

const (
	MapboxService ServiceType = 0
	WMSService    ServiceType = 1
	WMTSService   ServiceType = 2
	TileService   ServiceType = 3
	CesiumService ServiceType = 4
)

type Service struct {
	setting.ProxyInstance
	Id            string
	Type          ServiceType
	Service       service.Service
	Grids         map[string]geo.Grid
	Sources       map[string]layer.Layer
	InfoSources   map[string]layer.InfoLayer
	LegendSources map[string]layer.LegendLayer
	Caches        map[string]cache.Manager
	mu            sync.RWMutex
}

func NewService(dataset *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) *Service {
	ret := &Service{
		Id:            dataset.Id,
		Grids:         make(map[string]geo.Grid),
		Sources:       make(map[string]layer.Layer),
		Caches:        make(map[string]cache.Manager),
		InfoSources:   make(map[string]layer.InfoLayer),
		LegendSources: make(map[string]layer.LegendLayer),
	}
	ret.load(dataset, globals, fac)
	return ret
}

func (s *Service) load(dataset *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	s.loadGrids(dataset)
	s.loadSources(dataset, globals, fac)
	s.loadCaches(dataset, globals, fac)
	s.loadService(dataset, globals, fac)
}

func (s *Service) loadGrids(dataset *setting.ProxyService) {
	for k, g := range dataset.Grids {
		s.Grids[k] = setting.ConvertGridOpts(&g)
	}
}

func (s *Service) loadSources(dataset *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	for k, src := range dataset.Sources {
		switch source := src.(type) {
		case *setting.WMSSource:
			if source.Opts.FeatureInfo != nil && *source.Opts.FeatureInfo {
				s.InfoSources[k] = setting.LoadWMSInfoSource(source, globals)
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

func (s *Service) loadCaches(dataset *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
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

func (s *Service) loadService(dataset *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) {
	switch srv := dataset.Service.(type) {
	case *setting.TMSService:
		s.Service = setting.LoadTMSService(srv, s)
	case *setting.WMSService:
		s.Service = setting.LoadWMSService(srv, s, globals)
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

func (s *Service) GetId() string {
	return s.Id
}

func (s *Service) GetServiceType() ServiceType {
	return s.Type
}

func (s *Service) GetGrid(name string) geo.Grid {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if g, ok := s.Grids[name]; ok {
		return g
	}
	return nil
}

func (s *Service) GetSource(name string) layer.Layer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if g, ok := s.Sources[name]; ok {
		return g
	}
	return nil
}

func (s *Service) GetCache(name string) cache.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if l, ok := s.Caches[name]; ok {
		return l
	}
	return nil
}

func (s *Service) GetCacheSource(name string, opt tile.TileOptions) layer.Layer {
	s.mu.RLock()
	defer s.mu.RUnlock()
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
		if opt == nil {
			opt = manager.GetTileOptions()
		}
		return cache.NewCacheSource(manager, cache_extent, opt, nil, true, manager.GetQueryBuffer(), manager.GetReprojectSrcSrs(), manager.GetReprojectDstSrs())
	}
	return nil
}

func (s *Service) GetInfoSource(name string) layer.InfoLayer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if l, ok := s.InfoSources[name]; ok {
		return l
	}
	return nil
}

func (s *Service) GetLegendSource(name string) layer.LegendLayer {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

func (s *Service) Reload(newConfig *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) error {
	if newConfig == nil {
		return fmt.Errorf("new configuration is nil")
	}

	if s.Id != newConfig.Id {
		return fmt.Errorf("service ID mismatch: expected %s, got %s", s.Id, newConfig.Id)
	}

	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := s.stopService(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	s.load(newConfig, globals, fac)

	if err := s.startService(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func (s *Service) startService() error {
	if srv, ok := s.Service.(interface{ Start() error }); ok {
		return srv.Start()
	}
	return nil
}

func (s *Service) stopService() error {
	if srv, ok := s.Service.(interface{ Stop() error }); ok {
		return srv.Stop()
	}
	return nil
}

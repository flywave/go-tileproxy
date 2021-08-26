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

func NewDataset(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) *Dataset {
	ret := &Dataset{
		Identifier:    dataset.Identifier,
		Grids:         make(map[string]geo.Grid),
		Sources:       make(map[string]layer.Layer),
		Caches:        make(map[string]cache.Manager),
		InfoSources:   make(map[string]layer.InfoLayer),
		LegendSources: make(map[string]layer.LegendLayer),
	}
	ret.load(dataset, basePath, globals, preferred)
	return ret
}

func (s *Dataset) load(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) {
	s.loadGrids(dataset, basePath, globals, preferred)
	s.loadSources(dataset, basePath, globals, preferred)
	s.loadCaches(dataset, basePath, globals, preferred)
	s.loadService(dataset, basePath, globals, preferred)
}

func (s *Dataset) loadGrids(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) {
	for k, g := range dataset.Grids {
		s.Grids[k] = setting.ConvertGridOpts(&g)
	}
}

func (s *Dataset) loadSources(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) {
	for k, src := range dataset.Sources {
		switch source := src.(type) {
		case *setting.WMSSource:
			if source.Opts.FeatureInfo != nil && *source.Opts.FeatureInfo {
				s.InfoSources[k] = setting.LoadWMSInfoSource(source, basePath, preferred)
			} else if source.Opts.LegendGraphic != nil && *source.Opts.LegendGraphic {
				s.LegendSources[k] = setting.LoadWMSLegendsSource(source)
			} else {
				s.Sources[k] = setting.LoadWMSMapSource(source, s, preferred)
			}
		case *setting.TileSource:
			s.Sources[k] = setting.LoadTileSource(source, s)
		case *setting.MapboxTileSource:
			s.Sources[k] = setting.LoadMapboxTileSource(source, s)
		case *setting.LuokuangTileSource:
			s.Sources[k] = setting.LoadLuokuangTileSource(source, s)
		case *setting.ArcGISSource:
			if source.Opts.Featureinfo != nil && *source.Opts.Featureinfo {
				s.InfoSources[k] = setting.LoadArcGISInfoSource(source, preferred)
			} else {
				s.Sources[k] = setting.LoadArcGISSource(source, s, preferred)
			}
		}
	}
}

func (s *Dataset) loadCaches(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) {
	for k, c := range dataset.Caches {
		switch cache := c.(type) {
		case *setting.Caches:
			s.Caches[k] = setting.LoadCacheManager(cache, globals, s)
		}
	}
}

func (s *Dataset) loadService(dataset *setting.ProxyDataset, basePath string, globals *setting.GlobalsSetting, preferred geo.PreferredSrcSRS) {
	switch srv := dataset.Service.(type) {
	case *setting.TMSService:
		s.Service = setting.LoadTMSService(srv, s)
	case *setting.WMSService:
		s.Service = setting.LoadWMSService(srv, s, basePath, preferred)
	case *setting.MapboxService:
		s.Service = setting.LoadMapboxService(srv, s)
	case *setting.WMTSService:
		if srv.KVP != nil && *srv.KVP {
			s.Service = setting.LoadWMTSService(srv, s)
		} else if srv.Restful != nil && *srv.Restful {
			s.Service = setting.LoadWMTSRestfulService(srv, s)
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

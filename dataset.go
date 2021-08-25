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

type Dataset struct {
	setting.ProxyInstance
	Identifier   string
	Type         ServiceType
	Service      service.Service
	LegendLayers map[string]layer.LegendLayer
	InfoLayers   map[string]layer.InfoLayer
	MapLayers    map[string]layer.Layer
	Grids        map[string]geo.Grid
	Coverages    map[string]geo.Coverage
	Sources      map[string]sources.Source
}

func (s *Dataset) GetIdentifier() string {
	return s.Identifier
}

func (s *Dataset) GetServiceType() ServiceType {
	return s.Type
}

func (s *Dataset) GetLegendLayer(name string) layer.LegendLayer {
	if l, ok := s.LegendLayers[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetInfoLayer(name string) layer.InfoLayer {
	if l, ok := s.InfoLayers[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetMapLayer(name string) layer.Layer {
	if l, ok := s.MapLayers[name]; ok {
		return l
	}
	return nil
}

func (s *Dataset) GetGrid(name string) geo.Grid {
	if g, ok := s.Grids[name]; ok {
		return g
	}
	return nil
}

func (s *Dataset) GetCoverage(name string) geo.Coverage {
	if g, ok := s.Coverages[name]; ok {
		return g
	}
	return nil
}

func (s *Dataset) GetSource(name string) sources.Source {
	if g, ok := s.Sources[name]; ok {
		return g
	}
	return nil
}

func (s *Dataset) GetService() service.Service {
	return s.Service
}

func (s *Dataset) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

package service

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	_ "github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSService struct {
}

func (s *WMTSService) GetCapabilities() {

}

func (s *WMTSService) GetTile() {

}

func (s *WMTSService) GetFeatureInfo() {

}

func (s *WMTSService) authorizeTileLayer(tile_layer TileLayer, tile_request request.TileRequest) {

}

func (s *WMTSService) authorizedTileLayers() {

}

func (s *WMSService) checkRequest() {

}

type WMTSRestService struct {
}

type WMTSTileLayer struct{}

const (
	METERS_PER_DEEGREE = 111319.4907932736
)

func meter_per_unit(srs geo.Proj) float64 {
	if srs.IsLatLong() {
		return METERS_PER_DEEGREE
	}
	return 1
}

type TileMatrixSet struct {
}

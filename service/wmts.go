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

type WMTSTileLayer struct {
}

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
	grid     *geo.TileGrid
	name     string
	srs_name string
}

func NewTileMatrixSet(grid *geo.TileGrid) *TileMatrixSet {
	return &TileMatrixSet{grid: grid, name: grid.Name, srs_name: grid.Srs.SrsCode}
}

func (s *TileMatrixSet) GetTileMatrices() map[string]interface{} {
	ret := make(map[string]interface{})
	for level, res := range s.grid.Resolutions {
		x, y, z := s.grid.OriginTile(level, geo.ORIGIN_UL)
		bbox := s.grid.TileBBox([3]int{x, y, z}, false)
		topleft := []float64{bbox.Min[0], bbox.Max[1]}
		if s.grid.Srs.IsAxisOrderNE() {
			topleft = []float64{bbox.Max[1], bbox.Min[0]}
		}
		grid_size := s.grid.GridSizes[level]
		scale_denom := res / (0.28 / 1000) * meter_per_unit(s.grid.Srs)
		ret["identifier"] = level
		ret["topleft"] = topleft
		ret["grid_size"] = grid_size
		ret["scale_denom"] = scale_denom
		ret["tile_size"] = s.grid.TileSize
	}
	return ret
}

package service

import (
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	_ "github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSService struct {
	BaseService
	Metadata   map[string]string
	MaxTileAge time.Duration
	Layers     map[string]WMTSTileLayer
	MatrixSets map[string]TileMatrixSet
}

func (s *WMSService) getMatrixSets(layers []WMTSTileLayer) (map[string]WMTSTileLayer, map[string]TileMatrixSet) {
	/**sets := make(map[string]TileMatrixSet)
	        layers_grids = odict()
	        for _,layer := range layers {
	            grid = layer.grid
	            if !grid.supports_access_with_origin("nw") {
	                continue
				}
	            if grid.name not in sets {
	                    sets[grid.name] = TileMatrixSet(grid)
				}
	            layers_grids.setdefault(layer.name, odict())[grid.name] = layer
			}
	        wmts_layers = odict()
	        for layer_name, layers in layers_grids.items() {
	            wmts_layers[layer_name] = WMTSTileLayer(layers)
			}
	        return wmts_layers, sets.values()**/
	return nil, nil
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

type WMTSCapabilities struct {
}

type WMTSRestfulCapabilities struct {
	WMTSCapabilities
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

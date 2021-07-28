package service

import (
	"strconv"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
	_ "github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSService struct {
	BaseService
	Metadata    map[string]string
	MaxTileAge  time.Duration
	Layers      map[string]*WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats []string
}

func (s *WMTSService) getMatrixSets(layers []RenderLayer) (map[string]*WMTSTileLayer, map[string]*TileMatrixSet) {
	sets := make(map[string]*TileMatrixSet)
	layers_grids := make(map[string][]RenderLayer)
	for _, layer := range layers {
		grid := layer.GetGrid()
		if !grid.SupportsAccessWithOrigin(geo.ORIGIN_NW) {
			continue
		}
		if _, ok := sets[grid.Name]; !ok {
			sets[grid.Name] = NewTileMatrixSet(grid)
		}
		layers_grids[grid.Name] = append(layers_grids[grid.Name], layer)
	}
	wmts_layers := make(map[string]*WMTSTileLayer)
	for layer_name, layers := range layers_grids {
		wmts_layers[layer_name] = NewWMTSTileLayer(layers)
	}
	return wmts_layers, sets
}

func (s *WMTSService) serviceMetadata(tms_request request.Request) map[string]string {
	md := s.Metadata
	md["url"] = tms_request.Http.URL.Host
	return md
}

func (s *WMTSService) GetCapabilities(request request.Request) *Response {
	service := s.serviceMetadata(&request.WMTSRequest)
	layers := s.authorizedTileLayers()

	cap := newWMTSCapabilities(service, layers, s.MatrixSets, s.InfoFormats)

	result := cap.render(request)

	return NewResponse(result, 200, "", "application/xml")
}

func (s *WMTSService) GetTile(request request.Request) *Response {
	s.checkRequest(&request.WMTSRequest)

	tile_layer := s.Layers[request.Layer].Get(request.tilematrixset)
	if request.Format == nil {
		tf := tile.TileFormat(tile_layer.GetFormat())
		request.Format = &tf
	}

	s.checkRequestDimensions(tile_layer, &request.WMTSRequest)

	limited_to := s.authorizeTileLayer(tile_layer, request)

	decorate_tile := func(image tile.Source) tile.Source {
		query_extent := &geo.MapExtent{Srs: tile_layer.GetGrid().Srs, BBox: tile_layer.GetTileBBox(request, request.UseProfiles, false)}
		return s.DecorateImg(image, "wmts", []string{tile_layer.GetName()}, query_extent)
	}

	tile := tile_layer.Render(request, false, limited_to, decorate_tile)

	resp := NewResponse(tile.getBuffer(), -1, "", "image/"+tile.getFormat())
	resp.cacheHeaders(tile.getTimestamp(), []string{tile.getTimestamp().String(), strconv.Itoa(tile.getSize())},
		int(s.MaxTileAge.Seconds()))
	resp.makeConditional(request.Http)
	return resp
}

func (s *WMTSService) GetFeatureInfo(info_request request.Request) {

}

func (s *WMTSService) authorizeTileLayer(tile_layer RenderLayer, tile_request request.Request) geo.Coverage {
	return nil
}

func (s *WMTSService) authorizedTileLayers() []*WMTSTileLayer {
	return nil
}

func (s *WMTSService) checkRequestDimensions(tile_layer RenderLayer, request request.Request) {

}

func (s *WMTSService) checkRequest(request request.Request) {

}

const (
	DEFAULT_WMTS_TEMPLATE      = "/{{ .Layer }}/{{ .TileMatrixSet }}/{{ .TileMatrix }}/{{ .TileCol }}/{{ .TileRow }}.{{ .Format }}"
	DEFAULT_WMTS_INFO_TEMPLATE = "/{{ .Layer }}/{{ .TileMatrixSet }}/{{ .TileMatrix }}/{{ .TileCol }}/{{ .TileRow }}/{{ .I }}/{{ .J }}.{{ .InfoFormat }}"
)

type WMTSRestService struct {
	WMTSService
	names          []string
	requestMethods []string
	template       string
	infoTemplate   string
}

type WMTSTileLayer struct {
	layers    map[string]RenderLayer
	baseLayer RenderLayer
}

func NewWMTSTileLayer(layer []RenderLayer) *WMTSTileLayer {
	return nil
}

func (t *WMTSTileLayer) Get(p string) RenderLayer {
	return t.layers[p]
}

type WMTSCapabilities struct {
}

func (c *WMTSCapabilities) render(request request.WMTS100CapabilitiesRequest) []byte {
	return nil
}

func newWMTSCapabilities(md map[string]string, layers []*WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats []string) *WMTSCapabilities {
	return nil
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

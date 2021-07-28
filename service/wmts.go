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
	req := tms_request.(*request.BaseRequest)
	md := s.Metadata
	md["url"] = req.Http.URL.Host
	return md
}

func (s *WMTSService) GetCapabilities(req request.Request) *Response {
	tile_request := req.(*request.WMTS100CapabilitiesRequest)

	service := s.serviceMetadata(&tile_request.WMTSRequest)
	layers := s.authorizedTileLayers()

	cap := newWMTSCapabilities(service, layers, s.MatrixSets, s.InfoFormats)

	result := cap.render(tile_request)

	return NewResponse(result, 200, "", "application/xml")
}

func (s *WMTSService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.WMTS100TileRequest)
	s.checkRequest(&tile_request.WMTSRequest)

	tile_layer := s.Layers[tile_request.Layer].Get(tile_request.TileMatrixSet)
	if tile_request.Format == "" {
		tf := tile.TileFormat(tile_layer.GetFormat())
		tile_request.Format = tf
	}

	s.checkRequestDimensions(tile_layer, &tile_request.WMTSRequest)

	limited_to := s.authorizeTileLayer(tile_layer, tile_request)

	decorate_tile := func(image tile.Source) tile.Source {
		query_extent := &geo.MapExtent{Srs: tile_layer.GetGrid().Srs, BBox: tile_layer.GetTileBBox(tile_request, tile_request.UseProfiles, false)}
		return s.DecorateImg(image, "wmts", []string{tile_layer.GetName()}, query_extent)
	}

	tile := tile_layer.Render(tile_request, false, limited_to, decorate_tile)

	resp := NewResponse(tile.getBuffer(), -1, "", "image/"+tile.getFormat())
	resp.cacheHeaders(tile.getTimestamp(), []string{tile.getTimestamp().String(), strconv.Itoa(tile.getSize())},
		int(s.MaxTileAge.Seconds()))
	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *WMTSService) GetFeatureInfo(req request.Request) *Response  {
	infos := []*resource.FeatureInfo{}
	info_request := req.(*request.WMTS100FeatureInfoRequest)
	s.checkRequest(&info_request.WMTSRequest, s.InfoFormats)

	tile_layer := s.Layers[info_request.Layer].Get(info_request.TileMatrixSet)
	if info_request.Format == "" {
		tf := tile.TileFormat(tile_layer.GetFormat())
		info_request.Format = tf
	}

	feature_count = None

	if hasattr(request, 'params') {
		feature_count = request.params.get('feature_count', None)
	}

	bbox = tile_layer.grid.tile_bbox(request.tile)
	query = InfoQuery(bbox, tile_layer.grid.tile_size, tile_layer.grid.srs, request.pos,
					  request.infoformat, feature_count=feature_count)
	self.check_request_dimensions(tile_layer, request)
	coverage = self.authorize_tile_layer(tile_layer, request, featureinfo=True)

	if not tile_layer.info_sources {
		raise RequestError('layer %s not queryable' % str(request.layer),
			code='OperationNotSupported', request=request)
	}

	if coverage and not coverage.contains(query.coord, query.srs) {
		infos = []
	} else {
		for source in tile_layer.info_sources {
			info = source.get_info(query)
			if info is None {
				continue
			}
			infos.append(info)
		}
	}

	mimetype = request.infoformat

	if not infos{
		return NewResponse('', 200, "", mimetype)
	}

	resp, _ = resource.CombineDocs(infos)

	return NewResponse(resp, 200, "", mimetype)
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

func (c *WMTSCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
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

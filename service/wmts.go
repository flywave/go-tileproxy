package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"

	_ "github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSService struct {
	BaseService
	Metadata    map[string]string
	MaxTileAge  time.Duration
	Layers      map[string]WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats map[string]string
}

func NewWMTSService(layers []*TileLayer, md map[string]string, MaxTileAge *time.Duration, info_formats map[string]string) *WMTSRestService {
	return nil
}

func (s *WMTSService) getMatrixSets(tlayers []*TileLayer) (map[string]*WMTSTileLayer, map[string]*TileMatrixSet) {
	sets := make(map[string]*TileMatrixSet)
	layers_grids := make(map[string][]*TileLayer)
	for _, layer := range tlayers {
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

	service := s.serviceMetadata(tile_request)
	layers := s.authorizedTileLayers()

	cap := newWMTSCapabilities(service, layers, s.MatrixSets, s.InfoFormats)

	result := cap.render(tile_request)

	return NewResponse(result, 200, "application/xml")
}

func (s *WMTSService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.WMTS100TileRequest)
	s.checkRequest(tile_request, nil)

	params := request.NewWMTSTileRequestParams(tile_request.Params)

	tile_layer := s.Layers[params.GetLayer()][params.GetTileMatrixSet()]
	if params.GetFormat() == "" {
		tf := tile.TileFormat(tile_layer.GetFormat())
		params.SetFormat(tf)
	}

	s.checkRequestDimensions(tile_layer, tile_request)

	limited_to := s.authorizeTileLayer(tile_layer, tile_request, false)

	decorate_tile := func(image tile.Source) tile.Source {
		query_extent := &geo.MapExtent{Srs: tile_layer.GetGrid().Srs, BBox: tile_layer.GetTileBBox(tile_request, tile_request.UseProfiles, false)}
		return s.DecorateImg(image, "wmts", []string{tile_layer.GetName()}, query_extent)
	}

	tile := tile_layer.Render(tile_request, false, limited_to, decorate_tile)

	resp := NewResponse(tile.getBuffer(), -1, "image/"+tile.getFormat())
	resp.cacheHeaders(tile.getTimestamp(), []string{tile.getTimestamp().String(), strconv.Itoa(tile.getSize())},
		int(s.MaxTileAge.Seconds()))
	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *WMTSService) GetFeatureInfo(req request.Request) *Response {
	infos := []resource.FeatureInfoDoc{}
	info_request := req.(*request.WMTS100FeatureInfoRequest)

	params := request.NewWMTSFeatureInfoRequestParams(info_request.Params)

	infoformat := params.GetInfoformat()

	s.checkRequest(&info_request.WMTSRequest, &infoformat)

	tile_layer := s.Layers[params.GetLayer()][params.GetTileMatrixSet()]
	if params.GetFormat() == "" {
		tf := tile.TileFormat(tile_layer.GetFormat())
		params.SetFormat(tf)
	}

	var feature_count *int

	if req.GetParams() != nil {
		if v, ok := req.GetParams()["feature_count"]; ok {
			fc, _ := strconv.Atoi(v[0])
			feature_count = &fc
		}
	}

	bbox := tile_layer.GetGrid().TileBBox(params.GetCoord(), false)
	query := &layer.InfoQuery{BBox: bbox, Size: [2]uint32{tile_layer.GetGrid().TileSize[0], tile_layer.GetGrid().TileSize[1]}, Srs: tile_layer.GetGrid().Srs, Pos: params.GetPos(),
		InfoFormat: infoformat, FeatureCount: feature_count}
	s.checkRequestDimensions(tile_layer, req)

	coverage := s.authorizeTileLayer(tile_layer, req, true)

	if coverage != nil && !coverage.ContainsPoint(query.GetCoord(), query.Srs) {
		infos = nil
	} else {
		for _, source := range tile_layer.infoSources {
			info := source.GetInfo(query)
			if info == nil {
				continue
			}
			infos = append(infos, info)
		}
	}

	mimetype := infoformat

	if infos == nil || len(infos) == 0 {
		return NewResponse([]byte{}, 200, mimetype)
	}

	resp, _ := resource.CombineDocs(infos, nil)

	return NewResponse(resp, 200, mimetype)
}

func (s *WMTSService) authorizeTileLayer(tile_layer *TileLayer, tile_request request.Request, featureinfo bool) geo.Coverage {
	return nil
}

func (s *WMTSService) authorizedTileLayers() []WMTSTileLayer {
	ret := []WMTSTileLayer{}
	for _, v := range s.Layers {
		ret = append(ret, v)
	}
	return ret
}

func (s *WMTSService) checkRequestDimensions(tile_layer *TileLayer, request request.Request) {
	//
}

func (s *WMTSService) checkRequest(req request.Request, infoformat *string) error {
	switch wreq := req.(type) {
	case *request.WMTS100TileRequest:
		{
			params := request.NewWMTSTileRequestParams(wreq.Params)

			if _, ok := s.Layers[params.GetLayer()]; !ok {
				return errors.New("unknown layer: " + params.GetLayer())
			}
			if _, ok := s.Layers[params.GetLayer()][params.GetTileMatrixSet()]; !ok {
				return errors.New("unknown tilematrixset: " + params.GetTileMatrixSet())
			}
		}
	case *request.WMTS100CapabilitiesRequest:
		{
			//
		}
	case *request.WMTS100FeatureInfoRequest:
		{
			params := request.NewWMTSFeatureInfoRequestParams(wreq.Params)
			infoformat := params.GetInfoformat()
			if infoformat != "" {
				if strings.Contains(infoformat, "/") {
					if _, ok := s.InfoFormats[infoformat]; !ok {
						return errors.New("unknown infoformat: " + infoformat)
					}
				} else {
					if _, ok := s.InfoFormats[infoformat]; !ok {
						return errors.New("unknown infoformat: " + infoformat)
					}
					params.SetInfoformat(s.InfoFormats[infoformat])
				}
			}
		}
	}
	return nil
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

func NewWMTSRestService(layers []*TileLayer, md map[string]string, MaxTileAge *time.Duration, template string, fi_template string, info_formats map[string]string) *WMTSRestService {
	return nil
}

func (s *WMTSRestService) checkRequestDimensions(tile_layer *TileLayer, request request.Request) {
	//
}

type WMTSTileLayer map[string]*TileLayer

func NewWMTSTileLayer(layer []*TileLayer) *WMTSTileLayer {
	return nil
}

type WMTSCapabilities struct {
	Service     map[string]string
	Layers      []WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats map[string]string
}

func (c *WMTSCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	return nil
}

func newWMTSCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return nil
}

type WMTSRestfulCapabilities struct {
	WMTSCapabilities
}

func newWMTSRestfulCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return nil
}

func (c *WMTSRestfulCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	return nil
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

package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/ogc-osgeo/pkg/wmts100"
	"github.com/flywave/ogc-osgeo/pkg/wsc110"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type WMTSMetadata struct {
	Title             string
	Abstract          string
	KeywordList       []string
	URL               string
	Fees              *string
	AccessConstraints *string
	Provider          *wsc110.ServiceProvider
}

type WMTSService struct {
	BaseService
	Metadata    *WMTSMetadata
	MaxTileAge  *time.Duration
	Layers      map[string]WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats map[string]string
}

type WMTSServiceOptions struct {
	Layers      map[string]Provider
	Metadata    *WMTSMetadata
	MaxTileAge  *time.Duration
	InfoFormats map[string]string
}

func NewWMTSService(opts *WMTSServiceOptions) *WMTSService {
	ret := &WMTSService{InfoFormats: opts.InfoFormats, MaxTileAge: opts.MaxTileAge, Metadata: opts.Metadata}
	layer, ms := ret.getMatrixSets(opts.Layers)
	ret.Layers = layer
	ret.MatrixSets = ms
	ret.router = map[string]func(r request.Request) *Response{
		"tile": func(r request.Request) *Response {
			return ret.GetTile(r)
		},
		"featureinfo": func(r request.Request) *Response {
			return ret.GetFeatureInfo(r)
		},
		"capabilities": func(r request.Request) *Response {
			return ret.GetCapabilities(r)
		},
	}
	ret.requestParser = func(r *http.Request) request.Request {
		return request.MakeWMTSRequest(r, false)
	}
	return ret
}

func (s *WMTSService) getMatrixSets(tlayers map[string]Provider) (map[string]WMTSTileLayer, map[string]*TileMatrixSet) {
	sets := make(map[string]*TileMatrixSet)
	layers_grids := make(map[string]map[string]Provider)
	for _, layer := range tlayers {
		grid := layer.GetGrid()
		if !grid.SupportsAccessWithOrigin(geo.ORIGIN_NW) {
			continue
		}
		if _, ok := sets[grid.Name]; !ok {
			sets[grid.Name] = NewTileMatrixSet(grid)
		}
		if _, ok := layers_grids[layer.GetName()]; !ok {
			layers_grids[layer.GetName()] = make(map[string]Provider)
		}
		layers_grids[layer.GetName()][grid.Name] = layer
	}
	wmts_layers := make(map[string]WMTSTileLayer)
	for layer_name, layers := range layers_grids {
		wmts_layers[layer_name] = WMTSTileLayer(layers)
	}
	return wmts_layers, sets
}

func (s *WMTSService) serviceMetadata(tms_request request.Request) WMTSMetadata {
	req := tms_request.(*request.BaseRequest)
	md := *s.Metadata
	md.URL = req.Http.URL.Host
	return md
}

func (s *WMTSService) GetCapabilities(req request.Request) *Response {
	tile_request := req.(*request.WMTS100CapabilitiesRequest)

	service := s.serviceMetadata(tile_request)
	layers := s.authorizedTileLayers()

	cap := newWMTSCapabilities(&service, layers, s.MatrixSets, s.InfoFormats)
	result := cap.render(tile_request)

	return NewResponse(result, 200, "application/xml")
}

func (s *WMTSService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.WMTS100TileRequest)
	s.checkRequest(tile_request, nil)

	params := request.NewWMTSTileRequestParams(tile_request.Params)

	tileMatrixSet := params.GetTileMatrixSet()
	layer := params.GetLayer()

	var tile_layer Provider
	if layerSet, ok := s.Layers[layer]; ok {
		if p, ok := layerSet[tileMatrixSet]; ok {
			tile_layer = p
		} else {
			resp := NewRequestError(fmt.Sprintf("layer %s's tileMatrixSet %s not queryable", layer, tileMatrixSet), "OperationNotSupported", &WMTS100ExceptionHandler{}, req, false, nil)
			return resp.Render()
		}
	} else {
		resp := NewRequestError(fmt.Sprintf("layer %s's not queryable", layer), "OperationNotSupported", &WMTS100ExceptionHandler{}, req, false, nil)
		return resp.Render()
	}

	if params.GetFormat() == "" {
		tf := tile.TileFormat(tile_layer.GetFormat())
		params.SetFormat(tf)
	}

	tp := tile_layer.(*TileProvider)

	err := s.checkRequestDimensions(tp, tile_request)

	if err != nil {
		return err.Render()
	}

	limited_to := s.authorizeTileLayer(tp, tile_request, false)

	decorateTile := func(image tile.Source) tile.Source {
		cerr, bbox := tile_layer.GetTileBBox(tile_request, tile_request.UseProfiles, false)
		if cerr != nil {
			return nil
		}
		query_extent := &geo.MapExtent{Srs: tile_layer.GetGrid().Srs, BBox: bbox}
		return s.DecorateTile(image, "wmts", []string{tile_layer.GetName()}, query_extent)
	}

	err, tile := tile_layer.Render(tile_request, false, limited_to, decorateTile)
	if err != nil {
		return err.Render()
	}

	resp := NewResponse(tile.getBuffer(), -1, tile.GetFormatMime())

	if s.MaxTileAge != nil {
		timestrs := []string{}

		if tile.getTimestamp() != nil {
			timestrs = append(timestrs, tile.getTimestamp().String())
		}

		timestrs = append(timestrs, strconv.Itoa(tile.getSize()))

		resp.cacheHeaders(tile.getTimestamp(), timestrs, int(s.MaxTileAge.Seconds()))
	}

	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *WMTSService) GetFeatureInfo(req request.Request) *Response {
	infos := []resource.FeatureInfoDoc{}
	info_request := req.(*request.WMTS100FeatureInfoRequest)

	params := request.NewWMTSFeatureInfoRequestParams(info_request.Params)

	infoformat := params.GetInfoformat()

	err := s.checkRequest(&info_request.WMTSRequest, &infoformat)

	if err != nil {
		return err.Render()
	}

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

	tp := tile_layer.(*TileProvider)

	err = s.checkRequestDimensions(tp, req)

	if err != nil {
		return err.Render()
	}

	coverage := s.authorizeTileLayer(tp, req, true)

	if tp.infoSources == nil {
		resp := NewRequestError(fmt.Sprintf("layer %s not queryable", params.GetLayer()), "OperationNotSupported", &WMTS100ExceptionHandler{}, req, false, nil)
		return resp.Render()
	}

	if coverage != nil && !coverage.ContainsPoint(query.GetCoord(), query.Srs) {
		infos = nil
	} else {
		for _, source := range tp.infoSources {
			info := source.GetInfo(query)
			if info == nil {
				continue
			}
			infos = append(infos, info)
		}
	}

	mimetype := infoformat

	if len(infos) == 0 {
		return NewResponse([]byte{}, 200, mimetype)
	}

	resp, _ := resource.CombineDocs(infos, nil)

	return NewResponse(resp, 200, mimetype)
}

func (s *WMTSService) authorizeTileLayer(_ *TileProvider, _ request.Request, _ bool) geo.Coverage {
	return nil
}

func (s *WMTSService) authorizedTileLayers() []WMTSTileLayer {
	ret := []WMTSTileLayer{}
	for _, v := range s.Layers {
		ret = append(ret, v)
	}
	return ret
}

func (s *WMTSService) checkRequestDimensions(_ *TileProvider, _ request.Request) *RequestError {
	return nil
}

func (s *WMTSService) checkRequest(req request.Request, _ *string) *RequestError {
	switch wreq := req.(type) {
	case *request.WMTS100TileRequest:
		{
			params := request.NewWMTSTileRequestParams(wreq.Params)

			if _, ok := s.Layers[params.GetLayer()]; !ok {
				return NewRequestError("unknown layer: "+params.GetLayer(), "InvalidParameterValue", &WMTS100ExceptionHandler{}, req, false, nil)
			}
			if _, ok := s.Layers[params.GetLayer()][params.GetTileMatrixSet()]; !ok {
				return NewRequestError("unknown tilematrixset: "+params.GetTileMatrixSet(), "InvalidParameterValue", &WMTS100ExceptionHandler{}, req, false, nil)
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
						return NewRequestError("unknown infoformat: "+infoformat, "InvalidParameterValue", &WMTS100ExceptionHandler{}, req, false, nil)
					}
				} else {
					if _, ok := s.InfoFormats[infoformat]; !ok {
						return NewRequestError("unknown infoformat: "+infoformat, "InvalidParameterValue", &WMTS100ExceptionHandler{}, req, false, nil)
					}
					params.SetInfoformat(s.InfoFormats[infoformat])
				}
			}
		}
	}
	return nil
}

const (
	DEFAULT_WMTS_TEMPLATE      = "/{Layer}/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}.{Format}"
	DEFAULT_WMTS_INFO_TEMPLATE = "/{Layer}/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}/{I}/{J}.{InfoFormat}"
)

type WMTSRestService struct {
	WMTSService
	names            []string
	requestMethods   []string
	template         string
	infoTemplate     string
	urlConverter     *request.URLTemplateConverter
	infoUrlConverter *request.URLTemplateConverter
}

type WMTSRestServiceOptions struct {
	Layers                     map[string]Provider
	Metadata                   *WMTSMetadata
	MaxTileAge                 *time.Duration
	InfoFormats                map[string]string
	RestfulTemplate            string
	RestfulFeatureinfoTemplate string
}

func NewWMTSRestService(opts *WMTSRestServiceOptions) *WMTSRestService {
	ret := &WMTSRestService{
		names:          []string{"wmts"},
		requestMethods: []string{"tile", "capabilities"},
		WMTSService: WMTSService{
			InfoFormats: opts.InfoFormats,
			MaxTileAge:  opts.MaxTileAge,
			Metadata:    opts.Metadata,
		},
	}
	lay, ms := ret.getMatrixSets(opts.Layers)
	ret.Layers = lay
	ret.MatrixSets = ms

	if opts.RestfulTemplate == "" {
		ret.template = DEFAULT_WMTS_TEMPLATE
	} else {
		ret.template = opts.RestfulTemplate
	}

	if opts.RestfulFeatureinfoTemplate == "" {
		ret.infoTemplate = DEFAULT_WMTS_INFO_TEMPLATE
	} else {
		ret.infoTemplate = opts.RestfulFeatureinfoTemplate
	}

	ret.urlConverter = request.NewURLTemplateConverter(ret.template)
	ret.infoUrlConverter = request.NewFeatureInfoURLTemplateConverter(ret.infoTemplate)
	ret.router = map[string]func(r request.Request) *Response{
		"tile": func(r request.Request) *Response {
			return ret.GetTile(r)
		},
		"featureinfo": func(r request.Request) *Response {
			return ret.GetFeatureInfo(r)
		},
		"capabilities": func(r request.Request) *Response {
			return ret.GetCapabilities(r)
		},
	}
	ret.requestParser = func(r *http.Request) request.Request {
		return request.MakeWMTSRequest(r, false)
	}
	return ret
}

type WMTSTileLayer map[string]Provider

func (l WMTSTileLayer) frist() Provider {
	for k := range l {
		return l[k]
	}
	return nil
}

func (l WMTSTileLayer) GetTitle() string {
	return ""
}

func (l WMTSTileLayer) GetName() string {
	p := l.frist()
	if p != nil {
		return p.GetName()
	}
	return ""
}

func (l WMTSTileLayer) GetGrids() []*geo.TileGrid {
	ret := []*geo.TileGrid{}
	for _, v := range l {
		ret = append(ret, v.GetGrid())
	}
	return ret
}

func (l WMTSTileLayer) GetGrid() *geo.TileGrid {
	p := l.frist()
	if p != nil {
		return p.GetGrid()
	}
	return nil
}

func (l WMTSTileLayer) LLBBox() vec2d.Rect {
	p := l.frist()
	if p != nil {
		return limitLLBBox(p.GetExtent().GetLLBBox())
	}
	return vec2d.Rect{}
}

func (l WMTSTileLayer) GetBBox() vec2d.Rect {
	p := l.frist()
	if p != nil {
		return p.GetBBox()
	}
	return vec2d.Rect{}
}

func (l WMTSTileLayer) GetSrs() geo.Proj {
	p := l.frist()
	if p != nil {
		return p.GetSrs()
	}
	return nil
}

func (l WMTSTileLayer) GetFormatMimeType() string {
	p := l.frist()
	if p != nil {
		return p.GetFormatMimeType()
	}
	return ""
}

func (l WMTSTileLayer) GetFormat() string {
	p := l.frist()
	if p != nil {
		return p.GetFormat()
	}
	return ""
}

func (l WMTSTileLayer) GetTileBBox(request request.TiledRequest, use_profiles bool, limit bool) (*RequestError, vec2d.Rect) {
	p := l.frist()
	if p != nil {
		return p.GetTileBBox(request, use_profiles, limit)
	}
	return nil, vec2d.Rect{}
}

const (
	METERS_PER_DEEGREE = 111319.4907932736
)

func meterPerUnit(srs geo.Proj) float64 {
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
	return &TileMatrixSet{grid: grid, name: grid.Name, srs_name: grid.Srs.GetSrsCode()}
}

func (s *TileMatrixSet) GetTileMatrices() []map[string]string {
	ret := []map[string]string{}
	for level, res := range s.grid.Resolutions {
		m := make(map[string]string)
		tile_coord := s.grid.OriginTile(level, geo.ORIGIN_UL)
		bbox := s.grid.TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, false)
		topleft := []float64{bbox.Min[0], bbox.Max[1]}
		if s.grid.Srs.IsAxisOrderNE() {
			topleft = []float64{bbox.Max[1], bbox.Min[0]}
		}
		grid_size := s.grid.GridSizes[level]
		scale_denom := res / (0.28 / 1000) * meterPerUnit(s.grid.Srs)

		m["identifier"] = strconv.Itoa(level)
		m["topleft"] = fmt.Sprintf("%f %f", topleft[0], topleft[1])
		m["scale_denom"] = strconv.FormatFloat(scale_denom, 'f', -1, 64)
		m["tile_width"] = strconv.Itoa(int(s.grid.TileSize[0]))
		m["tile_height"] = strconv.Itoa(int(s.grid.TileSize[1]))
		m["matrix_width"] = strconv.Itoa(int(grid_size[0]))
		m["matrix_height"] = strconv.Itoa(int(grid_size[1]))

		ret = append(ret, m)
	}
	return ret
}

var (
	WMTSStatusCodes = map[string]int{
		"TileOutOfRange":        400,
		"MissingParameterValue": 400,
		"InvalidParameterValue": 400,
		"OperationNotSupported": 501,
	}
)

type WMTS100ExceptionHandler struct {
	ExceptionHandler
}

func (h *WMTS100ExceptionHandler) Render(err *RequestError) *Response {
	status_code := 500
	if sc, ok := WMTSStatusCodes[err.Code]; ok {
		status_code = sc
	}

	exp := wsc110.Exceptions(wmts100.NewExceptions(err.Message, err.Code))
	report := exp.ToReport("1.0.0")

	return NewResponse(report.ToBytes(), status_code, "text/xml")
}

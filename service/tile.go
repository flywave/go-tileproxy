package service

import (
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"
	tms100 "github.com/flywave/ogc-osgeo/pkg/tms100"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

var (
	TILE_SERVICE_NAMES           = []string{"tiles", "tms"}
	TILE_SERVICE_REQUEST_METHODS = []string{"map", "tms_capabilities", "tms_root_resource"}
)

type TileMetadata struct {
	Title    string
	Abstract string
	URL      string
}

type TileService struct {
	BaseService
	Layers             map[string]Provider
	Metadata           *TileMetadata
	MaxTileAge         *time.Duration
	UseDimensionLayers bool
	Origin             string
}

type TileServiceOptions struct {
	Layers             map[string]Provider
	Metadata           *TileMetadata
	MaxTileAge         *time.Duration
	UseDimensionLayers bool
	Origin             string
}

func NewTileService(opts *TileServiceOptions) *TileService {
	s := &TileService{Layers: opts.Layers, Metadata: opts.Metadata, MaxTileAge: opts.MaxTileAge, UseDimensionLayers: opts.UseDimensionLayers, Origin: opts.Origin}
	s.router = map[string]func(r request.Request) *Response{
		"map": func(r request.Request) *Response {
			return s.GetMap(r)
		},
		"tms_capabilities": func(r request.Request) *Response {
			return s.GetCapabilities(r)
		},
		"tms_root_resource": func(r request.Request) *Response {
			return s.RootResource(r)
		},
	}
	s.requestParser = func(r *http.Request) request.Request {
		return request.MakeTileRequest(r, true)
	}
	return s
}

func (s *TileService) GetMap(req request.Request) *Response {
	var tile_request *request.TileRequest
	switch r := req.(type) {
	case *request.TileRequest:
		tile_request = r
	case *request.TMSRequest:
		tile_request = &r.TileRequest
	}
	if s.Origin != "" && tile_request.Origin == "" {
		tile_request.Origin = s.Origin
	}
	layer, limit_to, err := s.getLayer(tile_request)

	if err != nil {
		return err.Render()
	}

	decorateTile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*TileProvider)
		var bbox vec2d.Rect
		err, bbox = layer.GetTileBBox(tile_request, tile_request.UseProfiles, false)
		if err != nil {
			return nil
		}
		query_extent := &geo.MapExtent{Srs: tilelayer.grid.srs, BBox: bbox}
		return s.DecorateTile(image, "tms", []string{tilelayer.GetName()}, query_extent)
	}

	err, t := layer.Render(tile_request, tile_request.UseProfiles, limit_to, decorateTile)
	if err != nil {
		return err.Render()
	}

	tile_format := tile.TileFormat(t.getFormat())
	if tile_format == "" {
		tile_format = tile.TileFormat(*tile_request.Format)
	}
	resp := NewResponse(t.getBuffer(), -1, tile_format.MimeType())
	if t.getCacheable() {
		resp.cacheHeaders(t.getTimestamp(), []string{t.getTimestamp().String(), strconv.Itoa(t.getSize())}, int(s.MaxTileAge.Seconds()))
	} else {
		resp.noCacheHeaders()
	}

	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *TileService) internalLayer(tile_request *request.TileRequest) Provider {
	var name string
	if v, ok := tile_request.Dimensions["_layer_spec"]; ok {
		name = tile_request.Layer + "_" + v[0]
	} else {
		name = tile_request.Layer
	}

	if l, ok := s.Layers[name]; ok {
		return l
	}

	if l, ok := s.Layers[name+"_EPSG900913"]; ok {
		return l
	}

	if l, ok := s.Layers[name+"_EPSG4326"]; ok {
		return l
	}
	return nil
}

func (s *TileService) internalDimensionLayer(tile_request *request.TileRequest) Provider {
	var name string
	if v, ok := tile_request.Dimensions["_layer_spec"]; ok {
		name = tile_request.Layer + "_" + v[0]
	} else {
		name = tile_request.Layer
	}
	if l, ok := s.Layers[name]; ok {
		return l
	}
	return nil
}

func (s *TileService) getLayer(tile_request *request.TileRequest) (Provider, geo.Coverage, *RequestError) {
	var internal_layer Provider
	if s.UseDimensionLayers {
		internal_layer = s.internalDimensionLayer(tile_request)
	} else {
		internal_layer = s.internalLayer(tile_request)
	}
	if internal_layer == nil {
		return nil, nil, NewRequestError(fmt.Sprintf("unknown layer: %s", tile_request.Layer), "", &TMSExceptionHandler{}, tile_request, false, nil)
	}

	limit_to := s.authorizeTileLayer(internal_layer, tile_request)
	return internal_layer, limit_to, nil
}

func (s *TileService) authorizeTileLayer(_ Provider, _ *request.TileRequest) geo.Coverage {
	return nil
}

func (s *TileService) authorizedTileLayers() []Provider {
	ret := []Provider{}
	for _, v := range s.Layers {
		ret = append(ret, v)
	}
	return ret
}

func (s *TileService) GetCapabilities(req request.Request) *Response {
	var tile_request *request.TileRequest
	switch r := req.(type) {
	case *request.TileRequest:
		tile_request = r
	case *request.TMSRequest:
		tile_request = &r.TileRequest
	}
	service := s.serviceMetadata(tile_request)
	var result []byte
	if tile_request.Layer != "" {
		layer, _, err := s.getLayer(tile_request)

		if err != nil {
			return err.Render()
		}

		result = s.renderGetLayer(layer, &service)
	} else {
		layer := s.authorizedTileLayers()
		result = s.renderCapabilities(layer, service)
	}
	return NewResponse(result, 200, "text/xml")
}

func (s *TileService) serviceMetadata(tms_request *request.TileRequest) TileMetadata {
	md := *s.Metadata
	url := tms_request.Http.URL
	url.RawQuery = ""
	url.Fragment = ""
	md.URL = url.String()
	return md
}

func (s *TileService) renderCapabilities(layer []Provider, service TileMetadata) []byte {
	ser := tms100.TileMapService{Version: "1.0.0"}

	ser.Title = service.Title
	ser.Abstract = service.Abstract

	baseUrl := service.URL

	i := strings.LastIndex(baseUrl, "/")

	if i != -1 {
		baseUrl = baseUrl[:i]
	}
	for i := range layer {
		l := layer[i].(*TileProvider)
		tp := tms100.TileMapInfo{Title: l.GetName(), Srs: l.GetSrs().GetDef(), Profile: l.grid.profile, Href: baseUrl + "/" + l.GetName()}
		ser.TileMaps.TileMap = append(ser.TileMaps.TileMap, tp)
	}

	return ser.ToXML()
}

func (s *TileService) renderGetLayer(layer Provider, service *TileMetadata) []byte {
	tm := &tms100.TileMap{Version: "1.0.0", Title: layer.GetName(), Abstract: "", SRS: layer.GetGrid().Srs.GetDef()}
	bounds := &tm.BoundingBox
	bbox := layer.GetBBox()

	bounds.MinX = strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	bounds.MinY = strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)
	bounds.MaxX = strconv.FormatFloat(bbox.Max[0], 'f', -1, 64)
	bounds.MaxY = strconv.FormatFloat(bbox.Max[1], 'f', -1, 64)

	tm.Origin.X = strconv.FormatFloat(bbox.Min[0], 'f', -1, 64)
	tm.Origin.Y = strconv.FormatFloat(bbox.Min[1], 'f', -1, 64)

	format := &tm.TileFormat
	format.Width = strconv.Itoa(int(layer.GetGrid().TileSize[0]))
	format.Height = strconv.Itoa(int(layer.GetGrid().TileSize[1]))
	format.MimeType = layer.GetFormatMimeType()
	format.Extension = layer.GetFormat()

	tp := layer.(*TileProvider)
	tm.TileSets.Profile = tp.grid.profile

	tileset := tp.grid.GetTileSets()
	for level, res := range tileset {
		ssres := strconv.FormatFloat(res, 'f', -1, 64)
		ts := tms100.TileSet{Href: service.URL + "/" + strconv.Itoa(level), PPM: ssres, Order: strconv.Itoa(level)}
		tm.TileSets.TileSet = append(tm.TileSets.TileSet, ts)
	}
	return tm.ToXML()
}

func (s *TileService) renderRootResource(service *TileMetadata) []byte {
	ser := &tms100.Services{}
	tms := tms100.ServiceMapinfo{Version: "1.0.0"}

	tms.Title = service.Title

	var url string
	if service.URL != "" {
		url = service.URL
		i := strings.LastIndex(url, "/")
		if i == -1 {
			url = url + "/"
		}
		tms.Href = url + "1.0.0/"
	}

	ser.TileMapService = append(ser.TileMapService, tms)
	return ser.ToXML()
}

func (s *TileService) RootResource(req request.Request) *Response {
	var tile_request *request.TileRequest
	switch r := req.(type) {
	case *request.TileRequest:
		tile_request = r
	case *request.TMSRequest:
		tile_request = &r.TileRequest
	}
	service := s.serviceMetadata(tile_request)
	result := s.renderRootResource(&service)
	return NewResponse(result, 200, "text/xml")
}

type TileServiceGrid struct {
	srs              geo.Proj
	grid             *geo.TileGrid
	profile          string
	srs_name         string
	skip_first_level bool
	skip_odd_level   bool
}

func NewTileServiceGrid(grid *geo.TileGrid) *TileServiceGrid {
	ret := &TileServiceGrid{}
	ret.grid = grid
	ret.srs = grid.Srs
	ret.profile = ""
	if grid.Srs.GetSrsCode() == "EPSG:900913" && geo.BBoxEquals(*grid.BBox, geo.DEFAULT_SRS_BBOX["EPSG:900913"], 0.0001, 0.0001) {
		ret.profile = "global-mercator"
		ret.srs_name = "OSGEO:41001"
		ret.skip_first_level = true
	} else if grid.Srs.GetSrsCode() == "EPSG:4326" && geo.BBoxEquals(*grid.BBox, geo.DEFAULT_SRS_BBOX["EPSG:4326"], 0.0001, 0.0001) {
		ret.profile = "global-geodetic"
		ret.srs_name = "EPSG:4326"
		ret.skip_first_level = true
	} else {
		ret.profile = "local"
		ret.srs_name = grid.Srs.GetSrsCode()
		ret.skip_first_level = false
	}

	ret.skip_odd_level = false

	res_factor := grid.Resolutions[0] / grid.Resolutions[1]
	if res_factor == math.Sqrt(2) {
		ret.skip_odd_level = true
	}
	return ret
}

func (t *TileServiceGrid) internalLevel(level int) int {
	if t.skip_first_level {
		level += 1
		if t.skip_odd_level {
			level += 1
		}
	}
	if t.skip_odd_level {
		level *= 2
	}
	return level
}

func (t *TileServiceGrid) GetOrigin() string {
	return geo.OriginToString(t.grid.Origin)
}

func (t *TileServiceGrid) GetBBox() vec2d.Rect {
	first_level := t.internalLevel(0)
	grid_size := t.grid.GridSizes[first_level]
	return t.grid.TilesBBox([][3]int{{0, 0, first_level}, {int(grid_size[0]) - 1, int(grid_size[1]) - 1, first_level}})
}

func (t *TileServiceGrid) GetTileSets() []float64 {
	tile_sets := []float64{}
	num_levels := t.grid.Levels
	start := 0
	step := 1
	if t.skip_first_level {
		if t.skip_odd_level {
			start = 2
		} else {
			start = 1
		}
	}
	if t.skip_odd_level {
		step = 2
	}
	for i := start; i < int(num_levels); i += step {
		tile_sets = append(tile_sets, t.grid.Resolutions[i])
	}
	return tile_sets
}

func (t *TileServiceGrid) InternalTileCoord(tile_coord [3]int, use_profiles bool) []int {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	if int(z) < 0 {
		return nil
	}
	if use_profiles && t.skip_first_level {
		z += 1
	}
	if t.skip_odd_level {
		z *= 2
	}
	return t.grid.LimitTile([3]int{x, y, z})
}

func (t *TileServiceGrid) ExternalTileCoord(tile_coord [3]int, use_profiles bool) []int {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	if int(z) < 0 {
		return nil
	}
	if use_profiles && t.skip_first_level {
		z -= 1
	}
	if t.skip_odd_level {
		z = int(float64(z / 2))
	}
	return []int{x, y, z}
}

type imageResponse struct {
	TileResponse
	img       []byte
	timestamp time.Time
	format    string
	size      int
	cacheable bool
}

func newImageResponse(img []byte, format string, timestamp time.Time) *imageResponse {
	return &imageResponse{img: img, timestamp: timestamp, format: format, size: 0, cacheable: true}
}

func (r *imageResponse) getBuffer() []byte {
	return r.img
}

func (r *imageResponse) getTimestamp() *time.Time {
	return &r.timestamp
}

func (r *imageResponse) getFormat() string {
	return r.format
}

func (r *imageResponse) getSize() int {
	return r.size
}

func (r *imageResponse) getCacheable() bool {
	return r.cacheable
}

type tileResponse struct {
	TileResponse
	buf       []byte
	tile      *cache.Tile
	timestamp *time.Time
	format    string
	size      int
	cacheable bool
}

func newTileResponse(tile *cache.Tile, format *tile.TileFormat, timestamp *time.Time, image_opts tile.TileOptions) *tileResponse {
	return &tileResponse{buf: tile.GetSourceBuffer(format, image_opts), tile: tile, timestamp: timestamp, size: int(tile.Size), cacheable: tile.Cacheable}
}

func (r *tileResponse) getBuffer() []byte {
	return r.buf
}

func (r *tileResponse) getTimestamp() *time.Time {
	return r.timestamp
}

func (r *tileResponse) getFormat() string {
	return r.format
}

func (r *tileResponse) GetFormatMime() string {
	tf := tile.TileFormat(r.format)
	return tf.MimeType()
}

func (r *tileResponse) getSize() int {
	return r.size
}

func (r *tileResponse) getCacheable() bool {
	return r.cacheable
}

func (r *tileResponse) peekFormat() string {
	return imagery.PeekImageFormat(string(r.buf))
}

type TileProviderMetadata struct {
	Name  string
	Title string
}

type TileProvider struct {
	Provider
	metadata     *TileProviderMetadata
	tileManager  cache.Manager
	infoSources  []layer.InfoLayer
	dimensions   utils.Dimensions
	grid         *TileServiceGrid
	extent       *geo.MapExtent
	emptyTile    []byte
	errorHandler ExceptionHandler
}

type TileProviderOptions struct {
	Name         string
	Title        string
	Metadata     *TileProviderMetadata
	TileManager  cache.Manager
	InfoSources  []layer.InfoLayer
	Dimensions   utils.Dimensions
	ErrorHandler ExceptionHandler
}

func NewTileProvider(opts *TileProviderOptions) *TileProvider {
	ret := &TileProvider{
		metadata:     opts.Metadata,
		tileManager:  opts.TileManager,
		infoSources:  opts.InfoSources,
		dimensions:   opts.Dimensions,
		grid:         NewTileServiceGrid(opts.TileManager.GetGrid()),
		extent:       geo.MapExtentFromGrid(opts.TileManager.GetGrid()),
		errorHandler: opts.ErrorHandler,
	}
	return ret
}

func (t *TileProvider) GetExtent() *geo.MapExtent {
	return t.extent
}

func (t *TileProvider) GetName() string {
	return t.metadata.Name
}

func (t *TileProvider) GetGrid() *geo.TileGrid {
	return t.grid.grid
}

func (t *TileProvider) GetBBox() vec2d.Rect {
	return t.grid.GetBBox()
}

func (t *TileProvider) GetSrs() geo.Proj {
	return t.grid.srs
}

func (t *TileProvider) GetFormatMimeType() string {
	if t.tileManager != nil {
		return t.tileManager.GetFormat()
	}
	return "image/png"
}

func (t *TileProvider) GetFormat() string {
	formats := request.SplitMimeType(t.GetFormatMimeType())
	return formats[1]
}

func (t *TileProvider) getInternalTileCoord(tileRequest request.TiledRequest, useProfiles bool) (*RequestError, []int) {
	tile := tileRequest.GetTile()
	origin := tileRequest.GetOriginString()
	tile_coord := t.grid.InternalTileCoord([3]int{tile[0], tile[1], tile[2]}, useProfiles)
	if tile_coord == nil {
		return NewRequestError("The requested tile is outside the bounding box  of the tile map.", "TileOutOfRange", t.errorHandler, tileRequest, false, nil), nil
	}
	if origin == "nw" && !utils.ContainsString([]string{"ul", "nw"}, t.grid.GetOrigin()) {
		coords := t.grid.grid.FlipTileCoord(tile_coord[0], tile_coord[1], tile_coord[2])
		tile_coord = coords[:]
	} else if origin == "sw" && !utils.ContainsString([]string{"ll", "sw"}, t.grid.GetOrigin()) {
		coords := t.grid.grid.FlipTileCoord(tile_coord[0], tile_coord[1], tile_coord[2])
		tile_coord = coords[:]
	}
	return nil, tile_coord
}

func (t *TileProvider) emptyResponse() TileResponse {
	format := t.GetFormat()
	if t.emptyTile == nil {
		si := t.grid.grid.TileSize
		tile := cache.GetEmptyTile([2]uint32{si[0], si[1]}, t.tileManager.GetTileOptions())
		t.emptyTile = tile.GetBuffer(nil, nil)
	}
	return newImageResponse(t.emptyTile, format, time.Now())
}

func (tl *TileProvider) GetTileBBox(req request.TiledRequest, useProfiles bool, limit bool) (*RequestError, vec2d.Rect) {
	err, tile_coord := tl.getInternalTileCoord(req, useProfiles)
	if err != nil {
		return err, vec2d.Rect{}
	}
	return nil, tl.grid.grid.TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, limit)
}

func (tl *TileProvider) checkedDimensions(_ request.TiledRequest) utils.Dimensions {
	dimensions := make(utils.Dimensions)
	for dimension, values := range tl.dimensions {
		dimensions[dimension] = values
	}
	return dimensions
}

func (tl *TileProvider) Render(req request.TiledRequest, useProfiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse) {
	format := req.GetFormat()
	if format == nil || format.Extension() != tl.GetFormat() {
		return NewRequestError(fmt.Sprintf("invalid format (%s). this tile set only supports (%s)", tl.GetFormat(), tl.GetFormat()), "InvalidParameterValue", tl.errorHandler, req, false, nil), nil
	}
	_, tile_coord := tl.getInternalTileCoord(req, useProfiles)
	var tile_bbox vec2d.Rect
	coverage_intersects := false
	if coverage != nil {
		tile_bbox = tl.grid.grid.TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, false)
		if coverage.Contains(tile_bbox, tl.grid.srs) {
			//
		} else if coverage.Intersects(tile_bbox, tl.grid.srs) {
			coverage_intersects = true
		} else {
			return nil, tl.emptyResponse()
		}
	}

	dimensions := tl.checkedDimensions(req)

	t, _ := tl.tileManager.LoadTileCoord([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, dimensions, true)
	if t.Source == nil {
		return nil, tl.emptyResponse()
	}

	if decorateTile != nil {
		t.Source = decorateTile(t.Source)
	}

	if coverage_intersects {
		format := tile.TileFormat(tl.GetFormat())
		tile_opts := t.Source.GetTileOptions()
		s, err := cache.MaskImageSourceFromCoverage(t.Source, tile_bbox, tl.grid.srs, coverage, tile_opts)
		if err != nil {
			return NewRequestError("mask image error", "InvalidParameterValue", tl.errorHandler, req, false, nil), nil
		}
		nt := cache.NewTile(t.Coord)
		nt.Source = s
		return nil, newTileResponse(nt, &format, nil, tl.tileManager.GetTileOptions())
	}

	return nil, newTileResponse(t, format, nil, tl.tileManager.GetTileOptions())
}

type TMSExceptionHandler struct {
	ExceptionHandler
}

func (h *TMSExceptionHandler) Render(err *RequestError) *Response {
	te := tms100.Exception{Message: err.Message}
	si, _ := xml.MarshalIndent(te, "", "")
	return NewResponse(si, 400, "text/xml")
}

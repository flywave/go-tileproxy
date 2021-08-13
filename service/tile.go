package service

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	_ "github.com/flywave/ogc-specifications/pkg/tms100"
)

var (
	TILE_SERVICE_NAMES           = []string{"tiles", "tms"}
	TILE_SERVICE_REQUEST_METHODS = []string{"map", "tms_capabilities", "tms_root_resource"}
)

type TileService struct {
	BaseService
	Layers             map[string]TileProvider
	Metadata           map[string]string
	MaxTileAge         *time.Duration
	UseDimensionLayers bool
	Origin             string
}

func NewTileService(layers map[string]TileProvider, md map[string]string, max_tile_age *time.Duration, use_dimension_layers bool, origin string) *TileService {
	return &TileService{Layers: layers, Metadata: md, MaxTileAge: max_tile_age, UseDimensionLayers: use_dimension_layers, Origin: origin}
}

func (s *TileService) GetMap(tile_request *request.TileRequest) *Response {
	if s.Origin != "" && tile_request.Origin == "" {
		tile_request.Origin = s.Origin
	}
	layer, limit_to := s.getLayer(tile_request)

	decorate_tile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*TileLayer)
		query_extent := &geo.MapExtent{Srs: tilelayer.grid.srs, BBox: layer.GetTileBBox(tile_request, tile_request.UseProfiles, false)}
		return s.DecorateTile(image, "tms", []string{tilelayer.name}, query_extent)
	}

	tile := layer.Render(tile_request, tile_request.UseProfiles, limit_to, decorate_tile)
	tile_format := tile.getFormat()
	if tile_format == "" {
		tile_format = string(*tile_request.Format)
	}
	resp := NewResponse(tile.getBuffer(), -1, "image/"+string(tile_format))
	if tile.getCacheable() {
		resp.cacheHeaders(tile.getTimestamp(), []string{tile.getTimestamp().String(), strconv.Itoa(tile.getSize())},
			int(s.MaxTileAge.Seconds()))
	} else {
		resp.noCacheHeaders()
	}

	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *TileService) internalLayer(tile_request *request.TileRequest) TileProvider {
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

func (s *TileService) internalDimensionLayer(tile_request *request.TileRequest) TileProvider {
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

func (s *TileService) getLayer(tile_request *request.TileRequest) (TileProvider, geo.Coverage) {
	var internal_layer TileProvider
	if s.UseDimensionLayers {
		internal_layer = s.internalDimensionLayer(tile_request)
	} else {
		internal_layer = s.internalLayer(tile_request)
	}
	if internal_layer == nil {
		return nil, nil
	}

	limit_to := s.authorizeTileLayer(internal_layer, tile_request)
	return internal_layer, limit_to
}

func (s *TileService) authorizeTileLayer(tile_layer TileProvider, tile_request *request.TileRequest) geo.Coverage {
	return nil
}

func (s *TileService) authorizedTileLayers() []TileProvider {
	ret := []TileProvider{}
	for _, v := range s.Layers {
		ret = append(ret, v)
	}
	return ret
}

func (s *TileService) GetCapabilities(tms_request *request.TileRequest) *Response {
	service := s.serviceMetadata(tms_request)
	var result []byte
	if tms_request.Layer != "" {
		layer, _ := s.getLayer(tms_request)
		result = s.renderGetLayer([]TileProvider{layer}, service)
	} else {
		layer := s.authorizedTileLayers()
		result = s.renderCapabilities(layer, service)
	}
	return NewResponse(result, 200, "text/xml")
}

func (s *TileService) serviceMetadata(tms_request *request.TileRequest) map[string]string {
	md := s.Metadata
	md["url"] = tms_request.Http.URL.Host
	return md
}

func (s *TileService) renderCapabilities(layer []TileProvider, service map[string]string) []byte {
	return nil
}

func (s *TileService) renderGetLayer(layer []TileProvider, service map[string]string) []byte {
	return nil
}

func (s *TileService) renderRootResource(service map[string]string) []byte {
	return nil
}

func (s *TileService) RootResource(tms_request *request.TileRequest) *Response {
	service := s.serviceMetadata(tms_request)
	result := s.renderRootResource(service)
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
	ret.profile = ""
	if grid.Srs.SrsCode == "EPSG:900913" && geo.BBoxEquals(*grid.BBox, geo.DEFAULT_SRS_BBOX["EPSG:900913"], 0.0001, 0.0001) {
		ret.profile = "global-mercator"
		ret.srs_name = "OSGEO:41001"
		ret.skip_first_level = true
	} else if grid.Srs.SrsCode == "EPSG:4326" && geo.BBoxEquals(*grid.BBox, geo.DEFAULT_SRS_BBOX["EPSG:4326"], 0.0001, 0.0001) {
		ret.profile = "global-geodetic"
		ret.srs_name = "EPSG:4326"
		ret.skip_first_level = true
	} else {
		ret.profile = "local"
		ret.srs_name = grid.Srs.SrsCode
		ret.skip_first_level = false
	}

	ret.skip_odd_level = false

	res_factor := grid.Resolutions[0] / grid.Resolutions[1]
	if res_factor == math.Sqrt(2) {
		ret.skip_odd_level = true
	}
	return ret
}

func (t *TileServiceGrid) internal_level(level int) int {
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
	first_level := t.internal_level(0)
	grid_size := t.grid.GridSizes[first_level]
	return t.grid.TilesBBox([][3]int{{0, 0, first_level},
		{int(grid_size[0]) - 1, int(grid_size[1]) - 1, first_level}})

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
		z = int(math.Floor(float64(z / 2)))
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
	return &tileResponse{buf: tile.GetSourceBuffer(format, image_opts), tile: tile, timestamp: &tile.Timestamp, size: int(tile.Size), cacheable: tile.Cacheable}
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

func (r *tileResponse) getSize() int {
	return r.size
}

func (r *tileResponse) getCacheable() bool {
	return r.cacheable
}

func (r *tileResponse) peekFormat() string {
	return imagery.PeekImageFormat(string(r.buf))
}

type TileLayer struct {
	TileProvider
	name                  string
	title                 string
	metadata              map[string]string
	tileManager           cache.Manager
	infoSources           []layer.Layer
	dimensions            utils.Dimensions
	grid                  *TileServiceGrid
	extent                *geo.MapExtent
	empty_tile            []byte
	mixed_format          bool
	empty_response_as_png bool
}

func NewTileLayer(name string, title string, md map[string]string, tile_manager cache.Manager, info_sources []layer.Layer, dimensions utils.Dimensions) *TileLayer {
	ret := &TileLayer{name: name, title: title, metadata: md, tileManager: tile_manager, infoSources: info_sources, dimensions: dimensions, grid: NewTileServiceGrid(tile_manager.GetGrid()), extent: geo.MapExtentFromGrid(tile_manager.GetGrid()), mixed_format: true, empty_response_as_png: true}
	if v, ok := ret.metadata["format"]; ok {
		if strings.ToLower(v) == "true" {
			ret.mixed_format = true
		} else {
			ret.mixed_format = false
		}
	}
	return ret
}

func (t *TileLayer) GetName() string {
	return t.name
}

func (t *TileLayer) GetGrid() *geo.TileGrid {
	return t.grid.grid
}

func (t *TileLayer) GetBBox() vec2d.Rect {
	return t.grid.GetBBox()
}

func (t *TileLayer) GetSrs() geo.Proj {
	return t.grid.srs
}

func (t *TileLayer) getFormatMimeType() string {
	if t.mixed_format {
		return "image/png"
	}
	if f, ok := t.metadata["format"]; ok {
		return f
	}
	return "image/png"
}

func (t *TileLayer) GetFormat() string {
	formats := request.SplitMimeType(t.getFormatMimeType())
	return formats[1]
}

func (t *TileLayer) getInternalTileCoord(tile_request *request.TileRequest, use_profiles bool) (error, []int) {
	tile_coord := t.grid.InternalTileCoord([3]int{tile_request.Tile[0], tile_request.Tile[1], tile_request.Tile[2]}, use_profiles)
	if tile_coord == nil {
		return errors.New("The requested tile is outside the bounding box  of the tile map."), nil
	}
	if tile_request.Origin == "nw" && !utils.ContainsString([]string{"ul", "nw"}, t.grid.GetOrigin()) {
		x, y, z := t.grid.grid.FlipTileCoord(tile_coord[0], tile_coord[1], tile_coord[2])
		tile_coord = []int{x, y, z}
	} else if tile_request.Origin == "sw" && !utils.ContainsString([]string{"ll", "sw"}, t.grid.GetOrigin()) {
		x, y, z := t.grid.grid.FlipTileCoord(tile_coord[0], tile_coord[1], tile_coord[2])
		tile_coord = []int{x, y, z}
	}
	return nil, tile_coord
}

func (t *TileLayer) empty_response() TileResponse {
	var format string
	if t.empty_response_as_png {
		format = "png"
	} else {
		format = t.GetFormat()
	}
	if t.empty_tile == nil {
		si := t.grid.grid.TileSize
		img := imagery.NewBlankImageSource([2]uint32{si[0], si[1]}, &imagery.ImageOptions{Format: tile.TileFormat(format), Transparent: geo.NewBool(true)}, nil)
		t.empty_tile = img.GetBuffer(nil, nil)
	}
	return newImageResponse(t.empty_tile, format, time.Now())
}

func (tl *TileLayer) GetTileBBox(req request.Request, use_profiles bool, limit bool) vec2d.Rect {
	tile_request := req.(*request.TileRequest)
	_, tile_coord := tl.getInternalTileCoord(tile_request, use_profiles)
	return tl.grid.grid.TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, limit)
}

func (tl *TileLayer) checkedDimensions(request *request.TileRequest) utils.Dimensions {
	dimensions := make(utils.Dimensions)
	for dimension, values := range tl.dimensions {
		dimensions[dimension] = values
	}
	return dimensions
}

func (tl *TileLayer) Render(req request.Request, use_profiles bool, coverage geo.Coverage, decorate_tile func(image tile.Source) tile.Source) TileResponse {
	tile_request := req.(*request.TileRequest)
	if string(*tile_request.Format) != tl.GetFormat() {
		return nil
	}
	_, tile_coord := tl.getInternalTileCoord(tile_request, use_profiles)
	var tile_bbox vec2d.Rect
	coverage_intersects := false
	if coverage != nil {
		tile_bbox = tl.grid.grid.TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, false)
		if coverage.Contains(tile_bbox, tl.grid.srs) {
			//
		} else if coverage.Intersects(tile_bbox, tl.grid.srs) {
			coverage_intersects = true
		} else {
			return tl.empty_response()
		}
	}

	dimensions := tl.checkedDimensions(tile_request)

	_, t := tl.tileManager.LoadTileCoord([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, dimensions, true)
	if t.Source == nil {
		return tl.empty_response()
	}

	if decorate_tile != nil {
		t.Source = decorate_tile(t.Source)
	}
	var format *tile.TileFormat
	var image_opts *imagery.ImageOptions
	if coverage_intersects {
		if tl.empty_response_as_png {
			tf := tile.TileFormat("png")
			format = &tf
			image_opts = &imagery.ImageOptions{Transparent: geo.NewBool(true), Format: tile.TileFormat("png")}
		} else {
			tf := tile.TileFormat(tl.GetFormat())
			format = &tf
			image_opts = t.Source.GetTileOptions().(*imagery.ImageOptions)
		}

		t.Source = imagery.MaskImageSourceFromCoverage(
			t.Source, tile_bbox, tl.grid.srs, coverage, image_opts)

		return newTileResponse(t, format, nil, image_opts)
	}
	if tl.mixed_format {
		format = nil
	} else {
		format = tile_request.Format
	}
	return newTileResponse(t, format, nil, tl.tileManager.GetTileOptions())
}

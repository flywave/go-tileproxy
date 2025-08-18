package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxMetadata struct {
	Name string
	URL  string
}

type MapboxService struct {
	BaseService
	Tilesets   map[string]Provider
	Metadata   *MapboxMetadata
	MaxTileAge *time.Duration
}

type MapboxServiceOptions struct {
	Tilesets   map[string]Provider
	Metadata   *MapboxMetadata
	MaxTileAge *time.Duration
}

func NewMapboxService(opts *MapboxServiceOptions) *MapboxService {
	s := &MapboxService{
		Tilesets:   opts.Tilesets,
		Metadata:   opts.Metadata,
		MaxTileAge: opts.MaxTileAge,
	}
	if s.MaxTileAge == nil {
		max := time.Duration(math.MaxInt64)
		s.MaxTileAge = &max
	}
	s.router = map[string]func(r request.Request) *Response{
		"source.json": func(r request.Request) *Response {
			return s.GetTileJSON(r)
		},
		"tile": func(r request.Request) *Response {
			return s.GetTile(r)
		},
	}
	s.requestParser = func(r *http.Request) request.Request {
		return request.MakeMapboxRequest(r, false)
	}
	return s
}

func (s *MapboxService) GetTileJSON(req request.Request) *Response {
	tilejson_request := req.(*request.MapboxSourceJSONRequest)
	err, layer := s.getLayer(tilejson_request.LayerName, req)
	if err != nil {
		return err.Render()
	}
	tilelayer := layer.(*MapboxTileProvider)
	var data []byte
	switch tilejson_request.FileName {
	case "source":
		data = tilelayer.RenderTileJson(tilejson_request)
	case "tilestats":
		data = tilelayer.RenderTileStats(tilejson_request)
		if data == nil {
			st := resource.NewTileStats(tilejson_request.LayerName)
			data = st.ToJson()
		}
	}
	resp := NewResponse(data, 200, "application/json")
	return resp
}

func (s *MapboxService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.MapboxTileRequest)
	if tile_request.Origin == "" {
		tile_request.Origin = "nw"
	}
	err, layer := s.getLayer(tile_request.LayerName, req)
	if err != nil {
		return err.Render()
	}
	var format tile.TileFormat
	if tile_request.Format != nil {
		format = *tile_request.Format
	} else {
		format = tile.TileFormat("image/webp")
	}

	if layer == nil {
		return NewResponse(nil, 404, format.MimeType())
	}

	decorateTile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*MapboxTileProvider)
		cerr, bbox := layer.GetTileBBox(tile_request, false, false)
		if cerr != nil {
			return nil
		}
		query_extent := &geo.MapExtent{Srs: tilelayer.GetSrs(), BBox: bbox}
		return s.DecorateTile(image, "mapbox", []string{tilelayer.name}, query_extent)
	}

	err, t := layer.Render(tile_request, false, nil, decorateTile)
	if err != nil {
		return err.Render()
	}
	tile_format := tile.TileFormat(t.getFormat())
	if tile_format == "" {
		tile_format = tile.TileFormat(*tile_request.Format)
	}
	resp := NewResponse(t.getBuffer(), 200, tile_format.MimeType())
	if t.getCacheable() {
		resp.cacheHeaders(t.getTimestamp(), []string{t.getTimestamp().String(), strconv.Itoa(t.getSize())}, int(s.MaxTileAge.Seconds()))
	} else {
		resp.noCacheHeaders()
	}

	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *MapboxService) getLayer(id string, req request.Request) (*RequestError, Provider) {
	if l, ok := s.Tilesets[id]; ok {
		return nil, l
	}
	return NewRequestError(fmt.Sprintf("Tileset %s does not exist", id), "Tileset_Not_Exist", &MapboxExceptionHandler{}, req, false, nil), nil
}

type MapboxTileType uint32

const (
	MapboxVector    MapboxTileType = 0
	MapboxRaster    MapboxTileType = 1
	MapboxRasterDem MapboxTileType = 2
)

type MapboxLayerMetadata struct {
	Name        string
	Attribution *string
	Description *string
	Legend      *string
	FillZoom    *uint32
	URL         string
}

type MapboxTileProvider struct {
	Provider
	name            string
	metadata        *MapboxLayerMetadata
	tileManager     cache.Manager
	extent          *geo.MapExtent
	empty_tile      []byte
	type_           MapboxTileType
	zoomRange       [2]int
	tilejsonSource  layer.MapboxSourceJSONLayer
	tileStatsSource layer.MapboxTileStatsLayer
	vectorLayers    []*resource.VectorLayer
}

func GetMapboxTileType(tp string) MapboxTileType {
	switch tp {
	case "vector":
		return MapboxVector
	case "raster":
		return MapboxRaster
	case "raster-dem":
		return MapboxRasterDem
	}
	return MapboxVector
}

func MapboxTileTypeToString(tp MapboxTileType) string {
	switch tp {
	case MapboxVector:
		return "vector"
	case MapboxRaster:
		return "raster"
	case MapboxRasterDem:
		return "raster-dem"
	}
	return "vector"
}

type MapboxTileOptions struct {
	Name            string
	Type            MapboxTileType
	Metadata        *MapboxLayerMetadata
	TileManager     cache.Manager
	TilejsonSource  layer.MapboxSourceJSONLayer
	TileStatsSource layer.MapboxTileStatsLayer
	VectorLayers    []*resource.VectorLayer
	ZoomRange       *[2]int
}

func NewMapboxTileProvider(opts *MapboxTileOptions) *MapboxTileProvider {
	ret := &MapboxTileProvider{
		name:            opts.Name,
		type_:           opts.Type,
		metadata:        opts.Metadata,
		tileManager:     opts.TileManager,
		extent:          geo.MapExtentFromGrid(opts.TileManager.GetGrid()),
		tilejsonSource:  opts.TilejsonSource,
		tileStatsSource: opts.TileStatsSource,
		vectorLayers:    opts.VectorLayers,
	}
	if opts.ZoomRange != nil {
		ret.zoomRange = *opts.ZoomRange
	} else {
		ret.zoomRange = [2]int{0, 20}
	}
	return ret
}

func (t *MapboxTileProvider) GetMaxZoom() int {
	return t.zoomRange[1]
}

func (t *MapboxTileProvider) GetMinZoom() int {
	return t.zoomRange[0]
}

func (t *MapboxTileProvider) GetExtent() *geo.MapExtent {
	return t.extent
}

func (t *MapboxTileProvider) GetName() string {
	return t.name
}

func (t *MapboxTileProvider) GetGrid() *geo.TileGrid {
	return t.tileManager.GetGrid()
}

func (t *MapboxTileProvider) GetBBox() vec2d.Rect {
	return *t.GetGrid().BBox
}

func (t *MapboxTileProvider) GetLonlatBBox() vec2d.Rect {
	bbx := *t.GetGrid().BBox
	srs := t.GetGrid().Srs
	dest := geo.NewProj("EPSG:4326")
	ps := srs.TransformTo(dest, []vec2d.T{bbx.Min, bbx.Max})
	return vec2d.Rect{Min: ps[0], Max: ps[1]}
}

func (t *MapboxTileProvider) GetSrs() geo.Proj {
	return t.GetGrid().Srs
}

func (t *MapboxTileProvider) IsVector() bool {
	return t.type_ == MapboxVector
}

func (t *MapboxTileProvider) IsRaster() bool {
	return t.type_ == MapboxRaster
}

func (t *MapboxTileProvider) IsRasterDem() bool {
	return t.type_ == MapboxRasterDem
}

func (t *MapboxTileProvider) GetFormatMimeType() string {
	format := tile.TileFormat(t.tileManager.GetRequestFormat())
	return format.MimeType()
}

func (t *MapboxTileProvider) GetFormat() string {
	return t.tileManager.GetFormat()
}

func (t *MapboxTileProvider) GetTileBBox(req request.TiledRequest, useProfiles bool, limit bool) (*RequestError, vec2d.Rect) {
	tileRequest := req.(*request.TileRequest)
	tile_coord := tileRequest.Tile
	return nil, t.GetGrid().TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, limit)
}

func (t *MapboxTileProvider) emptyResponse() TileResponse {
	format := t.GetFormat()
	if t.empty_tile == nil {
		si := t.GetGrid().TileSize
		tile := cache.GetEmptyTile([2]uint32{si[0], si[1]}, t.tileManager.GetTileOptions())
		t.empty_tile = tile.GetBuffer(nil, nil)
	}
	return newImageResponse(t.empty_tile, format, time.Now())
}

func (tl *MapboxTileProvider) Render(req request.TiledRequest, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse) {
	tile_request := req.(*request.MapboxTileRequest)
	if string(*tile_request.Format) != tl.GetFormat() {
		return NewRequestError("Not Found", "Not_Found", &MapboxExceptionHandler{}, tile_request, false, nil), nil
	}

	tile_coord := tile_request.Tile

	if tl.GetMaxZoom() < tile_coord[2] || tl.GetMinZoom() > tile_coord[2] {
		return NewRequestError("Zoom out of range", "Zoom out of range", &MapboxExceptionHandler{}, tile_request, false, nil), nil
	}

	var tile_bbox vec2d.Rect
	if coverage != nil {
		tile_bbox = tl.GetGrid().TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, false)
		if coverage.Contains(tile_bbox, tl.GetGrid().Srs) {
			//
		} else if coverage.Intersects(tile_bbox, tl.GetGrid().Srs) {
			//
		} else {
			return nil, tl.emptyResponse()
		}
	}

	t, _ := tl.tileManager.LoadTileCoord([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, nil, true)
	if t.Source == nil {
		return nil, tl.emptyResponse()
	}
	format := tile_request.Format
	return nil, newTileResponse(t, format, nil, tl.tileManager.GetTileOptions())
}

func (c *MapboxTileProvider) convertTileJson(tilejson *resource.TileJSON, req *request.MapboxSourceJSONRequest, md *MapboxLayerMetadata) []byte {
	url := req.LayerName + "/{z}/{x}/{y}." + c.GetFormat()
	url = strings.ReplaceAll(url, "//", "/")
	url = md.URL + url
	tilejson.Tiles = []string{url}

	if len(tilejson.VectorLayers) > 0 {
		tilejson.Type = resource.VECTOR
	}
	if tilejson.Type == "" {
		if tilejson.ID == resource.MAPBOX_STATELLITE {
			tilejson.Type = resource.RASTER
		} else {
			tilejson.Type = resource.RASTER_DEM
		}
	}
	return tilejson.GetData()
}

func (c *MapboxTileProvider) serviceMetadata(tms_request *request.MapboxSourceJSONRequest) MapboxLayerMetadata {
	md := *c.metadata
	strs := strings.Split(tms_request.Http.RequestURI, tms_request.LayerName)
	md.URL = c.tileManager.SiteURL() + strs[0]
	return md
}

func (c *MapboxTileProvider) RenderTileStats(req *request.MapboxSourceJSONRequest) []byte {
	if c.tileStatsSource != nil {
		states := c.tileStatsSource.GetTileStats(req.LayerName)
		data, err := json.Marshal(states)
		if err != nil {
			return nil
		}
		return data
	}
	return nil
}

func (c *MapboxTileProvider) RenderTileJson(req *request.MapboxSourceJSONRequest) []byte {
	md := c.serviceMetadata(req)
	if c.tilejsonSource != nil {
		styles := c.tilejsonSource.GetTileJSON(req.LayerName)
		styles.Name = c.name
		return c.convertTileJson(styles, req, &md)
	}

	tilejson := &resource.TileJSON{Type: MapboxTileTypeToString(c.type_)}

	bbox := c.GetLonlatBBox()

	tilejson.Bounds[0], tilejson.Bounds[1], tilejson.Bounds[2], tilejson.Bounds[3] = float32(bbox.Min[0]), float32(bbox.Min[1]), float32(bbox.Max[0]), float32(bbox.Max[1])
	tilejson.Center[0], tilejson.Center[1], tilejson.Center[2] = float32(bbox.Min[0]+(bbox.Max[0]-bbox.Min[0])/2), float32(bbox.Min[1]+(bbox.Max[1]-bbox.Min[1])/2), 0

	tilejson.Format = c.GetFormat()
	tilejson.ID = req.LayerName

	tilejson.Name = md.Name
	tilejson.MinZoom = uint32(math.Min(float64(c.zoomRange[0]), float64(c.zoomRange[1])))
	tilejson.MaxZoom = uint32(math.Max(float64(c.zoomRange[0]), float64(c.zoomRange[1])))

	if md.Attribution != nil {
		tilejson.Attribution = *md.Attribution
	}
	if md.Description != nil {
		tilejson.Description = *md.Description
	}
	if md.Legend != nil {
		tilejson.Legend = md.Legend
	}
	if md.FillZoom != nil {
		tilejson.FillZoom = *md.FillZoom
	}

	tilejson.Scheme = "xyz"
	tilejson.Version = "1.0.0"
	tilejson.TilejsonVersion = "3.0.0"

	url := md.URL + req.LayerName + "/{z}/{x}/{y}." + c.GetFormat()

	tilejson.VectorLayers = c.vectorLayers[:]

	tilejson.Tiles = append(tilejson.Tiles, url)

	return tilejson.ToJson()
}

type MapboxException struct {
	Message string `json:"message"`
}

var (
	MapboxExceptionMessages = map[string]string{
		"Invalid_Range":          "Invalid Range",
		"Too_Many_Font":          "Maximum of 10 font faces permitted",
		"No_Token":               "Not Authorized - No Token",
		"Invalid_Token":          "Not Authorized - Invalid Token",
		"Forbidden":              "Forbidden",
		"Requires_Token":         "This endpoint requires a token with %s scope",
		"Tileset_Not_Exist":      "Tileset %s does not exist",
		"Tile_Not_Found":         "Tile not found",
		"Style_Not_Found":        "Style not found",
		"Not_Found":              "Not Found",
		"Invalid_Zoom_Level":     "Zoom level must be between 0-30.",
		"Tileset_Not_Ref_Vector": "Tileset does not reference vector data",
		"Invalid_Quality_Value":  "Invalid quality value %s for raster format %s",
	}
	MapboxExceptionCodes = map[string]int{
		"Invalid_Range":          400,
		"Too_Many_Font":          400,
		"No_Token":               401,
		"Invalid_Token":          401,
		"Forbidden":              403,
		"Requires_Token":         403,
		"Tileset_Not_Exist":      404,
		"Tile_Not_Found":         404,
		"Style_Not_Found":        404,
		"Not_Found":              404,
		"Invalid_Zoom_Level":     422,
		"Tileset_Not_Ref_Vector": 422,
		"Invalid_Quality_Value":  422,
	}
)

type MapboxExceptionHandler struct {
	ExceptionHandler
}

func (h *MapboxExceptionHandler) Render(request_error *RequestError) *Response {
	status_code := 500
	if sc, ok := MapboxExceptionCodes[request_error.Code]; ok {
		status_code = sc
	}
	return NewResponse([]byte(request_error.Message), status_code, "application/json")
}

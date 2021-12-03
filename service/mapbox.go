package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/style"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxMetadata struct {
	Name string
	URL  string
}

type MapboxService struct {
	BaseService
	Tilesets   map[string]Provider
	Styles     map[string]*StyleProvider
	Fonts      map[string]*StyleProvider
	Metadata   *MapboxMetadata
	MaxTileAge *time.Duration
}

type MapboxServiceOptions struct {
	Tilesets   map[string]Provider
	Styles     map[string]*StyleProvider
	Metadata   *MapboxMetadata
	MaxTileAge *time.Duration
}

func NewMapboxService(opts *MapboxServiceOptions) *MapboxService {
	s := &MapboxService{Tilesets: opts.Tilesets, Styles: opts.Styles, Fonts: make(map[string]*StyleProvider), Metadata: opts.Metadata, MaxTileAge: opts.MaxTileAge}
	s.router = map[string]func(r request.Request) *Response{
		"tilejson": func(r request.Request) *Response {
			return s.GetTileJSON(r)
		},
		"tile": func(r request.Request) *Response {
			return s.GetTile(r)
		},
		"style": func(r request.Request) *Response {
			return s.GetStyle(r)
		},
		"sprite": func(r request.Request) *Response {
			return s.GetSprite(r)
		},
		"glyphs": func(r request.Request) *Response {
			return s.GetGlyphs(r)
		},
	}
	for _, sty := range opts.Styles {
		for _, fontId := range sty.getFonts() {
			s.Fonts[fontId] = sty
		}
		sty.metadata = opts.Metadata
	}
	s.requestParser = func(r *http.Request) request.Request {
		return request.MakeMapboxRequest(r, false)
	}
	return s
}

func (s *MapboxService) GetTileJSON(req request.Request) *Response {
	tilejson_request := req.(*request.MapboxTileJSONRequest)
	err, layer := s.getLayer(tilejson_request.TilesetID, req)
	if err != nil {
		return err.Render()
	}
	tilelayer := layer.(*MapboxTileProvider)

	data := tilelayer.RenderTileJson(tilejson_request)

	resp := NewResponse(data, 200, "application/json")
	return resp
}

func (s *MapboxService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.MapboxTileRequest)
	if tile_request.Origin == "" {
		tile_request.Origin = "nw"
	}
	err, layer := s.getLayer(tile_request.TilesetID, req)
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
		err, bbox := layer.GetTileBBox(tile_request, false, false)
		if err != nil {
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

func (s *MapboxService) GetStyle(req request.Request) *Response {
	style_req := req.(*request.MapboxStyleRequest)

	if st, ok := s.Styles[style_req.StyleID]; ok {
		resp := st.fetch(style_req)
		return NewResponse(resp, 200, "application/json")
	}

	resp := NewRequestError("Style not found", "Style_Not_Found", &MapboxExceptionHandler{}, style_req, false, nil)
	return resp.Render()
}

func (s *MapboxService) GetSprite(req request.Request) *Response {
	sprite_req := req.(*request.MapboxSpriteRequest)

	if st, ok := s.Styles[sprite_req.StyleID]; ok {
		if sprite_req.Format == nil {
			resp := st.fetchSprite(sprite_req)
			return NewResponse(resp, 200, "application/json")
		} else {
			resp := st.fetchSprite(sprite_req)
			return NewResponse(resp, 200, sprite_req.Format.MimeType())
		}
	}

	resp := NewRequestError("Style not found", "Style_Not_Found", &MapboxExceptionHandler{}, sprite_req, false, nil)
	return resp.Render()
}

func (s *MapboxService) GetGlyphs(req request.Request) *Response {
	glyphs_req := req.(*request.MapboxGlyphsRequest)

	if st, ok := s.Fonts[glyphs_req.Font]; ok {
		resp := st.fetchGlyph(glyphs_req)
		return NewResponse(resp, 200, "application/x-protobuf")
	}

	resp := NewRequestError("Not found", "Not_Found", &MapboxExceptionHandler{}, glyphs_req, false, nil)
	return resp.Render()
}

type StyleProvider struct {
	metadata     *MapboxMetadata
	styleSource  *sources.MapboxStyleSource
	glyphsSource *sources.MapboxGlyphsSource
}

func NewStyleProvider(style *sources.MapboxStyleSource, glyphs *sources.MapboxGlyphsSource) *StyleProvider {
	return &StyleProvider{styleSource: style, glyphsSource: glyphs}
}

func (c *StyleProvider) getFonts() []string {
	return c.glyphsSource.Fonts
}

func (c *StyleProvider) serviceMetadata(tms_request *request.MapboxStyleRequest) MapboxMetadata {
	md := *c.metadata
	md.URL = tms_request.Http.URL.Host
	return md
}

func (c *StyleProvider) convertTileJson(style_ *resource.Style, req *request.MapboxStyleRequest) []byte {
	metadata := c.serviceMetadata(req)
	sprite_url := metadata.URL + "/v1/sprites/{style_id}"
	glyphs_url := metadata.URL + "/v1/fonts/{fontstack}/{range}.pbf"
	styleJson := style_.GetStyle()
	styleJson.Sprite = &sprite_url
	styleJson.Glyphs = &glyphs_url
	for k, data := range styleJson.Sources {
		var src style.Source
		dec := json.NewDecoder(bytes.NewBuffer(data))
		if err := dec.Decode(&src); err != nil {
			return nil
		}

		if strings.Contains(src.URL, "mapbox://") {
			src.URL = strings.ReplaceAll(src.URL, "mapbox://", metadata.URL+"/v4/")
		}

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode(src); err != nil {
			return nil
		}
		styleJson.Sources[k] = buf.Bytes()
	}
	return style_.GetData()
}

func (c *StyleProvider) fetch(req *request.MapboxStyleRequest) []byte {
	styles := c.styleSource.GetStyle(req.StyleID)
	return c.convertTileJson(styles, req)
}

func (c *StyleProvider) fetchSprite(req *request.MapboxSpriteRequest) []byte {
	styles := c.styleSource.GetSpriteJSON(req.StyleID)
	return styles.GetData()
}

func (c *StyleProvider) fetchGlyph(req *request.MapboxGlyphsRequest) []byte {
	query := &layer.GlyphsQuery{Start: req.Start, End: req.End, Font: req.Font}
	glyphs := c.glyphsSource.GetGlyphs(query)
	return glyphs.GetData()
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
	name           string
	metadata       *MapboxLayerMetadata
	tileManager    cache.Manager
	extent         *geo.MapExtent
	empty_tile     []byte
	type_          MapboxTileType
	zoomRange      [2]int
	tilejsonSource layer.MapboxTileJSONLayer
	vectorLayers   []*resource.VectorLayer
}

func GetMapboxTileType(tp string) MapboxTileType {
	if tp == "vector" {
		return MapboxVector
	} else if tp == "raster" {
		return MapboxRaster
	} else if tp == "rasterDem" {
		return MapboxRasterDem
	}
	return MapboxVector
}

type MapboxTileOptions struct {
	Name           string
	Type           MapboxTileType
	Metadata       *MapboxLayerMetadata
	TileManager    cache.Manager
	TilejsonSource layer.MapboxTileJSONLayer
	VectorLayers   []*resource.VectorLayer
	ZoomRange      *[2]int
}

func NewMapboxTileProvider(opts *MapboxTileOptions) *MapboxTileProvider {
	ret := &MapboxTileProvider{
		name:           opts.Name,
		type_:          opts.Type,
		metadata:       opts.Metadata,
		tileManager:    opts.TileManager,
		extent:         geo.MapExtentFromGrid(opts.TileManager.GetGrid()),
		tilejsonSource: opts.TilejsonSource,
		vectorLayers:   opts.VectorLayers,
	}
	if opts.ZoomRange != nil {
		ret.zoomRange = *opts.ZoomRange
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
	return t.tileManager.GetRequestFormat()
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

func (c *MapboxTileProvider) convertTileJson(tilejson *resource.TileJSON, req *request.MapboxTileJSONRequest) []byte {
	md := c.serviceMetadata(req)
	url := md.URL + "/v4/" + req.TilesetID + "/{z}/{x}/{y}." + c.GetFormat()
	tilejson.Tiles = append(tilejson.Tiles, url)
	return tilejson.GetData()
}

func (c *MapboxTileProvider) serviceMetadata(tms_request *request.MapboxTileJSONRequest) MapboxLayerMetadata {
	md := *c.metadata
	md.URL = tms_request.Http.URL.Host
	return md
}

func (c *MapboxTileProvider) RenderTileJson(req *request.MapboxTileJSONRequest) []byte {
	if c.tilejsonSource != nil {
		styles := c.tilejsonSource.GetTileJSON(req.TilesetID)
		return c.convertTileJson(styles, req)
	}
	md := c.serviceMetadata(req)

	tilejson := &resource.TileJSON{}

	bbox := c.GetBBox()

	tilejson.Bounds[0], tilejson.Bounds[1], tilejson.Bounds[2], tilejson.Bounds[3] = float32(bbox.Min[0]), float32(bbox.Min[1]), float32(bbox.Max[0]), float32(bbox.Max[1])
	tilejson.Center[0], tilejson.Center[1], tilejson.Center[2] = float32(bbox.Min[0]+(bbox.Max[0]-bbox.Min[0])/2), float32(bbox.Min[1]+(bbox.Max[1]-bbox.Min[1])/2), 0

	tilejson.Format = c.GetFormat()
	tilejson.ID = req.TilesetID

	tilejson.Name = md.Name

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

	url := md.URL + "/v4/" + req.TilesetID + "/{z}/{x}/{y}." + c.GetFormat()

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

package service

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxService struct {
	BaseService
	Tilesets   map[string]Provider
	Styles     map[string]*StyleProvider
	Fonts      map[string]*GlyphProvider
	Metadata   map[string]string
	MaxTileAge *time.Duration
}

func NewMapboxService(layers map[string]Provider, styles map[string]*StyleProvider, fonts map[string]*GlyphProvider, md map[string]string, max_tile_age *time.Duration) *MapboxService {
	s := &MapboxService{Tilesets: layers, Styles: styles, Fonts: fonts, Metadata: md, MaxTileAge: max_tile_age}
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
	s.requestParser = func(r *http.Request) request.Request {
		return request.MakeMapboxRequest(r, true)
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
	resp := NewResponse(t.getBuffer(), -1, tile_format.MimeType())
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
		resp := st.fetch(glyphs_req)
		return NewResponse(resp, 200, "application/x-protobuf")
	}

	resp := NewRequestError("Not found", "Not_Found", &MapboxExceptionHandler{}, glyphs_req, false, nil)
	return resp.Render()
}

type GlyphProvider struct {
	source *sources.MapboxGlyphsSource
}

func NewGlyphProvider(source *sources.MapboxGlyphsSource) *GlyphProvider {
	return &GlyphProvider{source: source}
}

func (c *GlyphProvider) fetch(req *request.MapboxGlyphsRequest) []byte {
	query := &layer.GlyphsQuery{Font: req.Font, Start: req.Start, End: req.End}
	glyphs := c.source.GetGlyphs(query)
	return glyphs.GetData()
}

type StyleProvider struct {
	source *sources.MapboxStyleSource
}

func NewStyleProvider(source *sources.MapboxStyleSource) *StyleProvider {
	return &StyleProvider{source: source}
}

func (c *StyleProvider) fetch(req *request.MapboxStyleRequest) []byte {
	query := &layer.StyleQuery{StyleID: req.StyleID}
	styles := c.source.GetStyle(query)
	return styles.GetData()
}

func (c *StyleProvider) fetchSprite(req *request.MapboxSpriteRequest) []byte {
	query := &layer.SpriteQuery{StyleQuery: layer.StyleQuery{StyleID: req.StyleID}}
	if req.Retina != nil {
		query.Retina = req.Retina
	}
	if req.Format != nil {
		query.Format = req.Format
		styles := c.source.GetSprite(query)
		return styles.GetData()
	}
	styles := c.source.GetSpriteJSON(query)
	return styles.GetData()
}

type MapboxTileType uint32

const (
	MapboxVector    MapboxTileType = 0
	MapboxRaster    MapboxTileType = 1
	MapboxRasterDem MapboxTileType = 2
)

type MapboxTileProvider struct {
	Provider
	name           string
	metadata       map[string]string
	tileManager    cache.Manager
	extent         *geo.MapExtent
	empty_tile     []byte
	type_          MapboxTileType
	zoomRange      [2]int
	tilejsonSource *sources.MapboxTileJSONSource
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

func NewMapboxTileProvider(name string, tp MapboxTileType, md map[string]string, tileManager cache.Manager, tilejsonSource *sources.MapboxTileJSONSource, vectorLayers []*resource.VectorLayer, zoomRange *[2]int) *MapboxTileProvider {
	ret := &MapboxTileProvider{name: name, type_: tp, metadata: md, tileManager: tileManager, extent: geo.MapExtentFromGrid(tileManager.GetGrid()), tilejsonSource: tilejsonSource, vectorLayers: vectorLayers}
	if zoomRange != nil {
		ret.zoomRange = *zoomRange
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
	if f, ok := t.metadata["format"]; ok {
		return f
	}
	if t.type_ == MapboxVector {
		return "application/vnd.mapbox-vector-tile"
	}
	return "image/webp"
}

func (t *MapboxTileProvider) GetFormat() string {
	formats := request.SplitMimeType(t.GetFormatMimeType())
	if formats[1] == "vnd.mapbox-vector-tile" {
		return "mvt"
	}
	return formats[1]
}

func (t *MapboxTileProvider) GetTileBBox(req request.Request, useProfiles bool, limit bool) (*RequestError, vec2d.Rect) {
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

func (tl *MapboxTileProvider) Render(req request.Request, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse) {
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

func (c *MapboxTileProvider) RenderTileJson(req *request.MapboxTileJSONRequest) []byte {
	if c.tilejsonSource != nil {
		query := &layer.TileJSONQuery{TilesetID: req.TilesetID}

		styles := c.tilejsonSource.GetTileJSON(query)
		return styles.GetData()
	}

	tilejson := &resource.TileJSON{}

	bbox := c.GetBBox()

	tilejson.Bounds[0], tilejson.Bounds[1], tilejson.Bounds[2], tilejson.Bounds[3] = float32(bbox.Min[0]), float32(bbox.Min[1]), float32(bbox.Max[0]), float32(bbox.Max[1])
	tilejson.Center[0], tilejson.Center[1], tilejson.Center[2] = float32(bbox.Min[0]+(bbox.Max[0]-bbox.Min[0])/2), float32(bbox.Min[1]+(bbox.Max[1]-bbox.Min[1])/2), 0

	tilejson.Format = c.GetFormat()
	tilejson.Id = req.TilesetID

	tilejson.Name = c.metadata["name"]

	if attr, ok := c.metadata["attribution"]; ok {
		tilejson.Attribution = attr
	}
	if desc, ok := c.metadata["description"]; ok {
		tilejson.Description = desc
	}
	if legend, ok := c.metadata["legend"]; ok {
		tilejson.Legend = &legend
	}
	if fillzoom, ok := c.metadata["fillzoom"]; ok {
		z, _ := strconv.Atoi(fillzoom)
		tilejson.FillZoom = uint32(z)
	}
	tilejson.Scheme = "xyz"
	tilejson.Version = "1.0.0"
	tilejson.TilejsonVersion = "3.0.0"

	url := c.metadata["url"] + "/" + req.Version + "/" + req.TilesetID + "/{z}/{x}/{y}." + c.GetFormat()

	tilejson.VectorLayers = c.vectorLayers[:]

	tilejson.Tiles = append(tilejson.Tiles, url)
	return tilejson.ToJson()
}

type mapboxException struct {
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

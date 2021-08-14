package service

import (
	"strconv"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxService struct {
	BaseService
	Tilesets   map[string]Provider
	Styles     map[string]*StyleProvider
	Fonts      map[string]*GlyphProvider
	Metadata   map[string]string
	MaxTileAge *time.Duration
	Origin     string
}

func (s *MapboxService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.MapboxTileRequest)
	if s.Origin != "" && tile_request.Origin == "" {
		tile_request.Origin = s.Origin
	}
	layer := s.getLayer(tile_request)
	var format tile.TileFormat
	if tile_request.Format != nil {
		format = *tile_request.Format
	} else {
		format = tile.TileFormat("image/png")
	}

	if layer == nil {
		return NewResponse(nil, 404, format.MimeType())
	}

	if tile_request.AccessToken == "" {
		if token, ok := s.Metadata["access_token"]; ok {
			tile_request.AccessToken = token
		} else {
			return NewResponse(nil, 401, format.MimeType())
		}
	}

	decorateTile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*TileProvider)
		query_extent := &geo.MapExtent{Srs: tilelayer.grid.srs, BBox: layer.GetTileBBox(tile_request, false, false)}
		return s.DecorateTile(image, "tms", []string{tilelayer.name}, query_extent)
	}

	t := layer.Render(tile_request, false, nil, decorateTile)
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

func (s *MapboxService) getLayer(tile_request *request.MapboxTileRequest) Provider {
	id := tile_request.TilesetID
	if l, ok := s.Tilesets[id]; ok {
		return l
	}
	return nil
}

func (s *MapboxService) GetStyle(req request.Request) *Response {
	style_req := req.(*request.MapboxStyleRequest)

	if style_req.AccessToken == "" {
		if token, ok := s.Metadata["access_token"]; ok {
			style_req.AccessToken = token
		} else {
			return NewResponse(nil, 401, "")
		}
	}

	if st, ok := s.Styles[style_req.StyleID]; ok {
		resp := st.fetch(style_req)
		return NewResponse(resp, 200, "application/json")
	}

	return NewResponse(nil, 404, "")
}

func (s *MapboxService) GetSprite(req request.Request) *Response {
	sprite_req := req.(*request.MapboxSpriteRequest)

	if sprite_req.AccessToken == "" {
		if token, ok := s.Metadata["access_token"]; ok {
			sprite_req.AccessToken = token
		} else {
			return NewResponse(nil, 401, "")
		}
	}

	if st, ok := s.Styles[sprite_req.StyleID]; ok {
		if sprite_req.Format == nil {
			resp := st.fetchSprite(sprite_req)
			return NewResponse(resp, 200, "application/json")
		} else {
			resp := st.fetchSprite(sprite_req)
			return NewResponse(resp, 200, sprite_req.Format.MimeType())
		}
	}

	return NewResponse(nil, 404, "")
}

func (s *MapboxService) GetGlyphs(req request.Request) *Response {
	glyphs_req := req.(*request.MapboxGlyphsRequest)

	if st, ok := s.Fonts[glyphs_req.Font]; ok {
		resp := st.fetch(glyphs_req)
		return NewResponse(resp, 200, "application/x-protobuf")
	}

	return NewResponse(nil, 404, "")
}

type GlyphProvider struct {
}

func (c *GlyphProvider) fetch(req *request.MapboxGlyphsRequest) []byte {
	return nil
}

type StyleProvider struct {
}

func (c *StyleProvider) fetch(req *request.MapboxStyleRequest) []byte {
	return nil
}

func (c *StyleProvider) fetchSprite(req *request.MapboxSpriteRequest) []byte {
	return nil
}

type MapboxTileType uint32

const (
	MapboxVector    MapboxTileType = 0
	MapboxRaster    MapboxTileType = 1
	MapboxRasterDem MapboxTileType = 2
)

type MapboxTileProvider struct {
	Provider
	name        string
	title       string
	metadata    map[string]string
	tileManager cache.Manager
	extent      *geo.MapExtent
	empty_tile  []byte
	type_       MapboxTileType
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

func (t *MapboxTileProvider) getFormatMimeType() string {
	if f, ok := t.metadata["format"]; ok {
		return f
	}
	if t.type_ == MapboxVector {
		return "application/vnd.mapbox-vector-tile"
	}
	return "image/webp"
}

func (t *MapboxTileProvider) GetFormat() string {
	formats := request.SplitMimeType(t.getFormatMimeType())
	if formats[1] == "vnd.mapbox-vector-tile" {
		return "mvt"
	}
	return formats[1]
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

func (tl *MapboxTileProvider) Render(req request.Request, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) TileResponse {
	tile_request := req.(*request.MapboxTileRequest)
	if string(*tile_request.Format) != tl.GetFormat() {
		return nil
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
			return tl.emptyResponse()
		}
	}

	_, t := tl.tileManager.LoadTileCoord([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, nil, true)
	if t.Source == nil {
		return tl.emptyResponse()
	}
	format := tile_request.Format
	return newTileResponse(t, format, nil, tl.tileManager.GetTileOptions())
}

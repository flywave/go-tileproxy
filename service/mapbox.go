package service

import (
	"strconv"
	"time"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MapboxService struct {
	BaseService
	Tilesets   map[string]RenderLayer
	Styles     map[string]*MapboxStyle
	Fonts      map[string]*MapboxGlyph
	Metadata   map[string]string
	MaxTileAge *time.Duration
	Origin     string
}

func (s *MapboxService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.MapboxTileRequest)
	if s.Origin != "" && tile_request.Origin == "" {
		tile_request.Origin = s.Origin
	}
	layer := s.GetLayer(tile_request)
	var format string
	if tile_request.Format != nil {
		format = string(*tile_request.Format)
	} else {
		format = "image/png"
	}

	if layer == nil {
		return NewResponse(nil, 404, format)
	}

	if tile_request.AccessToken == "" {
		if token, ok := s.Metadata["access_token"]; ok {
			tile_request.AccessToken = token
		} else {
			return NewResponse(nil, 401, format)
		}
	}

	decorate_tile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*TileLayer)
		query_extent := &geo.MapExtent{Srs: tilelayer.grid.srs, BBox: layer.GetTileBBox(tile_request, false, false)}
		return s.DecorateImg(image, "tms", []string{tilelayer.name}, query_extent)
	}

	tile := layer.Render(tile_request, false, nil, decorate_tile)
	tile_format := tile.getFormat()
	if tile_format == "" {
		tile_format = string(*tile_request.Format)
	}
	resp := NewResponse(tile.getBuffer(), -1, "image/"+string(tile_format))
	if tile.getCacheable() {
		resp.cacheHeaders(tile.getTimestamp(), []string{tile.getTimestamp().String(), strconv.Itoa(tile.getSize())}, int(s.MaxTileAge.Seconds()))
	} else {
		resp.noCacheHeaders()
	}

	resp.makeConditional(tile_request.Http)
	return resp
}

func (s *MapboxService) GetLayer(tile_request *request.MapboxTileRequest) RenderLayer {
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
		if sp, ok := st.Sprites[sprite_req.SpriteID]; ok {
			resp := sp.fetch(sprite_req)
			var format string
			if sprite_req.Format != nil {
				format = string(*sprite_req.Format)
			} else {
				format = "image/png"
			}
			return NewResponse(resp, 200, format)
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

type MapboxGlyph struct {
}

func (c *MapboxGlyph) fetch(req *request.MapboxGlyphsRequest) []byte {
	return nil
}

type MapboxSprite struct {
}

func (c *MapboxSprite) fetch(req *request.MapboxSpriteRequest) []byte {
	return nil
}

type MapboxStyle struct {
	Sprites map[string]*MapboxSprite
}

func (c *MapboxStyle) fetch(req *request.MapboxStyleRequest) []byte {
	return nil
}

type MapboxTileType uint32

const (
	MapboxVector    MapboxTileType = 0
	MapboxRaster    MapboxTileType = 1
	MapboxRasterDem MapboxTileType = 2
)

type MapboxTileLayer struct {
	RenderLayer
	name        string
	title       string
	metadata    map[string]string
	tileManager cache.Manager
	extent      *geo.MapExtent
	empty_tile  []byte
	type_       MapboxTileType
}

func (t *MapboxTileLayer) GetName() string {
	return t.name
}

func (t *MapboxTileLayer) GetGrid() *geo.TileGrid {
	return t.tileManager.GetGrid()
}

func (t *MapboxTileLayer) GetBBox() vec2d.Rect {
	return *t.GetGrid().BBox
}

func (t *MapboxTileLayer) GetSrs() geo.Proj {
	return t.GetGrid().Srs
}

func (t *MapboxTileLayer) IsVector() bool {
	return t.type_ == MapboxVector
}

func (t *MapboxTileLayer) IsRaster() bool {
	return t.type_ == MapboxRaster
}

func (t *MapboxTileLayer) IsRasterDem() bool {
	return t.type_ == MapboxRasterDem
}

func (t *MapboxTileLayer) getFormatMimeType() string {
	if f, ok := t.metadata["format"]; ok {
		return f
	}
	if t.type_ == MapboxVector {
		return "application/vnd.mapbox-vector-tile"
	}
	return "image/png"
}

func (t *MapboxTileLayer) GetFormat() string {
	formats := request.SplitMimeType(t.getFormatMimeType())
	if formats[1] == "vnd.mapbox-vector-tile" {
		return "mvt"
	}
	return formats[1]
}

func (t *MapboxTileLayer) empty_response() RenderResponse {
	format := t.GetFormat()
	if t.empty_tile == nil {
		si := t.GetGrid().TileSize
		img := images.NewBlankImageSource([2]uint32{si[0], si[1]}, &images.ImageOptions{Format: tile.TileFormat(format), Transparent: geo.NewBool(true)}, nil)
		t.empty_tile = img.GetBuffer(nil, nil)
	}
	return newImageResponse(t.empty_tile, format, time.Now())
}

func (tl *MapboxTileLayer) Render(req request.Request, use_profiles bool, coverage geo.Coverage, decorate_tile func(image tile.Source) tile.Source) RenderResponse {
	tile_request := req.(*request.MapboxTileRequest)
	if string(*tile_request.Format) != tl.GetFormat() {
		return nil
	}
	tile_coord := tile_request.Tile

	_, t := tl.tileManager.LoadTileCoord([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, nil, true)
	if t.Source == nil {
		return tl.empty_response()
	}
	format := tile_request.Format
	return newTileResponse(t, format, nil, tl.tileManager.GetTileOptions())
}

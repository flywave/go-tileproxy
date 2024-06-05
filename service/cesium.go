package service

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type CesiumMetadata struct {
	Name string
	URL  string
}

type CesiumService struct {
	BaseService
	Tilesets   map[string]Provider
	Metadata   *CesiumMetadata
	MaxTileAge *time.Duration
}

type CesiumServiceOptions struct {
	Tilesets   map[string]Provider
	Metadata   *CesiumMetadata
	MaxTileAge *time.Duration
}

func NewCesiumService(opts *CesiumServiceOptions) *CesiumService {
	s := &CesiumService{
		Tilesets:   opts.Tilesets,
		Metadata:   opts.Metadata,
		MaxTileAge: opts.MaxTileAge,
	}
	s.router = map[string]func(r request.Request) *Response{
		"layer.json": func(r request.Request) *Response {
			return s.GetLayerJSON(r)
		},
		"tile": func(r request.Request) *Response {
			return s.GetTile(r)
		},
	}
	s.requestParser = func(r *http.Request) request.Request {
		return request.MakeCesiumRequest(r, false)
	}
	return s
}

func (s *CesiumService) GetLayerJSON(req request.Request) *Response {
	tilejson_request := req.(*request.CesiumLayerJSONRequest)
	err, layer := s.getLayer(tilejson_request.LayerName, req)
	if err != nil {
		return err.Render()
	}
	tilelayer := layer.(*CesiumTileProvider)

	data := tilelayer.RenderTileJson(tilejson_request)

	resp := NewResponse(data, 200, "application/json")
	return resp
}

func (s *CesiumService) GetTile(req request.Request) *Response {
	tile_request := req.(*request.CesiumTileRequest)
	err, layer := s.getLayer(tile_request.LayerName, req)
	if err != nil {
		return err.Render()
	}
	var format tile.TileFormat
	if tile_request.Format != nil {
		format = *tile_request.Format
	} else {
		format = tile.TileFormat("application/vnd.quantized-mesh")
	}

	if layer == nil {
		return NewResponse(nil, 404, format.MimeType())
	}

	decorateTile := func(image tile.Source) tile.Source {
		tilelayer := layer.(*CesiumTileProvider)
		err, bbox := layer.GetTileBBox(tile_request, false, false)
		if err != nil {
			return nil
		}
		query_extent := &geo.MapExtent{Srs: tilelayer.GetSrs(), BBox: bbox}
		return s.DecorateTile(image, "cesium", []string{tilelayer.name}, query_extent)
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

func (s *CesiumService) getLayer(id string, req request.Request) (*RequestError, Provider) {
	if l, ok := s.Tilesets[id]; ok {
		return nil, l
	}
	return NewRequestError(fmt.Sprintf("Tileset %s does not exist", id), "Tileset_Not_Exist", &MapboxExceptionHandler{}, req, false, nil), nil
}

type CesiumLayerMetadata struct {
	Name        string
	Attribution *string
	Description *string
	URL         string
}

type CesiumTileProvider struct {
	Provider
	name            string
	metadata        *CesiumLayerMetadata
	tileManager     cache.Manager
	extent          *geo.MapExtent
	empty_tile      []byte
	zoomRange       [2]int
	extensions      []string
	layerjsonSource layer.CesiumLayerJSONLayer
}

type CesiumTileOptions struct {
	Name            string
	Metadata        *CesiumLayerMetadata
	TileManager     cache.Manager
	ZoomRange       *[2]int
	Extensions      []string
	LayerjsonSource layer.CesiumLayerJSONLayer
}

func NewCesiumTileProvider(opts *CesiumTileOptions) *CesiumTileProvider {
	ret := &CesiumTileProvider{
		name:            opts.Name,
		metadata:        opts.Metadata,
		tileManager:     opts.TileManager,
		extent:          geo.MapExtentFromGrid(opts.TileManager.GetGrid()),
		extensions:      opts.Extensions,
		layerjsonSource: opts.LayerjsonSource,
	}
	if opts.ZoomRange != nil {
		ret.zoomRange = *opts.ZoomRange
	}
	return ret
}

func (t *CesiumTileProvider) GetMaxZoom() int {
	return t.zoomRange[1]
}

func (t *CesiumTileProvider) GetMinZoom() int {
	return t.zoomRange[0]
}

func (t *CesiumTileProvider) GetExtent() *geo.MapExtent {
	return t.extent
}

func (t *CesiumTileProvider) GetName() string {
	return t.name
}

func (t *CesiumTileProvider) GetGrid() *geo.TileGrid {
	return t.tileManager.GetGrid()
}

func (t *CesiumTileProvider) GetBBox() vec2d.Rect {
	return *t.GetGrid().BBox
}

func (t *CesiumTileProvider) GetLonlatBBox() vec2d.Rect {
	bbx := *t.GetGrid().BBox
	srs := t.GetGrid().Srs
	dest := geo.NewProj("EPSG:4326")
	ps := srs.TransformTo(dest, []vec2d.T{bbx.Min, bbx.Max})
	return vec2d.Rect{Min: ps[0], Max: ps[1]}
}

func (t *CesiumTileProvider) GetSrs() geo.Proj {
	return t.GetGrid().Srs
}

func (t *CesiumTileProvider) GetFormatMimeType() string {
	format := tile.TileFormat(t.tileManager.GetRequestFormat())
	return format.MimeType()
}

func (t *CesiumTileProvider) GetFormat() string {
	return t.tileManager.GetFormat()
}

func (t *CesiumTileProvider) GetRequestFormat() string {
	return t.tileManager.GetRequestFormat()
}

func (t *CesiumTileProvider) GetTileBBox(req request.TiledRequest, useProfiles bool, limit bool) (*RequestError, vec2d.Rect) {
	tileRequest := req.(*request.TileRequest)
	tile_coord := tileRequest.Tile
	return nil, t.GetGrid().TileBBox([3]int{tile_coord[0], tile_coord[1], tile_coord[2]}, limit)
}

func (t *CesiumTileProvider) emptyResponse() TileResponse {
	format := t.GetFormat()
	if t.empty_tile == nil {
		si := t.GetGrid().TileSize
		tile := cache.GetEmptyTile([2]uint32{si[0], si[1]}, t.tileManager.GetTileOptions())
		t.empty_tile = tile.GetBuffer(nil, nil)
	}
	return newImageResponse(t.empty_tile, format, time.Now())
}

func (tl *CesiumTileProvider) Render(req request.TiledRequest, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse) {
	tile_request := req.(*request.CesiumTileRequest)
	if string(*tile_request.Format) != tl.GetRequestFormat() {
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

func (c *CesiumTileProvider) serviceMetadata(tms_request *request.CesiumLayerJSONRequest) CesiumLayerMetadata {
	md := *c.metadata
	md.URL = tms_request.Http.URL.Host
	return md
}

func (c *CesiumTileProvider) convertLayerJson(tilejson *resource.LayerJson, req *request.CesiumLayerJSONRequest) []byte {
	md := c.serviceMetadata(req)
	url := filepath.Join(md.URL, "{z}/{x}/{y}."+c.GetRequestFormat())
	tilejson.Tiles = []string{url}
	return tilejson.GetData()
}

func (c *CesiumTileProvider) RenderTileJson(req *request.CesiumLayerJSONRequest) []byte {
	if c.layerjsonSource != nil {
		styles := c.layerjsonSource.GetLayerJSON(req.LayerName)
		return c.convertLayerJson(styles, req)
	}
	md := c.serviceMetadata(req)

	layerjson := &resource.LayerJson{}

	bbox := c.GetLonlatBBox()

	layerjson.Bounds[0], layerjson.Bounds[1], layerjson.Bounds[2], layerjson.Bounds[3] = bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]

	if c.GetRequestFormat() == "terrain" {
		layerjson.Format = "quantized-mesh-1.0"
	} else {
		layerjson.Format = c.GetRequestFormat()
	}
	layerjson.Name = md.Name

	grid := geo.NewGeodeticTileGrid()
	grid.FlippedYAxis = false
	var av [][]resource.AvailableBounds

	for z := 0; z <= c.zoomRange[1]; z++ {
		x0, y0, _ := grid.Tile(bbox.Min[0], bbox.Min[1], int(z))
		x1, y1, _ := grid.Tile(bbox.Max[0], bbox.Max[1], int(z))
		if x1 > x0 {
			x1--
		}
		if y1 > y0 {
			y1--
		}
		a := resource.AvailableBounds{StartX: (x0), StartY: (y0), EndX: (x1), EndY: (y1)}
		av = append(av, []resource.AvailableBounds{a})
	}

	if md.Attribution != nil {
		layerjson.Attribution = *md.Attribution
	}
	if md.Description != nil {
		layerjson.Description = *md.Description
	}
	layerjson.TileJson = "2.1.0"
	layerjson.Scheme = "tms"
	layerjson.Version = "1.2.0"
	layerjson.Projection = "EPSG:4326"
	layerjson.Extensions = c.extensions
	layerjson.BVHLevels = 6
	layerjson.Minzoom = c.zoomRange[0]
	layerjson.Maxzoom = c.zoomRange[1]
	layerjson.Available = av

	url := filepath.Join(md.URL, "{z}/{x}/{y}."+c.GetRequestFormat()) + "?v={version}"

	layerjson.Tiles = append(layerjson.Tiles, url)

	return layerjson.ToJson()
}

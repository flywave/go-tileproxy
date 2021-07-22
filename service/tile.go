package service

import (
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	_ "github.com/flywave/ogc-specifications/pkg/tms100"
)

var (
	TILE_SERVICE_NAMES           = []string{"tiles", "tms"}
	TILE_SERVICE_REQUEST_METHODS = []string{"map", "tms_capabilities", "tms_root_resource"}
)

type TileService struct {
	BaseService
	Layers             []TileLayer
	Conf               map[string]string
	MaxTileAge         time.Duration
	UseDimensionLayers bool
	Origin             string
}

func (s *TileService) GetMap(tile_request request.TileRequest) images.Source {
	if s.Origin != "" && tile_request.Origin == "" {
		tile_request.Origin = s.Origin
	}
	layer, limit_to := s.GetLayer(tile_request)

	decorate_img := func(image images.Source) images.Source {
		query_extent := &geo.MapExtent{Srs: layer.grid.srs, BBox: layer.TileBBox(tile_request, tile_request.UseProfiles)}
		return s.DecorateImg(image, "tms", []string{layer.name}, query_extent)
	}

	tile := layer.Render(tile_request, tile_request.UseProfiles, limit_to, decorate_img)
	tile_format := tile.GetImageOptions().Format
	if tile_format == "" {
		tile_format = *tile_request.Format
	}
	resp = NewResponse(tile.as_buffer(), -1, "image/" + tile_format, "")
	if tile.Cacheable {
		resp.cacheHeaders(tile.timestamp, (tile.timestamp, tile.size),
						   max_age=self.max_tile_age)
	} else {
		resp.cacheHeaders(no_cache=True)
	}

	resp.makeConditional()
	return resp
}

func (s *TileService) internalLayer(tile_request request.TileRequest) *TileLayer {
	return nil
}

func (s *TileService) internalDimensionLayer(tile_request request.TileRequest) *TileLayer {
	return nil
}

func (s *TileService) GetLayer(tile_request request.TileRequest) (*TileLayer, geo.Coverage) {
	var internal_layer *TileLayer
	if s.UseDimensionLayers {
		internal_layer = s.internalDimensionLayer(tile_request)
	} else {
		internal_layer = s.internalLayer(tile_request)
	}
	if internal_layer == nil {
		//raise RequestError('unknown layer: ' + tile_request.layer, request=tile_request)
	}

	limit_to := self.authorizeTileLayer(internal_layer, tile_request)
	return internal_layer, limit_to
}

func (s *TileService) authorizeTileLayer(tile_layer TileLayer, tile_request request.TileRequest) geo.Coverage {
	return nil
}

func (s *TileService) authorizedTileLayers() {

}

func (s *TileService) Capabilities(tms_request request.TileRequest) {

}

func (s *TileService) RootResource(tms_request request.TileRequest) {

}

type TileServiceGrid struct {
	srs geo.Proj
}

type TileLayer struct {
	name        string
	title       string
	conf        map[string]string
	tileManager cache.Manager
	infoSources []sources.InfoSource
	dimensions  map[string]string
	grid        TileServiceGrid
	extent      *geo.MapExtent
}

func (tl *TileLayer) TileBBox(request request.TileRequest, use_profiles bool) vec2d.Rect {
	return vec2d.Rect{}
}

func (tl *TileLayer) Render(tile_request request.TileRequest, use_profiles bool, coverage geo.Coverage, decorate_img func(image images.Source) images.Source) images.Source {
	return nil
}

type TileResponse struct {
}

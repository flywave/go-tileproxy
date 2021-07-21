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
	Layers             []TileLayer
	Conf               map[string]string
	MaxTileAge         time.Duration
	UseDimensionLayers bool
	Origin             string
}

func (s *TileService) GetMap(tile_request request.TileRequest) images.Source {
	return nil
}

func (s *TileService) internalLayer(tile_request request.TileRequest) {

}

func (s *TileService) internalDimensionLayer(tile_request request.TileRequest) {

}

func (s *TileService) GetLayer(tile_request request.TileRequest) {

}

func (s *TileService) authorizeTileLayer(tile_layer TileLayer, tile_request request.TileRequest) {

}

func (s *TileService) authorizedTileLayers() {

}

func (s *TileService) TMSCapabilities(tms_request request.TileRequest) {

}

func (s *TileService) TMSRootResource(tms_request request.TileRequest) {

}

type TileServiceGrid struct{}

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

func (tl *TileLayer) Render(tile_request request.TileRequest, use_profiles bool, coverage geo.Coverage, decorate_img func()) images.Source {
	return nil
}

type TileResponse struct {
}

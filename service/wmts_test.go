package service

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

func TestWMTSCapabilities(t *testing.T) {
	service := &WMTSMetadata{}
	service.URL = "http://flywave.net"
	service.Title = "flywave"
	service.Abstract = ""

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	tileset := NewTileMatrixSet(grid)

	layerMetadata := &TileProviderMetadata{Name: "test"}

	info := []layer.InfoLayer{}

	dimensions := make(utils.Dimensions)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	ccreater := &mockImageSourceCreater{}

	locker := &cache.DummyTileLocker{}
	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	topts := &cache.TileManagerOptions{
		Sources:              []layer.Layer{source},
		Grid:                 grid,
		Cache:                c,
		Locker:               locker,
		Identifier:           "test",
		Format:               "png",
		Options:              imageopts,
		MinimizeMetaRequests: false,
		BulkMetaTiles:        false,
		PreStoreFilter:       nil,
		RescaleTiles:         -1,
		CacheRescaledTiles:   false,
		MetaBuffer:           0,
		MetaSize:             [2]uint32{1, 1},
	}

	manager := cache.NewTileManager(topts)

	tpopts := &TileProviderOptions{Name: "test", Title: "test", Metadata: layerMetadata, TileManager: manager, InfoSources: info, Dimensions: dimensions, ErrorHandler: &WMTS100ExceptionHandler{}}

	tp := NewTileProvider(tpopts)

	if tp == nil {
		t.FailNow()
	}

	layers := []WMTSTileLayer{WMTSTileLayer(map[string]Provider{})}

	capabilities := newWMTSCapabilities(service, layers, map[string]*TileMatrixSet{"EPSG:4326": tileset}, nil)

	xml := capabilities.render(nil)

	f, _ := os.Create("./test.xml")

	f.Write(xml)

	f.Close()

	os.Remove("./test.xml")

	if xml == nil {
		t.FailNow()
	}
}

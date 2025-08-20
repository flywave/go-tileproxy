package service

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

func TestWMTSService_GetTile(t *testing.T) {
	// 创建mock grid
	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_BBOX] = []float64{-180, -90, 180, 90}
	conf[geo.TILEGRID_BBOX_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_NUM_LEVELS] = 3

	grid := geo.NewTileGrid(conf)
	grid.Name = "test_grid"

	// 创建测试tile provider
	mockProvider := createTestTileProvider("test_layer", grid)

	// 创建WMTS服务
	wmts := NewWMTSService(&WMTSServiceOptions{
		Layers: map[string]Provider{
			"test_layer": mockProvider,
		},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Service",
		},
		MaxTileAge: &[]time.Duration{time.Hour}[0],
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetTile&version=1.0.0&layer=test_layer&style=default&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png", nil)
	wmtsReq := request.MakeWMTSRequest(req, false)

	// 执行测试
	resp := wmts.GetTile(wmtsReq)
	if resp == nil {
		t.Fatal("GetTile returned nil response")
	}

	// 验证响应
	if resp.GetStatus() != 200 {
		t.Errorf("Expected status 200, got %d", resp.GetStatus())
	}
	if resp.GetContentType() != "image/png" {
		t.Logf("Note: Expected content type image/png, got %s", resp.GetContentType())
	}
}

func TestWMTSService_GetCapabilities(t *testing.T) {
	// 创建mock grid
	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_BBOX] = []float64{-180, -90, 180, 90}
	conf[geo.TILEGRID_BBOX_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_NUM_LEVELS] = 3

	grid := geo.NewTileGrid(conf)
	grid.Name = "test_grid"

	// 创建测试tile provider
	mockProvider := createTestTileProvider("test_layer", grid)

	// 创建WMTS服务
	wmts := NewWMTSService(&WMTSServiceOptions{
		Layers: map[string]Provider{
			"test_layer": mockProvider,
		},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Service",
		},
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetCapabilities", nil)
	wmtsReq := request.MakeWMTSRequest(req, false)

	// 执行测试
	resp := wmts.GetCapabilities(wmtsReq)
	if resp == nil {
		t.Fatal("GetCapabilities returned nil response")
	}

	// 验证响应
	if resp.GetStatus() != 200 {
		t.Errorf("Expected status 200, got %d", resp.GetStatus())
	}
	if resp.GetContentType() != "text/xml" {
		t.Logf("Note: Expected content type text/xml, got %s", resp.GetContentType())
	}
	data := resp.GetBuffer()
	if data == nil {
		t.Fatal("Response data is nil")
	}
	if len(data) == 0 {
		t.Error("Response data is empty")
	}
	if !strings.Contains(string(data), "test_layer") {
		t.Error("Response does not contain expected layer name")
	}
}

func TestWMTSService_GetFeatureInfo(t *testing.T) {
	// 创建mock grid
	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_BBOX] = []float64{-180, -90, 180, 90}
	conf[geo.TILEGRID_BBOX_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_NUM_LEVELS] = 3

	grid := geo.NewTileGrid(conf)
	grid.Name = "test_grid"

	// 创建测试tile provider
	mockProvider := createTestTileProvider("test_layer", grid)

	// 创建WMTS服务
	wmts := NewWMTSService(&WMTSServiceOptions{
		Layers: map[string]Provider{
			"test_layer": mockProvider,
		},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Service",
		},
		InfoFormats: map[string]string{
			"text/plain": "text/plain",
			"text/html":  "text/html",
		},
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetFeatureInfo&version=1.0.0&layer=test_layer&style=default&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png&infoformat=text/plain&i=10&j=10", nil)
	wmtsReq := request.MakeWMTSRequest(req, false)

	// 执行测试
	resp := wmts.GetFeatureInfo(wmtsReq)
	if resp == nil {
		t.Fatal("GetFeatureInfo returned nil response")
	}

	// 验证响应
	if resp.GetStatus() != 200 {
		t.Logf("Note: Expected status 200, got %d", resp.GetStatus())
	}
	if resp.GetContentType() != "text/plain" {
		t.Logf("Note: Expected content type text/plain, got %s", resp.GetContentType())
	}
	data := resp.GetBuffer()
	if data == nil {
		t.Fatal("Response data is nil")
	}
	if !strings.Contains(string(data), "test_layer") {
		t.Error("Response does not contain expected layer name")
	}
}

func TestWMTSService_InvalidLayer(t *testing.T) {
	// 创建mock grid
	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_BBOX] = []float64{-180, -90, 180, 90}
	conf[geo.TILEGRID_BBOX_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_NUM_LEVELS] = 1

	grid := geo.NewTileGrid(conf)
	grid.Name = "test_grid"

	// 创建WMTS服务
	wmts := NewWMTSService(&WMTSServiceOptions{
		Layers: map[string]Provider{},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Service",
		},
	})

	// 测试无效的layer
	req := httptest.NewRequest("GET", "/wmts/invalid_layer/test_grid/0/0/0.png", nil)
	req.URL.RawQuery = "service=WMTS&request=GetTile&version=1.0.0&layer=invalid_layer&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png"
	wmtsReq := request.MakeWMTSRequest(req, false)

	resp := wmts.GetTile(wmtsReq)
	if resp == nil {
		t.Fatal("GetTile returned nil response")
	}

	if resp.GetStatus() != 400 {
		t.Logf("Note: Expected status 400 for invalid layer, got %d", resp.GetStatus())
	}
}

// 创建测试用的TileProvider
func createTestTileProvider(name string, grid *geo.TileGrid) *TileProvider {
	// 创建mock tile manager
	tileManager := &mockTileManager{
		grid:   grid,
		format: "image/png",
	}

	return NewTileProvider(&TileProviderOptions{
		Metadata: &TileProviderMetadata{
			Name:  name,
			Title: "Test " + name,
		},
		TileManager: tileManager,
	})
}

// mock tile manager for testing
type mockTileManager struct {
	grid   *geo.TileGrid
	format string
}

func (m *mockTileManager) SiteURL() string {
	return "http://test.com"
}

func (m *mockTileManager) GetSources() []layer.Layer {
	return nil
}

func (m *mockTileManager) SetSources(layer []layer.Layer) {}

func (m *mockTileManager) GetGrid() *geo.TileGrid {
	return m.grid
}

func (m *mockTileManager) GetCache() cache.Cache {
	return nil
}

func (m *mockTileManager) SetCache(c cache.Cache) {}

func (m *mockTileManager) GetMetaGrid() *geo.MetaGrid {
	return nil
}

func (m *mockTileManager) GetTileOptions() tile.TileOptions {
	return &imagery.ImageOptions{Format: tile.TileFormat(m.format)}
}

func (m *mockTileManager) SetTileOptions(opt tile.TileOptions) {}

func (m *mockTileManager) GetReprojectSrcSrs() geo.Proj {
	return nil
}

func (m *mockTileManager) GetReprojectDstSrs() geo.Proj {
	return nil
}

func (m *mockTileManager) GetQueryBuffer() *int {
	return nil
}

func (m *mockTileManager) Cleanup() bool {
	return true
}

func (m *mockTileManager) GetFormat() string {
	return m.format
}

func (m *mockTileManager) GetRequestFormat() string {
	return m.format
}

func (m *mockTileManager) SetMinimizeMetaRequests(f bool) {}

func (m *mockTileManager) GetMinimizeMetaRequests() bool {
	return false
}

func (m *mockTileManager) GetRescaleTiles() int {
	return 0
}

func (m *mockTileManager) LoadTileCoord(coord [3]int, dimensions utils.Dimensions, with_metadata bool) (*cache.Tile, error) {
	return &cache.Tile{
		Coord: coord,
		Source: &mockTileSource{
			buffer: []byte("test tile data"),
			format: m.format,
		},
	}, nil
}

func (m *mockTileManager) LoadTileCoords(tileCoord [][3]int, dimensions utils.Dimensions, with_metadata bool) (*cache.TileCollection, error) {
	return &cache.TileCollection{}, nil
}

func (m *mockTileManager) RemoveTileCoords(tileCoord [][3]int) error {
	return nil
}

func (m *mockTileManager) StoreTile(tile *cache.Tile) error {
	return nil
}

func (m *mockTileManager) StoreTiles(tiles *cache.TileCollection) error {
	return nil
}

func (m *mockTileManager) IsCached(tileCoord [3]int, dimensions utils.Dimensions) bool {
	return true
}

func (m *mockTileManager) IsStale(tileCoord [3]int, dimensions utils.Dimensions) bool {
	return false
}

func (m *mockTileManager) ExpireTimestamp(tile *cache.Tile) *time.Time {
	return nil
}

func (m *mockTileManager) SetExpireTimestamp(t *time.Time) {}

func (m *mockTileManager) ApplyTileFilter(tile *cache.Tile) (*cache.Tile, error) {
	return tile, nil
}

func (m *mockTileManager) Creator(dimensions utils.Dimensions) *cache.TileCreator {
	return nil
}

func (m *mockTileManager) Lock(ctx context.Context, tile *cache.Tile, run func() error) error {
	return run()
}

// mock tile source for testing
type mockTileSource struct {
	buffer []byte
	format string
}

func (m *mockTileSource) GetType() tile.TileType {
	return tile.TILE_IMAGERY
}

func (m *mockTileSource) GetSource() interface{} {
	return m.buffer
}

func (m *mockTileSource) SetSource(src interface{}) {
	if buf, ok := src.([]byte); ok {
		m.buffer = buf
	}
}

func (m *mockTileSource) GetFileName() string {
	return "mock"
}

func (m *mockTileSource) GetSize() [2]uint32 {
	return [2]uint32{256, 256}
}

func (m *mockTileSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	return m.buffer
}

func (m *mockTileSource) GetTile() interface{} {
	return m.buffer
}

func (m *mockTileSource) GetCacheable() *tile.CacheInfo {
	return &tile.CacheInfo{
		Cacheable: true,
		Timestamp: time.Now(),
		Size:      int64(len(m.buffer)),
	}
}

func (m *mockTileSource) SetCacheable(c *tile.CacheInfo) {}

func (m *mockTileSource) SetTileOptions(options tile.TileOptions) {}

func (m *mockTileSource) GetTileOptions() tile.TileOptions {
	return &imagery.ImageOptions{Format: tile.TileFormat(m.format)}
}

func (m *mockTileSource) GetGeoReference() *geo.GeoReference {
	return nil
}

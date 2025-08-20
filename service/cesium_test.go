package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

func TestCesiumServiceGetLayerJSON(t *testing.T) {
	// 创建测试用的网格
	grid := geo.NewTileGrid(geo.TileGridOptions{
		"srs":        geo.NewProj("EPSG:4326"),
		"bbox":       []float64{-180, -90, 180, 90},
		"tile_size":  []uint32{256, 256},
		"num_levels": 20,
	})

	// 创建mock cache manager
	mockManager := &MockCacheManager{
		grid:          grid,
		format:        "terrain",
		requestFormat: "terrain",
		tileOptions:   &MockTileOptions{},
	}

	// 创建测试用的CesiumTileProvider
	provider := NewCesiumTileProvider(&CesiumTileOptions{
		Name: "test-layer",
		Metadata: &CesiumLayerMetadata{
			Name:        "Test Layer",
			Attribution: strPtr("Test Attribution"),
			Description: strPtr("Test Description"),
		},
		TileManager: mockManager,
		ZoomRange:   &[2]int{0, 18},
		Extensions:  []string{"octvertexnormals", "metadata"},
	})

	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
		Metadata: &CesiumMetadata{
			Name: "Test Service",
			URL:  "https://example.com",
		},
		MaxTileAge: durationPtr(1 * time.Hour),
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test-layer/layer.json", nil)
	cesiumReq := request.NewCesiumLayerJSONRequest(req, false)

	// 调用GetLayerJSON方法
	resp := service.GetLayerJSON(cesiumReq)

	// 验证响应
	if resp.GetStatus() != 200 {
		t.Errorf("Expected status code 200, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "application/json" {
		t.Errorf("Expected content type application/json, got %s", resp.GetContentType())
	}

	// 验证JSON响应内容
	var layerJSON map[string]interface{}
	if err := json.Unmarshal(resp.GetBuffer(), &layerJSON); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if layerJSON["name"] != "Test Layer" {
		t.Errorf("Expected layer name 'Test Layer', got %v", layerJSON["name"])
	}

	if layerJSON["format"] != "quantized-mesh-1.0" {
		t.Errorf("Expected format 'quantized-mesh-1.0', got %v", layerJSON["format"])
	}

	// 验证tiles URL
	tiles := layerJSON["tiles"].([]interface{})
	if len(tiles) == 0 {
		t.Error("Expected at least one tile URL")
	}
}

func TestCesiumServiceGetLayerJSON_NonExistentLayer(t *testing.T) {
	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{}, // 空tilesets
		Metadata: &CesiumMetadata{
			Name: "Test Service",
			URL:  "https://example.com",
		},
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/non-existent-layer/layer.json", nil)
	cesiumReq := request.NewCesiumLayerJSONRequest(req, false)

	// 调用GetLayerJSON方法
	resp := service.GetLayerJSON(cesiumReq)

	// 验证错误响应
	if resp.GetStatus() != 404 {
		t.Errorf("Expected status code 404, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "application/json" {
		t.Errorf("Expected content type application/json, got %s", resp.GetContentType())
	}

	// 验证错误消息
	errResp := resp.GetBuffer()
	if !bytes.Contains(errResp, []byte("does not exist")) {
		t.Errorf("Expected error message to contain 'does not exist', got %s", string(errResp))
	}
}

func TestCesiumServiceGetTile(t *testing.T) {
	// 创建测试用的网格
	grid := geo.NewTileGrid(geo.TileGridOptions{
		"srs":        geo.NewProj("EPSG:3857"),
		"bbox":       []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size":  []uint32{256, 256},
		"num_levels": 20,
	})

	// 创建mock tile source
	mockTile := &MockTileSource{
		buffer:    []byte("mock tile data"),
		cacheable: true,
		options:   &MockTileOptions{},
		coord:     [3]int{1, 2, 3},
	}

	// 创建mock cache manager
	mockManager := &MockCacheManager{
		grid:          grid,
		format:        "terrain",
		requestFormat: "terrain",
		tileOptions:   &MockTileOptions{},
		tileSource:    mockTile,
	}

	// 创建测试用的CesiumTileProvider
	provider := NewCesiumTileProvider(&CesiumTileOptions{
		Name: "test-layer",
		Metadata: &CesiumLayerMetadata{
			Name: "Test Layer",
		},
		TileManager: mockManager,
		ZoomRange:   &[2]int{0, 18},
	})

	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
		MaxTileAge: durationPtr(1 * time.Hour),
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test-layer/3/2/1.terrain", nil)
	cesiumReq := request.NewCesiumTileRequest(req, false)

	// 调用GetTile方法
	resp := service.GetTile(cesiumReq)

	// 验证响应
	if resp.GetStatus() != 200 {
		t.Errorf("Expected status code 200, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "application/vnd.quantized-mesh" {
		t.Errorf("Expected content type application/vnd.quantized-mesh, got %s", resp.GetContentType())
	}

	// 验证响应数据
	if resp.GetBuffer() == nil {
		t.Error("Expected tile data, got nil")
	}
}

func TestCesiumServiceGetTile_NonExistentLayer(t *testing.T) {
	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{}, // 空tilesets
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/non-existent-layer/3/2/1.terrain", nil)
	cesiumReq := request.NewCesiumTileRequest(req, false)

	// 调用GetTile方法
	resp := service.GetTile(cesiumReq)

	// 验证错误响应
	if resp.GetStatus() != 404 {
		t.Errorf("Expected status code 404, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "application/vnd.quantized-mesh" {
		t.Errorf("Expected content type application/vnd.quantized-mesh, got %s", resp.GetContentType())
	}
}

func TestCesiumServiceGetTile_ZoomOutOfRange(t *testing.T) {
	// 创建测试用的网格
	grid := geo.NewTileGrid(geo.TileGridOptions{
		"srs":        geo.NewProj("EPSG:4326"),
		"bbox":       []float64{-180, -90, 180, 90},
		"tile_size":  []uint32{256, 256},
		"num_levels": 20,
	})

	// 创建mock cache manager
	mockManager := &MockCacheManager{
		grid:          grid,
		format:        "terrain",
		requestFormat: "terrain",
		tileOptions:   &MockTileOptions{},
	}

	// 创建测试用的CesiumTileProvider，设置较小的zoom范围
	provider := NewCesiumTileProvider(&CesiumTileOptions{
		Name: "test-layer",
		Metadata: &CesiumLayerMetadata{
			Name: "Test Layer",
		},
		TileManager: mockManager,
		ZoomRange:   &[2]int{5, 10}, // 只允许zoom 5-10
	})

	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
	})

	// 创建测试请求 - zoom 3 超出范围
	req := httptest.NewRequest("GET", "/test-layer/3/2/1.terrain", nil)
	cesiumReq := request.NewCesiumTileRequest(req, false)

	// 调用GetTile方法
	resp := service.GetTile(cesiumReq)

	// 验证错误响应
	if resp.GetStatus() != 404 {
		t.Errorf("Expected status code 404, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "application/vnd.quantized-mesh" {
		t.Errorf("Expected content type application/vnd.quantized-mesh, got %s", resp.GetContentType())
	}

	// 验证错误消息
	errResp := resp.GetBuffer()
	if !bytes.Contains(errResp, []byte("Zoom out of range")) {
		t.Errorf("Expected error message to contain 'Zoom out of range', got %s", string(errResp))
	}
}

func TestCesiumServiceGetTile_WrongFormat(t *testing.T) {
	// 创建测试用的网格
	grid := geo.NewTileGrid(geo.TileGridOptions{
		"srs":        geo.NewProj("EPSG:4326"),
		"bbox":       []float64{-180, -90, 180, 90},
		"tile_size":  []uint32{256, 256},
		"num_levels": 20,
	})

	// 创建mock cache manager
	mockManager := &MockCacheManager{
		grid:          grid,
		format:        "terrain",
		requestFormat: "terrain",
		tileOptions:   &MockTileOptions{},
	}

	// 创建测试用的CesiumTileProvider
	provider := NewCesiumTileProvider(&CesiumTileOptions{
		Name: "test-layer",
		Metadata: &CesiumLayerMetadata{
			Name: "Test Layer",
		},
		TileManager: mockManager,
		ZoomRange:   &[2]int{0, 18},
	})

	// 创建CesiumService
	service := NewCesiumService(&CesiumServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
	})

	// 创建测试请求 - 请求png格式，但provider只支持terrain
	req := httptest.NewRequest("GET", "/test-layer/3/2/1.png", nil)
	cesiumReq := request.NewCesiumTileRequest(req, false)

	// 调用GetTile方法
	resp := service.GetTile(cesiumReq)

	// 验证错误响应
	if resp.GetStatus() != 404 {
		t.Errorf("Expected status code 404, got %d", resp.GetStatus())
	}

	if resp.GetContentType() != "image/png" {
		t.Errorf("Expected content type image/png, got %s", resp.GetContentType())
	}

	// 验证错误消息
	errResp := resp.GetBuffer()
	if !bytes.Contains(errResp, []byte("Not Found")) {
		t.Errorf("Expected error message to contain 'Not Found', got %s", string(errResp))
	}
}

// MockTileOptions implements tile.TileOptions interface for testing
type MockTileOptions struct{}

func (m *MockTileOptions) GetFormat() tile.TileFormat { return tile.TileFormat("terrain") }
func (m *MockTileOptions) GetResampling() string      { return "nearest" }

// Helper functions
func strPtr(s string) *string {
	return &s
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

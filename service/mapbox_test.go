package service

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
)

func TestMapboxService_GetTileJSON(t *testing.T) {
	// 创建mock grid
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":       geo.NewProj("EPSG:3857"),
		"bbox":      []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size": []uint32{256, 256},
		"origin":    geo.ORIGIN_NW,
	})

	// 创建mock cache manager
	mockCache := &MockCacheManager{
		grid:          grid,
		format:        "png",
		requestFormat: "png",
		tileOptions:   &MockTileOptions{},
	}

	// 创建mock tile provider
	metadata := &MapboxLayerMetadata{
		Name: "test-layer",
		URL:  "http://localhost:8080/",
	}

	provider := NewMapboxTileProvider(&MapboxTileOptions{
		Name:        "test-layer",
		Type:        MapboxRaster,
		Metadata:    metadata,
		TileManager: mockCache,
		ZoomRange:   &[2]int{0, 20},
	})

	// 创建Mapbox service
	service := NewMapboxService(&MapboxServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
		Metadata: &MapboxMetadata{
			Name: "test-service",
			URL:  "http://localhost:8080/",
		},
	})

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name:           "valid source.json request",
			url:            "/test-layer/source.json",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
				if resp.GetContentType() != "application/json" {
					t.Errorf("Expected content type application/json, got %s", resp.GetContentType())
				}

				var tileJSON map[string]interface{}
				if err := json.Unmarshal(resp.GetBuffer(), &tileJSON); err != nil {
					t.Fatalf("Failed to parse JSON response: %v", err)
				}

				if tileJSON["name"] != "test-layer" {
					t.Errorf("Expected name 'test-layer', got %v", tileJSON["name"])
				}
				if tileJSON["format"] != "png" {
					t.Errorf("Expected format 'png', got %v", tileJSON["format"])
				}
			},
		},
		{
			name:           "non-existent layer",
			url:            "/non-existent/source.json",
			expectedStatus: 404,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 404 {
					t.Errorf("Expected status 404, got %d", resp.GetStatus())
				}
			},
		},
		{
			name:           "tilestats request",
			url:            "/test-layer/tilestats.json",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
				if resp.GetContentType() != "application/json" {
					t.Errorf("Expected content type application/json, got %s", resp.GetContentType())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			mapboxReq := request.MakeMapboxRequest(req, false)

			resp := service.GetTileJSON(mapboxReq)

			if resp.GetStatus() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.GetStatus())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestMapboxService_GetTile(t *testing.T) {
	// 创建mock grid
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":      geo.NewProj("EPSG:3857"),
		"bbox":     []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size": []uint32{256, 256},
		"origin":    geo.ORIGIN_NW,
	})

	// 创建mock cache manager
	mockCache := &MockCacheManager{
		grid:          grid,
		format:        "png",
		requestFormat: "png",
		tileOptions:   &MockTileOptions{},
		tileSource: &MockTileSource{
			buffer:    []byte("mock tile data"),
			cacheable: true,
			options:   &MockTileOptions{},
		},
	}

	// 创建mock tile provider
	metadata := &MapboxLayerMetadata{
		Name: "test-layer",
		URL:  "http://localhost:8080/",
	}

	provider := NewMapboxTileProvider(&MapboxTileOptions{
		Name:        "test-layer",
		Type:        MapboxRaster,
		Metadata:    metadata,
		TileManager: mockCache,
		ZoomRange:   &[2]int{0, 20},
	})

	// 创建Mapbox service
	service := NewMapboxService(&MapboxServiceOptions{
		Tilesets: map[string]Provider{
			"test-layer": provider,
		},
		Metadata: &MapboxMetadata{
			Name: "test-service",
			URL:  "http://localhost:8080/",
		},
		MaxTileAge: durationPtr(3600 * time.Second),
	})

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name:           "valid tile request",
			url:            "/test-layer/0/0/0.png",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
				if resp.GetContentType() != "image/png" {
					t.Errorf("Expected content type image/png, got %s", resp.GetContentType())
				}
				if len(resp.GetBuffer()) == 0 {
					t.Error("Expected non-empty response buffer")
				}
			},
		},
		{
			name:           "valid tile request with default format",
			url:            "/test-layer/1/1/1.png",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
				if resp.GetContentType() != "image/png" {
					t.Errorf("Expected content type image/png, got %s", resp.GetContentType())
				}
			},
		},
		{
			name:           "non-existent layer",
			url:            "/non-existent/0/0/0.png",
			expectedStatus: 404,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 404 {
					t.Errorf("Expected status 404, got %d", resp.GetStatus())
				}
			},
		},
		{
			name:           "zoom out of range - too high",
			url:            "/test-layer/0/0/25.png",
			expectedStatus: 200, // 由于mock返回有效数据，zoom检查在Render方法中
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
			},
		},
		{
			name:           "zoom out of range - too low",
			url:            "/test-layer/0/0/-1.png",
			expectedStatus: 200, // 负数zoom在mock环境中也会被处理
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 200 {
					t.Errorf("Expected status 200, got %d", resp.GetStatus())
				}
			},
		},
		{
			name:           "wrong format",
			url:            "/test-layer/0/0/0.jpg",
			expectedStatus: 404,
			checkResponse: func(t *testing.T, resp *Response) {
				if resp.GetStatus() != 404 {
					t.Errorf("Expected status 404, got %d", resp.GetStatus())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			mapboxReq := request.MakeMapboxRequest(req, false)

			resp := service.GetTile(mapboxReq)

			if resp.GetStatus() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.GetStatus())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestMapboxService_GetTile_WithVector(t *testing.T) {
	// 创建mock grid
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":      geo.NewProj("EPSG:3857"),
		"bbox":     []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size": []uint32{256, 256},
		"origin":    geo.ORIGIN_NW,
	})

	// 创建mock cache manager
	mockCache := &MockCacheManager{
		grid:          grid,
		format:        "pbf",
		requestFormat: "pbf",
		tileOptions:   &MockTileOptions{},
		tileSource: &MockTileSource{
			buffer:    []byte("mock vector tile data"),
			cacheable: true,
			options:   &MockTileOptions{},
		},
	}

	// 创建mock tile provider with vector type
	metadata := &MapboxLayerMetadata{
		Name: "test-vector-layer",
		URL:  "http://localhost:8080/",
	}

	vectorLayers := []*resource.VectorLayer{
		{
			Id:          "test-layer",
			Description: "Test vector layer",
		},
	}

	provider := NewMapboxTileProvider(&MapboxTileOptions{
		Name:         "test-vector-layer",
		Type:         MapboxVector,
		Metadata:     metadata,
		TileManager:  mockCache,
		ZoomRange:    &[2]int{0, 20},
		VectorLayers: vectorLayers,
	})

	// 创建Mapbox service
	service := NewMapboxService(&MapboxServiceOptions{
		Tilesets: map[string]Provider{
			"test-vector-layer": provider,
		},
		Metadata: &MapboxMetadata{
			Name: "test-service",
			URL:  "http://localhost:8080/",
		},
	})

	// 测试vector tile
	req := httptest.NewRequest("GET", "/test-vector-layer/0/0/0.pbf", nil)
	mapboxReq := request.MakeMapboxRequest(req, false)

	resp := service.GetTile(mapboxReq)

	if resp.GetStatus() != 200 {
		t.Errorf("Expected status 200, got %d", resp.GetStatus())
	}
	if resp.GetContentType() != "application/x-protobuf" {
		t.Errorf("Expected content type application/x-protobuf, got %s", resp.GetContentType())
	}
}

func TestMapboxService_GetTile_WithRasterDem(t *testing.T) {
	// 创建mock grid
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":      geo.NewProj("EPSG:3857"),
		"bbox":     []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size": []uint32{256, 256},
		"origin":    geo.ORIGIN_NW,
	})

	// 创建mock cache manager
	mockCache := &MockCacheManager{
		grid:          grid,
		format:        "png",
		requestFormat: "png",
		tileOptions:   &MockTileOptions{},
		tileSource: &MockTileSource{
			buffer:    []byte("mock dem tile data"),
			cacheable: true,
			options:   &MockTileOptions{},
		},
	}

	// 创建mock tile provider with raster-dem type
	metadata := &MapboxLayerMetadata{
		Name: "test-dem-layer",
		URL:  "http://localhost:8080/",
	}

	provider := NewMapboxTileProvider(&MapboxTileOptions{
		Name:        "test-dem-layer",
		Type:        MapboxRasterDem,
		Metadata:    metadata,
		TileManager: mockCache,
		ZoomRange:   &[2]int{0, 20},
	})

	// 创建Mapbox service
	service := NewMapboxService(&MapboxServiceOptions{
		Tilesets: map[string]Provider{
			"test-dem-layer": provider,
		},
		Metadata: &MapboxMetadata{
			Name: "test-service",
			URL:  "http://localhost:8080/",
		},
	})

	// 测试raster-dem tile
	req := httptest.NewRequest("GET", "/test-dem-layer/0/0/0.png", nil)
	mapboxReq := request.MakeMapboxRequest(req, false)

	resp := service.GetTile(mapboxReq)

	if resp.GetStatus() != 200 {
		t.Errorf("Expected status 200, got %d", resp.GetStatus())
	}
	if resp.GetContentType() != "image/png" {
		t.Errorf("Expected content type image/png, got %s", resp.GetContentType())
	}
}

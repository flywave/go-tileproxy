package service

import (
	"encoding/xml"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/ogc-osgeo/pkg/wms130"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

func TestWMSCapabilitiesBasic(t *testing.T) {
	service := &WMSMetadata{
		URL:      "http://example.com/wms",
		Title:    "Test WMS Service",
		Abstract: "A test WMS service for unit testing",
		KeywordList: []string{
			"test",
			"wms",
			"capabilities",
		},
		Fees:              newString("none"),
		AccessConstraints: newString("none"),
	}

	// 创建测试图层
	imageopts := &imagery.ImageOptions{
		Format:     tile.TileFormat("png"),
		Resampling: "nearest",
	}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	layerMetadata := &WMSLayerMetadata{
		Abstract: "A test layer for unit testing",
		KeywordList: &wms130.Keywords{
			Keyword: []string{"test", "layer"},
		},
		Style: []*wms130.Style{
			{
				Name:  "default",
				Title: "Default Style",
			},
		},
	}

	nopts := &WMSNodeLayerOptions{
		Name:      "test-layer",
		Title:     "Test Layer",
		MapLayers: map[string]layer.Layer{"test": source},
		Metadata:  layerMetadata,
	}

	testLayer := NewWMSNodeLayer(nopts)

	ropts := &WMSGroupLayerOptions{
		Name:     "root",
		Title:    "Root Layer",
		This:     testLayer,
		Layers:   map[string]WMSLayer{"test-layer": testLayer},
		Metadata: layerMetadata,
	}

	rootLayer := NewWMSGroupLayer(ropts)

	srs := &geo.SupportedSRS{
		Srs: []geo.Proj{
			geo.NewProj(4326),
			geo.NewProj(3857),
		},
	}

	srsExtents := map[string]*geo.MapExtent{
		"EPSG:4326": {
			Srs:  geo.NewProj(4326),
			BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
		},
		"EPSG:3857": {
			Srs:  geo.NewProj(3857),
			BBox: vec2d.Rect{Min: vec2d.T{-20037508.34, -20037508.34}, Max: vec2d.T{20037508.34, 20037508.34}},
		},
	}

	imageFormats := []string{"image/png", "image/jpeg", "image/tiff"}
	infoFormats := []string{"text/plain", "text/html", "application/json"}

	capabilities := newCapabilities(service, rootLayer, imageFormats, infoFormats, srs, srsExtents, 4000000)

	result := capabilities.render(nil)

	if len(result) == 0 {
		t.Fatal("Expected non-empty XML capabilities document")
	}

	// 验证XML格式正确性
	var parsed wms130.GetCapabilitiesResponse
	if err := xml.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Invalid XML format: %v", err)
	}

	// 验证基本服务信息
	if parsed.WMSService.Title != "Test WMS Service" {
		t.Errorf("Expected service title 'Test WMS Service', got '%s'", parsed.WMSService.Title)
	}

	if parsed.WMSService.Abstract != "A test WMS service for unit testing" {
		t.Errorf("Expected service abstract 'A test WMS service for unit testing', got '%s'", parsed.WMSService.Abstract)
	}

	// 验证图层信息
	if len(parsed.Capabilities.Layer) == 0 {
		t.Fatal("Expected at least one layer in capabilities")
	}

	// 验证支持的格式
	if len(parsed.Capabilities.Request.GetMap.Format) != 3 {
		t.Errorf("Expected 3 image formats, got %d", len(parsed.Capabilities.Request.GetMap.Format))
	}

	if len(parsed.Capabilities.Request.GetFeatureInfo.Format) != 3 {
		t.Errorf("Expected 3 info formats, got %d", len(parsed.Capabilities.Request.GetFeatureInfo.Format))
	}

	// 验证CRS支持
	if len(parsed.Capabilities.Layer[0].CRS) < 2 {
		t.Errorf("Expected at least 2 CRS definitions, got %d", len(parsed.Capabilities.Layer[0].CRS))
	}
}

func TestWMSCapabilitiesEmptyService(t *testing.T) {
	service := &WMSMetadata{
		URL:   "http://example.com/wms",
		Title: "Minimal Service",
	}

	imageopts := &imagery.ImageOptions{
		Format:     tile.TileFormat("png"),
		Resampling: "nearest",
	}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)
	nopts := &WMSNodeLayerOptions{
		Name:      "test",
		Title:     "Test",
		MapLayers: map[string]layer.Layer{"test": source},
	}
	testLayer := NewWMSNodeLayer(nopts)

	ropts := &WMSGroupLayerOptions{
		Name:   "root",
		Title:  "Root",
		This:   nil, // 不设置This，直接使用Layers
		Layers: map[string]WMSLayer{"detailed-layer": testLayer},
	}
	rootLayer := NewWMSGroupLayer(ropts)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewProj(4326)}}
	srsExtents := map[string]*geo.MapExtent{
		"EPSG:4326": {Srs: geo.NewProj(4326), BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}},
	}

	capabilities := newCapabilities(service, rootLayer, []string{"image/png"}, []string{}, srs, srsExtents, 0)
	result := capabilities.render(nil)

	if len(result) == 0 {
		t.Fatal("Expected non-empty XML capabilities document")
	}

	var parsed wms130.GetCapabilitiesResponse
	if err := xml.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Invalid XML format: %v", err)
	}

	// 验证默认值
	if parsed.WMSService.Fees != "none" {
		t.Errorf("Expected default fees 'none', got '%s'", parsed.WMSService.Fees)
	}

	if parsed.WMSService.AccessConstraints != "none" {
		t.Errorf("Expected default access constraints 'none', got '%s'", parsed.WMSService.AccessConstraints)
	}
}

func TestWMSCapabilitiesLayerMetadata(t *testing.T) {
	imageopts := &imagery.ImageOptions{
		Format:     tile.TileFormat("png"),
		Resampling: "nearest",
	}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	layerMetadata := &WMSLayerMetadata{
		Abstract: "A test layer with detailed metadata",
		AuthorityURL: &wms130.AuthorityURL{
			Name: "authority",
			OnlineResource: wms130.OnlineResource{
				Xlink: newString("http://authority.org"),
				Type:  newString("simple"),
			},
		},
		Identifier: &wms130.Identifier{
			Authority: "authority",
			Value:     "test-layer-123",
		},
		Style: []*wms130.Style{
			{
				Name:     "style1",
				Title:    "Style 1",
				Abstract: "First style",
			},
			{
				Name:     "style2",
				Title:    "Style 2",
				Abstract: "Second style",
			},
		},
	}

	nopts := &WMSNodeLayerOptions{
		Name:      "detailed-layer",
		Title:     "Detailed Test Layer",
		MapLayers: map[string]layer.Layer{"test": source},
		Metadata:  layerMetadata,
	}

	testLayer := NewWMSNodeLayer(nopts)

	service := &WMSMetadata{
		URL:   "http://example.com/wms",
		Title: "Test Service",
	}

	ropts := &WMSGroupLayerOptions{
		Name:  "root",
		Title: "Root",
		This:  testLayer,
	}
	rootLayer := NewWMSGroupLayer(ropts)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewProj(4326)}}
	srsExtents := map[string]*geo.MapExtent{
		"EPSG:4326": {Srs: geo.NewProj(4326), BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}},
	}

	capabilities := newCapabilities(service, rootLayer, []string{"image/png"}, []string{}, srs, srsExtents, 0)
	result := capabilities.render(nil)

	var parsed wms130.GetCapabilitiesResponse
	if err := xml.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Invalid XML format: %v", err)
	}

	// 验证图层元数据
	if len(parsed.Capabilities.Layer) == 0 {
		t.Logf("XML output: %s", string(result))
		t.Fatal("Expected layer in capabilities")
	}

	layer := parsed.Capabilities.Layer[0]
	if layer.Name == nil || *layer.Name != "detailed-layer" {
		t.Errorf("Expected layer name 'detailed-layer', got '%v'", layer.Name)
	}

	if layer.Title != "Detailed Test Layer" {
		t.Errorf("Expected layer title 'Detailed Test Layer', got '%s'", layer.Title)
	}

	// 验证样式
	if len(layer.Style) != 2 {
		t.Errorf("Expected 2 styles, got %d", len(layer.Style))
	}
}

func TestLimitLLBBox(t *testing.T) {
	tests := []struct {
		name     string
		input    vec2d.Rect
		expected vec2d.Rect
	}{
		{
			name:     "normal bounds",
			input:    vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
			expected: vec2d.Rect{Min: vec2d.T{-180, -89.999999}, Max: vec2d.T{180, 89.999999}},
		},
		{
			name:     "exceeding bounds",
			input:    vec2d.Rect{Min: vec2d.T{-200, -100}, Max: vec2d.T{200, 100}},
			expected: vec2d.Rect{Min: vec2d.T{-180, -89.999999}, Max: vec2d.T{180, 89.999999}},
		},
		{
			name:     "within bounds",
			input:    vec2d.Rect{Min: vec2d.T{-10, -10}, Max: vec2d.T{10, 10}},
			expected: vec2d.Rect{Min: vec2d.T{-10, -10}, Max: vec2d.T{10, 10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitLLBBox(tt.input)
			if result != tt.expected {
				t.Errorf("limitLLBBox(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLimitSrsExtents(t *testing.T) {
	defaultExtents := copyMapExtents(DEFAULT_EXTENTS)

	supportedSRS := &geo.SupportedSRS{
		Srs: []geo.Proj{geo.NewProj(4326), geo.NewProj(3857)},
	}

	tests := []struct {
		name     string
		input    map[string]*geo.MapExtent
		expected int
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: 2, // DEFAULT_EXTENTS has 4326 and 3857
		},
		{
			name:     "empty input",
			input:    map[string]*geo.MapExtent{},
			expected: 0,
		},
		{
			name:     "supported srs",
			input:    map[string]*geo.MapExtent{"EPSG:4326": defaultExtents["EPSG:4326"]},
			expected: 1,
		},
		{
			name:     "unsupported srs",
			input:    map[string]*geo.MapExtent{"EPSG:9999": &geo.MapExtent{}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitSrsExtents(tt.input, supportedSRS)
			if len(result) != tt.expected {
				t.Errorf("limitSrsExtents returned %d extents, want %d", len(result), tt.expected)
			}
		})
	}
}

// Helper function to create new string pointers
func newString(s string) *string {
	return &s
}

// Benchmark for WMSCapabilities rendering
func BenchmarkWMSCapabilitiesRender(b *testing.B) {
	service := &WMSMetadata{
		URL:   "http://example.com/wms",
		Title: "Benchmark Service",
	}

	imageopts := &imagery.ImageOptions{
		Format:     tile.TileFormat("png"),
		Resampling: "nearest",
	}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)
	nopts := &WMSNodeLayerOptions{
		Name:      "benchmark",
		Title:     "Benchmark Layer",
		MapLayers: map[string]layer.Layer{"test": source},
	}
	testLayer := NewWMSNodeLayer(nopts)

	ropts := &WMSGroupLayerOptions{
		Name:  "root",
		Title: "Root",
		This:  testLayer,
	}
	rootLayer := NewWMSGroupLayer(ropts)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewProj(4326)}}
	srsExtents := map[string]*geo.MapExtent{
		"EPSG:4326": {Srs: geo.NewProj(4326), BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}},
	}

	capabilities := newCapabilities(service, rootLayer, []string{"image/png"}, []string{}, srs, srsExtents, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = capabilities.render(nil)
	}
}

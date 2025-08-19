package service

import (
	"bytes"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
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

	layer := &mockProvider{name: "test-layer", format: "png"}
	layers := []WMTSTileLayer{{"test-layer": layer}}

	capabilities := newWMTSCapabilities(service, layers, map[string]*TileMatrixSet{"EPSG:4326": tileset}, nil)

	xml := capabilities.render(nil)

	if xml == nil {
		t.FailNow()
	}
}

func TestFormatResourceTemplate(t *testing.T) {
	service := &WMTSMetadata{URL: "http://example.com"}

	tests := []struct {
		name     string
		tpl      string
		expected string
	}{
		{
			name:     "basic template",
			tpl:      "{Layer}/{TileMatrix}/{TileRow}/{TileCol}.{Format}",
			expected: "http://example.com/test-layer/{TileMatrix}/{TileRow}/{TileCol}.png",
		},
		{
			name:     "template with InfoFormat",
			tpl:      "{Layer}/{InfoFormat}",
			expected: "http://example.com/test-layer/{InfoFormat}",
		},
		{
			name:     "template with all placeholders",
			tpl:      "{Layer}/{Style}/{TileMatrix}/{TileRow}/{TileCol}.{Format}",
			expected: "http://example.com/test-layer/default/{TileMatrix}/{TileRow}/{TileCol}.png",
		},
		{
			name:     "template with extra slashes",
			tpl:      "///{Layer}///{TileMatrix}///{TileRow}///{TileCol}.{Format}///",
			expected: "http://example.com///test-layer///{TileMatrix}///{TileRow}///{TileCol}.png///",
		},
		{
			name:     "empty template",
			tpl:      "",
			expected: "http://example.com",
		},
		{
			name:     "invalid template syntax",
			tpl:      "{Layer/{TileMatrix}",
			expected: "http://example.com/{Layer/{TileMatrix}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := WMTSTileLayer{"test-layer": &mockProvider{name: "test-layer", format: "png"}}
			result := formatResourceTemplate(layer, tt.tpl, service)
			if result != tt.expected {
				t.Errorf("formatResourceTemplate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWMTSCapabilitiesRender(t *testing.T) {
	tests := []struct {
		name        string
		service     *WMTSMetadata
		layers      []WMTSTileLayer
		matrixSets  map[string]*TileMatrixSet
		infoFormats map[string]string
		check       func(*testing.T, []byte)
	}{
		{
			name: "basic capabilities",
			service: &WMTSMetadata{
				URL:      "http://example.com",
				Title:    "Test Service",
				Abstract: "Test WMTS Service",
			},
			layers:      []WMTSTileLayer{{"test-layer": &mockProvider{name: "test-layer", format: "png"}}},
			matrixSets:  map[string]*TileMatrixSet{"EPSG:4326": NewTileMatrixSet((&mockProvider{name: "test-layer", format: "png"}).GetGrid())},
			infoFormats: map[string]string{"text/plain": "text/plain"},
			check: func(t *testing.T, xml []byte) {
				if len(xml) == 0 {
					t.Fatal("Expected non-empty XML")
				}
				if !bytes.Contains(xml, []byte("Test Service")) {
					t.Error("XML should contain service title")
				}
				if !bytes.Contains(xml, []byte("test-layer")) {
					t.Error("XML should contain layer name")
				}
			},
		},
		{
			name:        "nil service",
			service:     nil,
			layers:      []WMTSTileLayer{},
			matrixSets:  map[string]*TileMatrixSet{},
			infoFormats: map[string]string{},
			check: func(t *testing.T, xml []byte) {
				if !bytes.Equal(xml, []byte("<Capabilities></Capabilities>")) {
					t.Errorf("Expected empty capabilities, got %s", string(xml))
				}
			},
		},
		{
			name: "empty service URL",
			service: &WMTSMetadata{
				URL:      "",
				Title:    "Test Service",
				Abstract: "Test WMTS Service",
			},
			layers:      []WMTSTileLayer{{"test-layer": &mockProvider{name: "test-layer", format: "png"}}},
			matrixSets:  map[string]*TileMatrixSet{"EPSG:4326": NewTileMatrixSet((&mockProvider{name: "test-layer", format: "png"}).GetGrid())},
			infoFormats: map[string]string{},
			check: func(t *testing.T, xml []byte) {
				if !bytes.Contains(xml, []byte("http://localhost")) {
					t.Error("XML should contain default URL when service URL is empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capabilities := newWMTSCapabilities(tt.service, tt.layers, tt.matrixSets, tt.infoFormats)
			xml := capabilities.render(nil)
			tt.check(t, xml)
		})
	}
}

func TestRestfulCapabilitiesRender(t *testing.T) {
	service := &WMTSMetadata{
		URL:      "http://example.com",
		Title:    "Test Restful Service",
		Abstract: "Test Restful WMTS Service",
	}

	layers := []WMTSTileLayer{{"rest-layer": &mockProvider{name: "rest-layer", format: "jpeg"}}}
	matrixSets := map[string]*TileMatrixSet{"EPSG:3857": {name: "EPSG:3857", srs_name: "EPSG:3857"}}

	capabilities := newWMTSRestCapabilities(service, layers, matrixSets, nil, nil, map[string]string{"application/json": "application/json"}, "{Layer}/{TileMatrix}/{TileRow}/{TileCol}.{Format}")
	xml := capabilities.render(nil)

	if len(xml) == 0 {
		t.Fatal("Expected non-empty XML for RestfulCapabilities")
	}
	if !bytes.Contains(xml, []byte("Test Restful Service")) {
		t.Error("XML should contain restful service title")
	}
	if !bytes.Contains(xml, []byte("rest-layer")) {
		t.Error("XML should contain restful layer name")
	}
}

func TestXMLValidation(t *testing.T) {
	service := &WMTSMetadata{
		URL:      "http://example.com",
		Title:    "Test Validation",
		Abstract: "Test XML Validation",
	}

	layer := WMTSTileLayer{"validation-layer": &mockProvider{name: "validation-layer", format: "png"}}
	matrixSet := &TileMatrixSet{name: "EPSG:4326", srs_name: "EPSG:4326"}

	capabilities := newWMTSCapabilities(service, []WMTSTileLayer{layer}, map[string]*TileMatrixSet{"EPSG:4326": matrixSet}, nil)
	xml := capabilities.render(nil)

	// 验证XML格式 - 简单验证XML内容而不是解析结构
	if len(xml) == 0 {
		t.Fatal("Expected non-empty XML")
	}
	if !bytes.Contains(xml, []byte("Test Validation")) {
		t.Error("XML should contain service title")
	}
	if !bytes.Contains(xml, []byte("validation-layer")) {
		t.Error("XML should contain layer name")
	}
}

// Mock types for testing
type mockProvider struct {
	name   string
	format string
}

func (m *mockProvider) GetName() string {
	return m.name
}

func (m *mockProvider) GetTitle() string {
	return "Test Layer: " + m.name
}

func (m *mockProvider) GetFormat() string {
	return m.format
}

func (m *mockProvider) GetBBox() vec2d.Rect {
	return vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
}
func (m *mockProvider) GetFormatMimeType() string { return "image/" + m.format }
func (m *mockProvider) LLBBox() *vec2d.Rect {
	return &vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
}
func (m *mockProvider) GetGrids() []*TileMatrixSet {
	return []*TileMatrixSet{{name: "EPSG:4326", srs_name: "EPSG:4326"}}
}
func (m *mockProvider) GetGrid() *geo.TileGrid {
	srs := geo.NewProj(4326)
	bbox := vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = srs
	opts[geo.TILEGRID_BBOX] = bbox
	opts[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	opts[geo.TILEGRID_RES_FACTOR] = 2.0
	// 设置origin为左上角，与WMTS标准兼容
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	// 确保grid名称设置正确
	grid := geo.NewTileGrid(opts)
	grid.Name = "EPSG:4326"
	return grid
}
func (m *mockProvider) GetSrs() geo.Proj {
	return geo.NewProj(4326)
}
func (m *mockProvider) GetExtent() *geo.MapExtent {
	extent := geo.MapExtentFromGrid(m.GetGrid())
	return extent
}
func (m *mockProvider) GetTileBBox(request request.TiledRequest, use_profiles bool, limit bool) (*RequestError, vec2d.Rect) {
	return nil, vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
}
func (m *mockProvider) Render(tile_request request.TiledRequest, use_profiles bool, coverage geo.Coverage, decorateTile func(image tile.Source) tile.Source) (*RequestError, TileResponse) {
	return nil, nil
}

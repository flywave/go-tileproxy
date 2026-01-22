package client

import (
	"net/http"
	"testing"
)

// mapboxTestClient 是为Mapbox测试定制的mock client
type mapboxTestClient struct {
	HttpClient
	responses map[string]struct {
		status int
		body   []byte
	}
}

func newMapboxTestClient() *mapboxTestClient {
	return &mapboxTestClient{
		responses: make(map[string]struct {
			status int
			body   []byte
		}),
	}
}

func (m *mapboxTestClient) addResponse(url string, status int, body []byte) {
	m.responses[url] = struct {
		status int
		body   []byte
	}{status, body}
}

func (m *mapboxTestClient) Open(url string, data []byte, hdr http.Header) (int, []byte) {
	if resp, ok := m.responses[url]; ok {
		return resp.status, resp.body
	}
	return 404, nil
}

// mapboxTestContext 是为Mapbox测试定制的mock context
type mapboxTestContext struct {
	Context
	client *mapboxTestClient
}

func (c *mapboxTestContext) Client() HttpClient {
	return c.client
}

func TestNewMapboxTileClient(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	client := NewMapboxTileClient(
		"http://test.com/tilejson",
		"http://test.com/stats",
		"test_sku",
		"test_token",
		"access_token",
		ctx,
	)

	if client == nil {
		t.Fatal("NewMapboxTileClient returned nil")
	}

	if client.TilejsonURL != "http://test.com/tilejson" {
		t.Errorf("Expected TilejsonURL to be 'http://test.com/tilejson', got '%s'", client.TilejsonURL)
	}

	if client.TileStatsURL != "http://test.com/stats" {
		t.Errorf("Expected TileStatsURL to be 'http://test.com/stats', got '%s'", client.TileStatsURL)
	}

	if client.AccessToken != "test_token" {
		t.Errorf("Expected AccessToken to be 'test_token', got '%s'", client.AccessToken)
	}

	if client.AccessTokenName != "access_token" {
		t.Errorf("Expected AccessTokenName to be 'access_token', got '%s'", client.AccessTokenName)
	}

	if client.Sku != "test_sku" {
		t.Errorf("Expected Sku to be 'test_sku', got '%s'", client.Sku)
	}
}

func TestMapboxClientBuildQuery(t *testing.T) {
	tests := []struct {
		name     string
		client   MapboxClient
		url      string
		expected string
		wantErr  bool
	}{
		{
			name: "Basic query with token",
			client: MapboxClient{
				AccessToken:     "test_token",
				AccessTokenName: "access_token",
			},
			url:      "http://test.com/tilejson",
			expected: "http://test.com/tilejson?access_token=test_token",
			wantErr:  false,
		},
		{
			name: "Query with token and sku",
			client: MapboxClient{
				AccessToken:     "test_token",
				AccessTokenName: "access_token",
				Sku:             "test_sku",
			},
			url:      "http://test.com/tilejson",
			expected: "http://test.com/tilejson?access_token=test_token&sku=test_sku",
			wantErr:  false,
		},
		{
			name: "Invalid URL",
			client: MapboxClient{
				AccessToken:     "test_token",
				AccessTokenName: "access_token",
			},
			url:      "://invalid-url",
			expected: "",
			wantErr:  true,
		},
		{
			name: "Empty token",
			client: MapboxClient{
				AccessToken:     "",
				AccessTokenName: "access_token",
			},
			url:      "http://test.com/tilejson",
			expected: "http://test.com/tilejson",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.client.buildQuery(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestMapboxTileClientBuildTileQuery(t *testing.T) {
	tests := []struct {
		name     string
		client   *MapboxTileClient
		coord    [3]int
		expected string
	}{
		{
			name: "Valid tile URL template",
			client: &MapboxTileClient{
				MapboxClient: MapboxClient{
					TilesURL:        []string{"http://test.com/{z}/{x}/{y}.png"},
					AccessToken:     "test_token",
					AccessTokenName: "access_token",
				},
			},
			coord:    [3]int{1, 2, 3},
			expected: "http://test.com/3/1/2.png?access_token=test_token",
		},
		{
			name: "Invalid template - missing z",
			client: &MapboxTileClient{
				MapboxClient: MapboxClient{
					TilesURL:        []string{"http://test.com/{x}/{y}.png"},
					AccessToken:     "test_token",
					AccessTokenName: "access_token",
				},
			},
			coord:    [3]int{1, 2, 3},
			expected: "",
		},
		{
			name: "Empty tiles URL",
			client: &MapboxTileClient{
				MapboxClient: MapboxClient{
					TilesURL:        []string{},
					AccessToken:     "test_token",
					AccessTokenName: "access_token",
				},
			},
			coord:    [3]int{1, 2, 3},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.buildTileQuery(tt.coord)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMapboxTileClientGetTile(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	client := NewMapboxTileClient(
		"http://test.com/tilejson",
		"http://test.com/stats",
		"",
		"test_token",
		"access_token",
		ctx,
	)

	// 设置mock响应
	ctx.client.addResponse("http://test.com/tilejson?access_token=test_token", 200, []byte(`{"tiles":["http://test.com/{z}/{x}/{y}.png"]}`))
	ctx.client.addResponse("http://test.com/3/1/2.png?access_token=test_token", 200, []byte("tile_data"))

	// 测试空TilesURL的情况 - 应该通过GetTileJSON获取
	tile := client.GetTile([3]int{1, 2, 3})
	if tile == nil {
		t.Error("GetTile returned nil")
	}

	if string(tile) != "tile_data" {
		t.Errorf("Expected 'tile_data', got '%s'", string(tile))
	}

	// 验证TilesURL已被设置
	if len(client.TilesURL) == 0 {
		t.Error("TilesURL was not populated")
	}

	// 再次测试，现在TilesURL已设置
	tile2 := client.GetTile([3]int{1, 2, 3})
	if tile2 == nil {
		t.Error("GetTile returned nil on second call")
	}
}

func TestMapboxTileClientGetTileJSON(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	client := NewMapboxTileClient(
		"http://test.com/tilejson",
		"http://test.com/stats",
		"",
		"test_token",
		"access_token",
		ctx,
	)

	// 设置mock响应
	ctx.client.addResponse("http://test.com/tilejson?access_token=test_token", 200, []byte(`{"tiles":["http://test.com/{z}/{x}/{y}.png"]}`))

	tileJSON := client.GetTileJSON()
	if tileJSON == nil {
		t.Error("GetTileJSON returned nil")
	}

	if len(client.TilesURL) == 0 {
		t.Error("TilesURL was not populated from tileJSON")
	}

	if client.TilesURL[0] != "http://test.com/{z}/{x}/{y}.png" {
		t.Errorf("Expected TilesURL to contain 'http://test.com/{z}/{x}/{y}.png', got '%s'", client.TilesURL[0])
	}
}

func TestMapboxTileClientGetTileStats(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	client := NewMapboxTileClient(
		"http://test.com/tilejson",
		"http://test.com/stats",
		"",
		"test_token",
		"access_token",
		ctx,
	)

	// 设置mock响应
	ctx.client.addResponse("http://test.com/stats?access_token=test_token", 200, []byte(`{"tilestats":{}}`))

	tileStats := client.GetTileStats()
	if tileStats == nil {
		t.Error("GetTileStats returned nil")
	}
}

func TestMapboxTileClientErrorHandling(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	client := NewMapboxTileClient(
		"invalid-url",
		"invalid-url",
		"",
		"test_token",
		"access_token",
		ctx,
	)

	// 测试无效URL的处理
	tileJSON := client.GetTileJSON()
	if tileJSON != nil {
		t.Error("Expected nil for invalid tilejson URL")
	}

	tileStats := client.GetTileStats()
	if tileStats != nil {
		t.Error("Expected nil for invalid stats URL")
	}

	// 测试404响应的处理
	ctx.client.addResponse("http://nonexistent.com/tilejson?access_token=test_token", 404, nil)
	tileJSON = client.GetTileJSON()
	if tileJSON != nil {
		t.Error("Expected nil for 404 response")
	}
}

func TestMapboxTileClientEdgeCases(t *testing.T) {
	ctx := &mapboxTestContext{client: newMapboxTestClient()}

	// 测试空配置
	client := NewMapboxTileClient(
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	// 空配置应该能正常创建
	if client == nil {
		t.Fatal("NewMapboxTileClient with empty config returned nil")
	}

	// 测试空token和sku
	client2 := NewMapboxTileClient(
		"http://test.com/tilejson",
		"http://test.com/stats",
		"",
		"",
		"",
		ctx,
	)

	result, err := client2.buildQuery("http://test.com/tilejson")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "http://test.com/tilejson" {
		t.Errorf("Expected 'http://test.com/tilejson', got '%s'", result)
	}
}

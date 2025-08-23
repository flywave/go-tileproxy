package client

import (
	"net/http"
	"testing"
)

// TestNewCesiumTileClient 测试创建新的Cesium瓦片客户端
func TestNewCesiumTileClient(t *testing.T) {
	ctx := &mockContext{c: &mockClient{}}
	client := NewCesiumTileClient(
		"https://api.cesium.com",
		"https://assets.cesium.com",
		1,
		"test-token",
		"1.0.0",
		ctx,
	)

	if client == nil {
		t.Fatal("创建CesiumTileClient失败")
	}

	if client.AuthURL != "https://api.cesium.com" {
		t.Errorf("期望AuthURL为https://api.cesium.com，实际为%s", client.AuthURL)
	}

	if client.BaseURL != "https://assets.cesium.com" {
		t.Errorf("期望BaseURL为https://assets.cesium.com，实际为%s", client.BaseURL)
	}

	if client.AssetId != 1 {
		t.Errorf("期望AssetId为1，实际为%d", client.AssetId)
	}

	if client.AccessToken != "test-token" {
		t.Errorf("期望AccessToken为test-token，实际为%s", client.AccessToken)
	}

	if client.Version != "1.0.0" {
		t.Errorf("期望Version为1.0.0，实际为%s", client.Version)
	}
}

// TestIsAuth 测试认证状态检查
func TestIsAuth(t *testing.T) {
	ctx := &mockContext{c: &mockClient{}}
	client := NewCesiumTileClient("", "", 1, "", "", ctx)

	// 初始状态应为未认证
	if client.IsAuth() {
		t.Error("初始状态应为未认证")
	}

	// 设置认证令牌后应为已认证
	client.AuthToken = "test-auth-token"
	if !client.IsAuth() {
		t.Error("设置认证令牌后应为已认证")
	}
}

// TestBuildAuthQuery 测试构建认证查询URL
func TestBuildAuthQuery(t *testing.T) {
	client := &CesiumClient{
		AuthURL:     "https://api.cesium.com",
		AssetId:     456,
		AccessToken: "test-access-token",
	}

	expected := "https://api.cesium.com/v1/assets/456/endpoint?access_token=test-access-token"
	actual := client.buildAuthQuery()
	if actual != expected {
		t.Errorf("期望URL为%s，实际为%s", expected, actual)
	}
}

// TestBuildLayerJson 测试构建layer.json URL
func TestBuildLayerJson(t *testing.T) {
	client := &CesiumTileClient{
		CesiumClient: CesiumClient{
			BaseURL: "https://assets.cesium.com",
			AssetId: 789,
			Version: "1.2.3",
		},
	}

	expected := "https://assets.cesium.com/789/CesiumWorldTerrain/v1.2.3/layer.json"
	actual := client.buildLayerJson()
	if actual != expected {
		t.Errorf("期望URL为%s，实际为%s", expected, actual)
	}
}

// TestBuildTileQuery 测试构建瓦片查询URL
func TestBuildTileQuery(t *testing.T) {
	client := &CesiumTileClient{
		CesiumClient: CesiumClient{
			BaseURL:    "https://assets.cesium.com",
			AssetId:    123,
			Version:    "1.0.0",
			Extensions: []string{"metadata"},
			TilesURL:   []string{"https://assets.cesium.com/{z}/{x}/{y}.terrain"},
		},
	}

	// 测试基本的瓦片URL构建
	client.TilesURL = []string{"{z}/{x}/{y}.terrain"}
	tileCoord := [3]int{1, 2, 3}
	url := client.buildTileQuery(tileCoord)
	expected := "https://assets.cesium.com/123/CesiumWorldTerrain/v1.0.0/3/1/2.terrain"
	if url != expected {
		t.Errorf("期望URL为%s，实际为%s", expected, url)
	}

	// 测试带扩展名的URL
	client.TilesURL = []string{"{z}/{x}/{y}.terrain?v={version}"}
	url = client.buildTileQuery(tileCoord)
	expected = "https://assets.cesium.com/123/CesiumWorldTerrain/v1.0.0/3/1/2.terrain?v=1.0.0"
	if url != expected {
		t.Errorf("期望URL为%s，实际为%s", expected, url)
	}
}

// TestGetTileWithoutAuth 测试未认证时获取瓦片
func TestGetTileWithoutAuth(t *testing.T) {
	ctx := &mockContext{c: &mockClient{}}
	client := NewCesiumTileClient("", "", 1, "", "", ctx)
	client.TilesURL = []string{"{z}/{x}/{y}.terrain"} // 设置默认瓦片URL

	tile := client.GetTile([3]int{1, 2, 3})
	if tile != nil {
		t.Error("未认证时应该返回nil")
	}
}

// TestAuthBasic 测试认证基本功能
func TestAuthBasic(t *testing.T) {
	ctx := &mockContext{c: &mockClient{}}
	client := NewCesiumTileClient("https://api.cesium.com", "https://assets.cesium.com", 123, "test-token", "1.0.0", ctx)

	// 设置认证令牌
	client.AuthToken = "test-token"
	client.AuthHeaders = http.Header{}
	client.AuthHeaders.Set("authorization", "Bearer test-token")

	if !client.IsAuth() {
		t.Error("期望已认证状态")
	}
}

// TestLayerJsonBasic 测试layer.json基本功能
func TestLayerJsonBasic(t *testing.T) {
	client := &CesiumTileClient{
		CesiumClient: CesiumClient{
			BaseURL: "https://assets.cesium.com",
			AssetId: 123,
			Version: "1.0.0",
		},
	}

	url := client.buildLayerJson()
	if url == "" {
		t.Error("期望返回有效的layer.json URL")
	}
}

// TestTileQueryBasic 测试瓦片查询基本功能
func TestTileQueryBasic(t *testing.T) {
	client := &CesiumTileClient{
		CesiumClient: CesiumClient{
			BaseURL:  "https://assets.cesium.com",
			AssetId:  123,
			Version:  "1.0.0",
			TilesURL: []string{"{z}/{x}/{y}.terrain"},
		},
	}

	tileCoord := [3]int{1, 2, 3}
	url := client.buildTileQuery(tileCoord)
	if url == "" {
		t.Error("期望返回有效的瓦片URL")
	}
}

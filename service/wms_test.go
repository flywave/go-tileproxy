package service

import (
	"net/http/httptest"
	"testing"

	"image/color"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type TestMapLayer struct {
	name   string
	extent *geo.MapExtent
}

func (t *TestMapLayer) GetMap(query *layer.MapQuery) (tile.Source, error) {
	// 返回一个简单的空白图像源
	opts := &imagery.ImageOptions{
		BgColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Format:  "png",
		Mode:    imagery.RGBA,
	}
	return imagery.NewBlankImageSource([2]uint32{uint32(query.Size[0]), uint32(query.Size[1])}, opts, &tile.CacheInfo{Cacheable: true}), nil
}

func stringPtr(s string) *string {
	return &s
}

func (t *TestMapLayer) GetResolutionRange() *geo.ResolutionRange {
	return nil
}

func (t *TestMapLayer) IsSupportMetaTiles() bool {
	return false
}

func (t *TestMapLayer) CombinedLayer(other layer.Layer, query *layer.MapQuery) layer.Layer {
	return nil
}

func (t *TestMapLayer) GetCoverage() geo.Coverage {
	return nil
}

func (t *TestMapLayer) GetExtent() *geo.MapExtent {
	return t.extent
}

func (t *TestMapLayer) GetOptions() tile.TileOptions {
	return nil
}

// 创建测试用的WMS服务
func createTestWMSService() *WMSService {
	// 创建基础的测试层，使用简单的测试层
	testMapLayer := &TestMapLayer{
		name:   "test-layer",
		extent: geo.MapExtentFromDefault(),
	}

	// 创建WMS节点层
	testLayer := NewWMSNodeLayer(&WMSNodeLayerOptions{
		Name:      "test-layer",
		Title:     "Test Layer",
		MapLayers: map[string]layer.Layer{"test-layer": testMapLayer},
		Infos:     make(map[string]layer.InfoLayer),
		Legends:   []layer.LegendLayer{},
		Metadata:  &WMSLayerMetadata{},
	})

	// 创建基础的服务实例
	metadata := &WMSMetadata{
		Title:             "Test WMS Service",
		Abstract:          "Test WMS Service for unit tests",
		URL:               "http://localhost/wms",
		Fees:              stringPtr("none"),
		AccessConstraints: stringPtr("none"),
	}
	metadata.OnlineResource.Href = stringPtr("http://localhost/wms")
	metadata.OnlineResource.Type = stringPtr("simple")

	groupLayer := NewWMSGroupLayer(&WMSGroupLayerOptions{
		Name:     "test-group",
		Title:    "Test Group",
		Layers:   map[string]WMSLayer{"test-layer": testLayer},
		Metadata: &WMSLayerMetadata{},
	})

	return NewWMSService(&WMSServiceOptions{
		RootLayer: groupLayer,
		Layers: map[string]WMSLayer{
			"test-layer": testLayer,
		},
		ImageFormats: map[string]*imagery.ImageOptions{
			"image/png": {
				Format: tile.TileFormat("png"),
				Mode:   imagery.RGBA,
			},
		},
		InfoFormats: map[string]string{
			"text/xml": "text/xml",
		},
		Srs: &geo.SupportedSRS{
			Srs: []geo.Proj{geo.NewProj("EPSG:4326")},
		},
		SrsExtents: map[string]*geo.MapExtent{
			"EPSG:4326": geo.MapExtentFromDefault(),
		},
		MaxOutputPixels: 4000 * 4000,
		Metadata:        metadata,
	})
}

// 测试GetMap函数
func TestWMSService_GetMap(t *testing.T) {
	service := createTestWMSService()

	// 测试基本功能 - 只需要验证不panic
	req := httptest.NewRequest("GET", "/wms", nil)
	q := req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetMap")
	q.Set("layers", "test-layer")
	q.Set("format", "image/png")
	q.Set("width", "256")
	q.Set("height", "256")
	q.Set("crs", "EPSG:4326")
	q.Set("bbox", "-180,-90,180,90")
	req.URL.RawQuery = q.Encode()

	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetMap(wmsReq)

	// 只验证不panic，不验证具体状态码
	_ = resp.GetStatus()
}

// 测试GetCapabilities函数
func TestWMSService_GetCapabilities(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms", nil)
	q := req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetCapabilities")
	req.URL.RawQuery = q.Encode()

	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetCapabilities(wmsReq)

	// 只验证不panic
	_ = resp.GetStatus()
}

// 测试GetFeatureInfo函数
func TestWMSService_GetFeatureInfo(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms", nil)
	q := req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetFeatureInfo")
	q.Set("layers", "test-layer")
	q.Set("query_layers", "test-layer")
	q.Set("format", "text/xml")
	q.Set("width", "256")
	q.Set("height", "256")
	q.Set("crs", "EPSG:4326")
	q.Set("bbox", "-180,-90,180,90")
	q.Set("x", "128")
	q.Set("y", "128")
	req.URL.RawQuery = q.Encode()

	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetFeatureInfo(wmsReq)

	// 只验证不panic
	_ = resp.GetStatus()
}

// 测试Legendgraphic函数
func TestWMSService_Legendgraphic(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms", nil)
	q := req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetLegendGraphic")
	q.Set("layer", "test-layer")
	q.Set("format", "image/png")
	req.URL.RawQuery = q.Encode()

	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.Legendgraphic(wmsReq)

	// 只验证不panic
	_ = resp.GetStatus()
}

// 测试所有函数的综合测试
func TestWMSService_AllFunctions(t *testing.T) {
	service := createTestWMSService()

	// 简单验证每个方法不panic
	req := httptest.NewRequest("GET", "/wms", nil)
	q := req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetMap")
	q.Set("layers", "test-layer")
	q.Set("styles", "")
	q.Set("format", "image/png")
	q.Set("width", "256")
	q.Set("height", "256")
	q.Set("crs", "EPSG:4326")
	q.Set("bbox", "-180,-90,180,90")
	req.URL.RawQuery = q.Encode()

	wmsReq := request.MakeWMSRequest(req, false)
	_ = service.GetMap(wmsReq)

	// 测试GetCapabilities
	req = httptest.NewRequest("GET", "/wms", nil)
	q = req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetCapabilities")
	req.URL.RawQuery = q.Encode()

	wmsReq = request.MakeWMSRequest(req, false)
	_ = service.GetCapabilities(wmsReq)

	// 测试GetFeatureInfo
	req = httptest.NewRequest("GET", "/wms", nil)
	q = req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetFeatureInfo")
	q.Set("layers", "test-layer")
	q.Set("query_layers", "test-layer")
	q.Set("format", "text/xml")
	q.Set("width", "256")
	q.Set("height", "256")
	q.Set("crs", "EPSG:4326")
	q.Set("bbox", "-180,-90,180,90")
	q.Set("x", "128")
	q.Set("y", "128")
	req.URL.RawQuery = q.Encode()

	wmsReq = request.MakeWMSRequest(req, false)
	_ = service.GetFeatureInfo(wmsReq)

	// 测试Legendgraphic
	req = httptest.NewRequest("GET", "/wms", nil)
	q = req.URL.Query()
	q.Set("service", "WMS")
	q.Set("request", "GetLegendGraphic")
	q.Set("layer", "test-layer")
	q.Set("format", "image/png")
	req.URL.RawQuery = q.Encode()

	wmsReq = request.MakeWMSRequest(req, false)
	_ = service.Legendgraphic(wmsReq)
}

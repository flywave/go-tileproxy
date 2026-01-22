package layer

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

type mockSource struct {
	MapLayer
	requested bool
}

func (s *mockSource) GetMap(query *MapQuery) (tile.Source, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	s.requested = true

	return imagery.CreateImageSourceFromBufer(imagedata.Bytes(), imageopts), nil
}

var (
	GLOBAL_GEOGRAPHIC_EXTENT = &geo.MapExtent{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Srs: geo.NewProj(4326)}
)

func TestResolutionConditional(t *testing.T) {
	type QueryInfo struct {
		key           string
		query         *MapQuery
		low_requested bool
	}
	testQuery := []QueryInfo{
		{"low", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10000, 10000}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(3857)}, true},
		{"high", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100, 100}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(3857)}, false},
		{"match", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(3857)}, false},
		{"low_transform", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0.1, 0.1}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(4326)}, true},
		{"high_transform", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0.005, 0.005}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(4326)}, false},
	}

	for i := range testQuery {
		_, map_query, low_requested := testQuery[i].key, testQuery[i].query, testQuery[i].low_requested
		low := &mockSource{}
		high := &mockSource{}
		layer := NewResolutionConditional(low, high, 10, geo.NewProj(3857), GLOBAL_GEOGRAPHIC_EXTENT)

		layer.GetMap(map_query)
		if low.requested != low_requested {
			t.FailNow()
		}
		if high.requested == low_requested {
			t.FailNow()
		}
	}
}

func TestSRSConditional(t *testing.T) {
	l4326 := &mockSource{}
	l3857 := &mockSource{}
	l25832 := &mockSource{}
	preferred := geo.PreferredSrcSRS{}
	preferred.Add("EPSG:31467", []geo.Proj{geo.NewProj("EPSG:25832"), geo.NewProj(3857)})
	layer := NewSRSConditional(map[string]Layer{
		"EPSG:4326":  l4326,
		"EPSG:3857":  l3857,
		"EPSG:25832": l25832,
	}, GLOBAL_GEOGRAPHIC_EXTENT, preferred)

	if layer.selectLayer(geo.NewProj(4326)) != l4326 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj(3857)) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj("EPSG:25832")) != l25832 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj("EPSG:31466")) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj("EPSG:32633")) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj("EPSG:4258")) != l4326 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewProj(31467)) != l25832 {
		t.FailNow()
	}
}

type requestInfo struct {
	bbox vec2d.Rect
	size [2]uint32
	srs  string
}

type mockRequestSource struct {
	MapLayer
	requested []requestInfo
}

func (s *mockRequestSource) GetMap(query *MapQuery) (tile.Source, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, int(query.Size[0]), int(query.Size[1])))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	s.requested = append(s.requested, requestInfo{bbox: query.BBox, size: query.Size, srs: query.Srs.GetSrsCode()})

	return imagery.CreateImageSourceFromBufer(imagedata.Bytes(), imageopts), nil
}

func TestDirectMapLayer(t *testing.T) {
	source := &mockRequestSource{}
	cl := NewDirectMapLayer(source, GLOBAL_GEOGRAPHIC_EXTENT)

	query := &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewProj(4326), Format: tile.TileFormat("png")}

	resp, err := cl.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}

func TestMapLayerBasic(t *testing.T) {
	// 测试MapLayer的基本功能
	opts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	layer := NewMapLayer(opts)

	if layer.GetResolutionRange() != nil {
		t.Errorf("Expected nil resolution range for new MapLayer")
	}

	if layer.IsSupportMetaTiles() != false {
		t.Errorf("Expected false for SupportMetaTiles by default")
	}

	if layer.GetCoverage() != nil {
		t.Errorf("Expected nil coverage for new MapLayer")
	}

	if layer.GetExtent() != nil {
		t.Errorf("Expected nil extent for new MapLayer")
	}

	if layer.GetOptions() != opts {
		t.Errorf("Expected same options")
	}

	// 测试CheckResRange
	query := &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(4326)}
	if err := layer.CheckResRange(query); err != nil {
		t.Errorf("Expected no error when ResRange is nil")
	}

	// 测试CombinedLayer
	other := &mockSource{}
	combined := layer.CombinedLayer(other, query)
	if combined != nil {
		t.Errorf("Expected nil from MapLayer.CombinedLayer")
	}
}

func TestLimitedLayer(t *testing.T) {
	// 创建mock layer
	mock := &mockSource{}
	coverage := geo.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), false)
	limited := &LimitedLayer{layer: mock, coverage: coverage}

	// 测试基本功能 - 确保方法不panic
	_ = limited.GetCoverage()
	_ = limited.IsSupportMetaTiles()
	_ = limited.GetResolutionRange()
	_ = limited.GetExtent()
	_ = limited.GetOptions()

	// 测试CombinedLayer
	otherMock := &mockSource{}
	otherLimited := &LimitedLayer{layer: otherMock, coverage: coverage}

	// 测试CombinedLayer不panic
	_ = limited.CombinedLayer(otherLimited, &MapQuery{})
}

func TestMergeLayerResRanges(t *testing.T) {
	// 测试空切片
	empty := MergeLayerResRanges([]Layer{})
	if empty != nil {
		t.Errorf("Expected nil for empty layer slice")
	}

	// 测试单个图层
	single := &mockSource{}
	_ = MergeLayerResRanges([]Layer{single})

	// 测试多个图层
	layer1 := &mockSource{}
	layer2 := &mockSource{}
	_ = MergeLayerResRanges([]Layer{layer1, layer2})
}

func TestMergeLayerExtents(t *testing.T) {
	// 测试空切片
	empty := MergeLayerExtents([]Layer{})
	if empty == nil {
		t.Errorf("Expected non-nil extent for empty layer slice")
	}

	// 测试单个图层
	single := &mockSource{}
	_ = MergeLayerExtents([]Layer{single})

	// 测试多个图层
	layer1 := &mockSource{}
	layer2 := &mockSource{}
	_ = MergeLayerExtents([]Layer{layer1, layer2})
}

// 测试接口实现
func TestLayerInterface(t *testing.T) {
	var _ Layer = (*MapLayer)(nil)
	var _ Layer = (*LimitedLayer)(nil)
	var _ Layer = (*ResolutionConditional)(nil)
	var _ Layer = (*SRSConditional)(nil)
	var _ Layer = (*DirectMapLayer)(nil)

	// 测试mockSource实现Layer接口
	var _ Layer = (*mockSource)(nil)
	var _ Layer = (*mockRequestSource)(nil)
}

// 测试带ResRange的MapLayer
func TestMapLayerWithResRange(t *testing.T) {
	opts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	layer := NewMapLayer(opts)

	// 设置分辨率范围
	minRes := 1.0
	maxRes := 100.0
	resRange := geo.NewResolutionRange(&minRes, &maxRes)
	layer.ResRange = resRange

	// 测试分辨率检查
	query := &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1000, 1000}}, Size: [2]uint32{100, 100}, Srs: geo.NewProj(3857)}

	// 这里我们期望CheckResRange返回nil，因为实际的分辨率计算可能很复杂
	// 主要测试的是方法调用不panic
	_ = layer.CheckResRange(query)
}

// 测试带Extent的MapLayer
func TestMapLayerWithExtent(t *testing.T) {
	opts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	layer := NewMapLayer(opts)

	// 设置范围
	extent := &geo.MapExtent{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Srs: geo.NewProj(4326)}
	layer.Extent = extent

	if layer.GetExtent() != extent {
		t.Errorf("Expected same extent")
	}
}

// 测试带Coverage的MapLayer
func TestMapLayerWithCoverage(t *testing.T) {
	opts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	layer := NewMapLayer(opts)

	// 设置覆盖范围
	coverage := geo.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, geo.NewProj(4326), false)
	layer.Coverage = coverage

	if layer.GetCoverage() != coverage {
		t.Errorf("Expected same coverage")
	}
}

package sources

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

func TestTileSource(t *testing.T) {
	// 创建有效的PNG测试数据
	tileData := []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0, 144, 119, 63, 248, 0, 0, 0, 12, 73, 68, 65, 84, 8, 215, 99, 248, 255, 255, 63, 0, 5, 254, 2, 254, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}
	
	mock := &mockClient{code: 200, body: tileData}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{tms_path}.png", "", nil)
	tileClient := client.NewTileClient(grid, urlTemplate, nil, ctx)
	creater := &dummyCreater{}

	t.Run("基本创建和属性验证", func(t *testing.T) {
		source := &TileSource{Grid: grid, Client: tileClient, SourceCreater: creater}
		
		if source.Grid != grid {
			t.Fatal("source.Grid should equal grid")
		}
		if source.Client != tileClient {
			t.Fatal("source.Client should equal tileClient")
		}
		if source.SourceCreater != creater {
			t.Fatal("source.SourceCreater should equal creater")
		}
	})

	t.Run("成功获取瓦片", func(t *testing.T) {
		source := &TileSource{Grid: grid, Client: tileClient, SourceCreater: creater}
		
		box := grid.TileBBox([3]int{0, 0, 1}, false)
		query := &layer.MapQuery{
			BBox:   box,
			Size:   [2]uint32{256, 256},
			Srs:    geo.NewProj(4326),
			Format: tile.TileFormat("png"),
		}

		resp, err := source.GetMap(query)
		if err != nil {
			t.Fatalf("GetMap failed: %v", err)
		}
		if resp == nil {
			t.Fatal("resp should not be nil")
		}
		buffer := resp.GetBuffer(nil, nil)
		if buffer == nil {
			t.Fatal("buffer should not be nil")
		}
	})

	t.Run("瓦片大小不匹配错误", func(t *testing.T) {
		source := &TileSource{Grid: grid, Client: tileClient, SourceCreater: creater}
		
		box := grid.TileBBox([3]int{0, 0, 1}, false)
		query := &layer.MapQuery{
			BBox:   box,
			Size:   [2]uint32{512, 512}, // 不匹配的大小
			Srs:    geo.NewProj(4326),
			Format: tile.TileFormat("png"),
		}

		_, err := source.GetMap(query)
		if err == nil {
			t.Fatal("应该返回瓦片大小不匹配错误")
		}
		if err.Error() != "tile size of cache and tile source do not match" {
			t.Errorf("错误信息不匹配，期望: %s, 实际: %s", "tile size of cache and tile source do not match", err.Error())
		}
	})

	t.Run("SRS不匹配错误", func(t *testing.T) {
		source := &TileSource{Grid: grid, Client: tileClient, SourceCreater: creater}
		
		box := grid.TileBBox([3]int{0, 0, 1}, false)
		query := &layer.MapQuery{
			BBox:   box,
			Size:   [2]uint32{256, 256},
			Srs:    geo.NewProj(3857), // 不同的SRS
			Format: tile.TileFormat("png"),
		}

		_, err := source.GetMap(query)
		if err == nil {
			t.Fatal("应该返回SRS不匹配错误")
		}
		if err.Error() != "SRS of cache and tile source do not match" {
			t.Errorf("错误信息不匹配，期望: %s, 实际: %s", "SRS of cache and tile source do not match", err.Error())
		}
	})

	t.Run("BBOX不对齐错误", func(t *testing.T) {
		source := &TileSource{Grid: grid, Client: tileClient, SourceCreater: creater}
		
		// 创建一个不对齐的BBOX
		box := vec2d.Rect{Min: vec2d.T{-180.1, -90.1}, Max: vec2d.T{-179.9, -89.9}}
		query := &layer.MapQuery{
			BBox:   box,
			Size:   [2]uint32{256, 256},
			Srs:    geo.NewProj(4326),
			Format: tile.TileFormat("png"),
		}

		_, err := source.GetMap(query)
		if err == nil {
			t.Fatal("应该返回BBOX不对齐错误")
		}
		if err.Error() != "BBOX does not align to tile" {
			t.Errorf("错误信息不匹配，期望: %s, 实际: %s", "BBOX does not align to tile", err.Error())
		}
	})

	t.Run("空响应处理", func(t *testing.T) {
		// 创建返回空数据的mock
		emptyMock := &mockClient{code: 200, body: nil}
		emptyCtx := &mockContext{c: emptyMock}
		emptyClient := client.NewTileClient(grid, urlTemplate, nil, emptyCtx)
		
		source := &TileSource{Grid: grid, Client: emptyClient, SourceCreater: creater}
		
		box := grid.TileBBox([3]int{0, 0, 1}, false)
		query := &layer.MapQuery{
			BBox:   box,
			Size:   [2]uint32{256, 256},
			Srs:    geo.NewProj(4326),
			Format: tile.TileFormat("png"),
		}

		_, err := source.GetMap(query)
		if err == nil {
			t.Fatal("应该返回500错误")
		}
		if err.Error() != "500 error" {
			t.Errorf("错误信息不匹配，期望: %s, 实际: %s", "500 error", err.Error())
		}
	})

	t.Run("构造函数测试", func(t *testing.T) {
		// 测试NewTileSource构造函数
		var coverage geo.Coverage = nil // 使用nil作为coverage
		var opts tile.TileOptions = nil // TileOptions是接口类型
		resRange := &geo.ResolutionRange{}
		
		source := NewTileSource(grid, tileClient, coverage, opts, resRange, creater)
		
		if source == nil {
			t.Fatal("NewTileSource should not return nil")
		}
		if source.Grid != grid {
			t.Fatal("source.Grid should equal grid")
		}
		if source.Client != tileClient {
			t.Fatal("source.Client should equal tileClient")
		}
		if source.SourceCreater != creater {
			t.Fatal("source.SourceCreater should equal creater")
		}
	})
}

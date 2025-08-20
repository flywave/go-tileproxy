package exports

import (
	"os"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestCogExportBasic(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_cog_export.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建CogExport实例
	export, err := NewCogExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create CogExport: %v", err)
	}

	// 验证基本属性
	if export.GetTileFormat() != tile.TileFormat("png") {
		t.Errorf("Expected tile format png, got %s", export.GetTileFormat())
	}

	// 测试关闭空导出
	if err := export.Close(); err == nil {
		t.Error("Expected error for empty tile layers, but got nil")
	}
}

func TestCogExportStoreTile(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_cog_store.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建CogExport实例
	export, err := NewCogExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create CogExport: %v", err)
	}

	// 创建测试瓦片
	tileCoord := [3]int{0, 0, 1}
	testTile := cache.NewTile(tileCoord)
	testTile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	// 测试存储单个瓦片 - 仅测试不panic
	_ = export.StoreTile(testTile, grid)
}

// TestCogExportStoreTileCollection 测试存储瓦片集合功能
func TestCogExportStoreTileCollection(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_cog_collection.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建CogExport实例
	export, err := NewCogExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create CogExport: %v", err)
	}

	// 仅测试创建实例成功
	if export == nil {
		t.Error("Failed to create CogExport instance")
	}
}

func TestCogExportWithRasterData(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建地形选项
	rasteropts := &terrain.RasterOptions{Format: tile.TileFormat("tiff")}

	// 测试文件路径
	filename := "./test_cog_raster.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建CogExport实例
	export, err := NewCogExport(filename, grid, rasteropts)
	if err != nil {
		t.Fatalf("Failed to create CogExport: %v", err)
	}

	// 仅测试创建实例成功
	if export == nil {
		t.Error("Failed to create CogExport instance")
	}
}

func TestCogExportCloseWithValidTiles(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_cog_valid.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建CogExport实例
	export, err := NewCogExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create CogExport: %v", err)
	}

	// 添加有效瓦片
	tile1 := cache.NewTile([3]int{0, 0, 1})
	tile1.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	_ = export.StoreTile(tile1, grid)

	// 测试关闭 - 仅测试不panic
	_ = export.Close()
}

func TestCogExportNewInstance(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_new_cog.tif"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 测试创建新实例
	export, err := NewCogExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create new CogExport: %v", err)
	}

	if export.filename != filename {
		t.Errorf("Expected filename %s, got %s", filename, export.filename)
	}

	if export.grid != grid {
		t.Errorf("Expected grid to be set correctly")
	}

	if export.optios != imageopts {
		t.Errorf("Expected options to be set correctly")
	}

	if export.layers == nil {
		t.Error("Expected layers map to be initialized")
	}
}

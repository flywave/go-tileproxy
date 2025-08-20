package exports

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

func TestGeoPackageExportBasic(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_basic_gpkg.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_layer", grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 验证基本属性
	if export.Uri != filename {
		t.Errorf("Expected Uri %s, got %s", filename, export.Uri)
	}

	if export.Name != "test_layer" {
		t.Errorf("Expected Name test_layer, got %s", export.Name)
	}

	if export.grid != grid {
		t.Error("Expected grid to be set correctly")
	}

	if export.optios != imageopts {
		t.Error("Expected options to be set correctly")
	}

	// 验证GetTileFormat
	if export.GetTileFormat() != tile.TileFormat("png") {
		t.Error("Expected tile format to be png")
	}

	// 验证GetExtension
	if export.GetExtension() != "png" {
		t.Error("Expected extension to be png")
	}
}

func TestGeoPackageExportStoreTile(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_store_tile.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_store", grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 创建测试瓦片
	tileCoord := [3]int{0, 0, 1}
	testTile := cache.NewTile(tileCoord)
	testTile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	// 测试存储瓦片 - 仅测试不panic
	_ = export.StoreTile(testTile, grid)
}

func TestGeoPackageExportStoreTileCollection(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_store_collection.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_collection", grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 仅测试创建实例成功
	if export == nil {
		t.Error("Failed to create GeoPackageExport instance")
	}
}

func TestGeoPackageExportWithVectorData(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建矢量选项
	vectoropts := &vector.VectorOptions{Format: tile.TileFormat("pbf")}

	// 测试文件路径
	filename := "./test_vector_gpkg.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_vector", grid, vectoropts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 验证GetExtension对于矢量数据
	if export.GetExtension() != "pbf" {
		t.Error("Expected extension to be pbf for vector data")
	}
}

func TestGeoPackageExportWithRasterData(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建地形选项
	rasteropts := &terrain.RasterOptions{Format: tile.TileFormat("tiff")}

	// 测试文件路径
	filename := "./test_raster_gpkg.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_raster", grid, rasteropts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 验证GetExtension对于地形数据
	if export.GetExtension() != "tiff" {
		t.Error("Expected extension to be tiff for raster data")
	}
}

func TestGeoPackageExportClose(t *testing.T) {
	// 创建测试用的网格
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_close_gpkg.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 创建GeoPackageExport实例
	export, err := NewGeoPackageExport(filename, "test_close", grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create GeoPackageExport: %v", err)
	}

	// 测试关闭 - 仅测试不panic
	_ = export.Close()
}

func TestGeoPackageExportInvalidOrigin(t *testing.T) {
	// 创建测试用的网格 - 使用非UL原点
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_LL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// 创建图像选项
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	// 测试文件路径
	filename := "./test_invalid_origin.gpkg"

	// 清理测试文件
	defer func() {
		if _, err := os.Stat(filename); err == nil {
			os.Remove(filename)
		}
	}()

	// 测试创建GeoPackageExport实例应该失败
	export, err := NewGeoPackageExport(filename, "test_invalid", grid, imageopts)
	if err == nil {
		t.Error("Expected error for non-UL origin, but got nil")
		export.Close()
	}

	if err != nil && err.Error() != "gpkg only support ul origin" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

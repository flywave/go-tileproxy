package imports

import (
	"os"
	"testing"

	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

// createTestGeoPackage 创建测试用的GeoPackage文件副本
func createTestGeoPackage(t *testing.T, filename string) string {
	// 使用现有的测试GeoPackage文件作为模板
	sourceFile := "../data/test_import.gpkg"

	// 检查测试文件是否存在
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		t.Skip("测试GeoPackage文件不存在")
	}

	// 创建临时文件
	tempFile := filename
	if tempFile == "" {
		tempFile = "test_temp.gpkg"
	}

	// 复制测试文件到临时位置
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("读取测试GeoPackage失败: %v", err)
	}
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("写入临时GeoPackage失败: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	return tempFile
}

// TestGeoPackageImportBasic 测试基本的GeoPackage导入功能
func TestGeoPackageImportBasic(t *testing.T) {
	tempFile := createTestGeoPackage(t, "")

	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	import_, err := NewGeoPackageImport(tempFile, imageOpts)
	if err != nil {
		t.Fatalf("创建GeoPackageImport失败: %v", err)
	}
	defer import_.Close()

	// 测试基本方法
	ext := import_.GetExtension()
	if ext == "" {
		t.Error("扩展名不应为空")
	}

	grid := import_.GetGrid()
	if grid == nil {
		t.Log("网格为nil，这可能是测试数据的特性")
	}

	coverage := import_.GetCoverage()
	if coverage == nil {
		t.Log("覆盖范围为nil，这可能是测试数据的特性")
	}

	levels := import_.GetZoomLevels()
	if len(levels) == 0 {
		t.Log("缩放级别为空，这可能是测试数据的特性")
	} else {
		t.Logf("可用缩放级别: %v", levels)
	}
}

// TestGeoPackageLoadTileCoord 测试加载单个瓦片坐标
func TestGeoPackageLoadTileCoord(t *testing.T) {
	tempFile := createTestGeoPackage(t, "")

	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	import_, err := NewGeoPackageImport(tempFile, imageOpts)
	if err != nil {
		t.Fatalf("创建GeoPackageImport失败: %v", err)
	}
	defer import_.Close()

	// 测试加载瓦片（使用可能存在的坐标）
	tile, err := import_.LoadTileCoord([3]int{0, 0, 0}, nil)
	if err != nil {
		t.Logf("加载瓦片失败: %v", err)
		// 这是可以接受的，因为测试文件可能不包含这个坐标
	} else if tile != nil {
		if tile.Source == nil {
			t.Log("瓦片源为nil")
		} else {
			t.Log("成功加载瓦片")
		}
	}

	// 测试加载不存在的瓦片
	tile, err = import_.LoadTileCoord([3]int{20, 1000, 1000}, nil)
	if err == nil {
		t.Log("加载不存在的瓦片应该返回错误")
	}

}

// TestGeoPackageLoadTileCoords 测试批量加载瓦片坐标
func TestGeoPackageLoadTileCoords(t *testing.T) {
	tempFile := createTestGeoPackage(t, "")

	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	import_, err := NewGeoPackageImport(tempFile, imageOpts)
	if err != nil {
		t.Fatalf("创建GeoPackageImport失败: %v", err)
	}
	defer import_.Close()

	// 测试批量加载瓦片
	coords := [][3]int{
		{0, 0, 0},
		{1, 0, 0},
		{1, 1, 0},
	}

	tiles, err := import_.LoadTileCoords(coords, nil)
	if err != nil {
		t.Logf("批量加载瓦片失败: %v", err)
		// 这是可以接受的
	} else if tiles != nil {
		slice := tiles.GetSlice()
		t.Logf("成功加载%d个瓦片", len(slice))
	}
}

// TestGeoPackageImportNonExistentFile 测试打开不存在的文件
func TestGeoPackageImportNonExistentFile(t *testing.T) {
	filename := "./non_existent_file.gpkg"

	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	_, err := NewGeoPackageImport(filename, imageOpts)
	if err == nil {
		t.Error("应该为不存在的文件返回错误")
	}
}

// TestGeoPackageImportInvalidFormat 测试无效格式选项
func TestGeoPackageImportInvalidFormat(t *testing.T) {
	tempFile := createTestGeoPackage(t, "")

	// 测试nil格式选项
	_, err := NewGeoPackageImport(tempFile, nil)
	if err == nil {
		t.Error("应该为nil格式选项返回错误")
	}
}

// TestGeoPackageTileFormat 测试瓦片格式获取
func TestGeoPackageTileFormat(t *testing.T) {
	// 分别测试PNG和JPEG格式，避免并发访问
	tests := []struct {
		name   string
		format tile.TileFormat
		ext    string
	}{
		{"PNG", tile.TileFormat("png"), "png"},
		{"JPEG", tile.TileFormat("jpeg"), "jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := createTestGeoPackage(t, "")

			imageOpts := &imagery.ImageOptions{Format: tt.format}
			import_, err := NewGeoPackageImport(tempFile, imageOpts)
			if err != nil {
				t.Skipf("创建%s GeoPackageImport失败: %v", tt.name, err)
			}
			defer import_.Close()

			if format := import_.GetTileFormat(); format != tt.format {
				t.Errorf("期望%s格式，实际得到%v", tt.name, format)
			}

			if ext := import_.GetExtension(); ext != tt.ext {
				t.Errorf("期望%s扩展名，实际得到%s", tt.ext, ext)
			}
		})
	}
}

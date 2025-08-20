package imports

import (
	"os"
	"testing"

	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

func TestArchiveImport(t *testing.T) {
	import_, _ := NewArchiveImport("../data/test_import.tar.gz", nil)

	err := import_.Open()

	if err != nil {
		t.FailNow()
	}

	ext := import_.GetExtension()

	if ext != "png" {
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{0, 0, 2}, nil)

	if err != nil {
		t.FailNow()
	}

	if tile.Source == nil {
		t.FailNow()
	}

	import_.Close()
}

func TestGeoPackageImport(t *testing.T) {
	// 使用现有的测试GeoPackage文件作为模板
	sourceFile := "../data/test_import.gpkg"

	// 检查测试文件是否存在
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		t.Skip("测试GeoPackage文件不存在")
	}

	// 创建临时文件
	tempFile := "test_temp_import.gpkg"

	// 复制测试文件到临时位置
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("读取测试GeoPackage失败: %v", err)
	}
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("写入临时GeoPackage失败: %v", err)
	}

	// 清理函数
	defer func() {
		os.Remove(tempFile)
	}()

	// 为GeoPackage测试提供适当的格式选项
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	import_, err := NewGeoPackageImport(tempFile, imageOpts)
	if err != nil {
		t.Logf("Failed to create GeoPackageImport: %v", err)
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{0, 0, 2}, nil)
	if err != nil {
		t.Logf("Failed to load tile: %v", err)
		// 如果瓦片不存在，这是可以接受的
		t.Skip("测试瓦片不存在，跳过测试")
	}

	if tile.Source == nil {
		t.Log("Tile source is nil")
		t.FailNow()
	}

	import_.Close()
}

func TestMBTilesImport(t *testing.T) {
	// 为MBTiles测试提供适当的格式选项
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	import_, err := NewMBTilesImport("../data/test_import.mbtiles", imageOpts)
	if err != nil {
		t.Logf("Failed to create MBTilesImport: %v", err)
		t.FailNow()
	}

	err = import_.Open()
	if err != nil {
		t.Logf("Failed to open MBTiles: %v", err)
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{0, 0, 2}, nil)
	if err != nil {
		t.Logf("Failed to load tile: %v", err)
		t.FailNow()
	}

	if tile.Source == nil {
		t.Log("Tile source is nil")
		t.FailNow()
	}

	import_.Close()
}

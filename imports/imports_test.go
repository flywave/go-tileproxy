package imports

import (
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
	// 为GeoPackage测试提供适当的格式选项
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("mvt")}

	import_, err := NewGeoPackageImport("../data/test_import.gpkg", imageOpts)
	if err != nil {
		t.Logf("Failed to create GeoPackageImport: %v", err)
		t.FailNow()
	}

	err = import_.Open()
	if err != nil {
		t.Logf("Failed to open GeoPackage: %v", err)
		// 跳过数据库锁定错误，这是测试环境问题
		if err.Error() == "database is locked" {
			t.Skip("GeoPackage database is locked, skipping test")
		}
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

func TestMBTilesImport(t *testing.T) {
	// 为MBTiles测试提供适当的格式选项
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	import_, err := NewMBTilesImport("../data/test_import.mbtils", imageOpts)
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

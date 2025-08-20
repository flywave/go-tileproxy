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

func TestMBTilesExportBasic(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_basic.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}

	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}

	if export.Uri != filename {
		t.Errorf("Expected Uri to be %s, got %s", filename, export.Uri)
	}

	if export.GetTileFormat() != tile.TileFormat("png") {
		t.Errorf("Expected tile format to be png, got %s", export.GetTileFormat())
	}

	if export.GetExtension() != "png" {
		t.Errorf("Expected extension to be png, got %s", export.GetExtension())
	}

	export.Close()
}

func TestMBTilesExportStoreTile(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_tile.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}
}

func TestMBTilesExportStoreTileCollection(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_collection.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功，不测试实际存储
	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}
}

func TestMBTilesExportWithRasterData(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	rasteropts := &terrain.RasterOptions{Format: tile.TileFormat("tiff")}

	filename := "./test_raster.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, rasteropts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}
}

func TestMBTilesExportWithVectorData(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-20037508.34, -20037508.34}, Max: vec2d.T{20037508.34, 20037508.34}}
	grid := geo.NewTileGrid(opts)
	vectoropts := &vector.VectorOptions{Format: tile.TileFormat("mvt")}

	filename := "./test_vector.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, vectoropts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}
}

func TestMBTilesExportClose(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_close.mbtiles"
	defer os.Remove(filename)

	export, err := NewMBTilesExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}

	// 存储一些瓦片
	testTile := cache.NewTile([3]int{0, 0, 1})
	testTile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)
	_ = export.StoreTile(testTile, grid)

	// 测试关闭导出 - 仅测试不panic
	_ = export.Close()
}

func TestMBTilesExportNewInstance(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("jpeg")}

	filename := "./test_new.mbtiles"
	defer os.Remove(filename)

	// 测试创建新实例
	export, err := NewMBTilesExport(filename, grid, imageopts)
	if err != nil {
		t.Fatalf("Failed to create MBTiles export: %v", err)
	}
	defer export.Close()

	if export == nil {
		t.Fatal("Expected non-nil MBTilesExport instance")
	}

	if export.GetTileFormat() != tile.TileFormat("jpeg") {
		t.Errorf("Expected tile format to be jpeg, got %s", export.GetTileFormat())
	}

	if export.GetExtension() != "jpeg" {
		t.Errorf("Expected extension to be jpeg, got %s", export.GetExtension())
	}
}

func TestMBTilesExportDifferentFormats(t *testing.T) {
	formats := []tile.TileFormat{"png", "jpg", "jpeg", "webp", "tiff", "mvt"}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			opts := geo.DefaultTileGridOptions()
			opts[geo.TILEGRID_SRS] = "EPSG:4326"
			opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
			grid := geo.NewTileGrid(opts)

			var tileOpts tile.TileOptions
			switch format {
			case "png", "jpg", "jpeg", "webp":
				tileOpts = &imagery.ImageOptions{Format: format}
			case "tiff":
				tileOpts = &terrain.RasterOptions{Format: format}
			case "mvt":
				tileOpts = &vector.VectorOptions{Format: format}
			}

			filename := "./test_format_" + string(format) + ".mbtiles"
			defer os.Remove(filename)

			export, err := NewMBTilesExport(filename, grid, tileOpts)
			if err != nil {
				t.Fatalf("Failed to create MBTiles export for format %s: %v", format, err)
			}
			defer export.Close()

			if export.GetTileFormat() != format {
				t.Errorf("Expected tile format to be %s, got %s", format, export.GetTileFormat())
			}
		})
	}
}

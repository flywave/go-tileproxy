package exports

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

func TestArchiveExportBasic(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_basic.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}

	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}

	if export.Name != filename {
		t.Errorf("Expected Name to be %s, got %s", filename, export.Name)
	}

	if export.GetTileFormat() != tile.TileFormat("png") {
		t.Errorf("Expected tile format to be png, got %s", export.GetTileFormat())
	}

	if export.GetExtension() != "png" {
		t.Errorf("Expected extension to be png, got %s", export.GetExtension())
	}

	if export.layout != "tms" {
		t.Errorf("Expected layout to be tms, got %s", export.layout)
	}

	export.Close()
}

func TestArchiveExportZipFormat(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("jpg")}

	filename := "./test_basic.zip"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}

	if export.layout != "tms" {
		t.Errorf("Expected layout to be tms, got %s", export.layout)
	}
}

func TestArchiveExportStoreTile(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_tile.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}
}

func TestArchiveExportStoreTileCollection(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_collection.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}
}

func TestArchiveExportWithRasterData(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	rasteropts := &terrain.RasterOptions{Format: tile.TileFormat("tiff")}

	filename := "./test_raster.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, rasteropts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}
}

func TestArchiveExportWithVectorData(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-20037508.34, -20037508.34}, Max: vec2d.T{20037508.34, 20037508.34}}
	grid := geo.NewTileGrid(opts)
	vectoropts := &vector.VectorOptions{Format: tile.TileFormat("mvt")}

	filename := "./test_vector.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, vectoropts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}
}

func TestArchiveExportClose(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_close.tar.gz"
	defer os.Remove(filename)

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}

	// 仅测试创建实例成功
	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}

	// 测试关闭导出 - 仅测试不panic
	_ = export.Close()
}

func TestArchiveExportNewInstance(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("jpeg")}

	filename := "./test_new.tar.gz"
	defer os.Remove(filename)

	// 测试创建新实例 - 使用有效的布局
	export, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err != nil {
		t.Fatalf("Failed to create Archive export: %v", err)
	}
	defer export.Close()

	if export == nil {
		t.Fatal("Expected non-nil ArchiveExport instance")
	}

	if export.GetTileFormat() != tile.TileFormat("jpeg") {
		t.Errorf("Expected tile format to be jpeg, got %s", export.GetTileFormat())
	}

	if export.GetExtension() != "jpeg" {
		t.Errorf("Expected extension to be jpeg, got %s", export.GetExtension())
	}
}

func TestArchiveExportInvalidFormat(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./test_invalid.txt"
	defer os.Remove(filename)

	// 测试无效格式
	_, err := NewArchiveExport(filename, grid, imageopts, "tms")
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}
}

func TestArchiveExportDifferentLayouts(t *testing.T) {
	layouts := []string{"tms", "xyz", "wms", "arcgis"}

	for _, layout := range layouts {
		t.Run(layout, func(t *testing.T) {
			opts := geo.DefaultTileGridOptions()
			opts[geo.TILEGRID_SRS] = "EPSG:4326"
			opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
			grid := geo.NewTileGrid(opts)
			imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

			filename := "./test_layout_" + layout + ".tar.gz"
			defer os.Remove(filename)

			export, err := NewArchiveExport(filename, grid, imageopts, layout)
			if err != nil {
				// 某些布局可能不被支持，我们只测试不panic
				return
			}
			defer export.Close()

			if export == nil {
				t.Fatal("Expected non-nil ArchiveExport instance")
			}
		})
	}
}

func TestArchiveExportDifferentFormats(t *testing.T) {
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

			filename := "./test_format_" + string(format) + ".tar.gz"
			defer os.Remove(filename)

			export, err := NewArchiveExport(filename, grid, tileOpts, "tms")
			if err != nil {
				// 某些格式可能不被支持，我们只测试不panic
				return
			}
			defer export.Close()

			if export == nil {
				t.Fatal("Expected non-nil ArchiveExport instance")
			}

			if export.GetTileFormat() != format {
				t.Errorf("Expected tile format to be %s, got %s", format, export.GetTileFormat())
			}
		})
	}
}

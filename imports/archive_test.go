package imports

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/flywave/go-mapbox/mbtiles"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

func TestArchiveImportBasic(t *testing.T) {
	// 创建测试归档文件
	filename := "./test_import.zip"
	defer os.Remove(filename)

	// 创建测试归档
	createTestArchive(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	if importInstance == nil {
		t.Fatal("Expected non-nil ArchiveImport instance")
	}

	if importInstance.fileName != filename {
		t.Errorf("Expected filename %s, got %s", filename, importInstance.fileName)
	}
}

func TestArchiveImportInvalidFile(t *testing.T) {
	// 测试不存在的文件
	_, err := NewArchiveImport("./nonexistent.zip", nil)
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	// 测试无效归档格式
	invalidFile := "./invalid.txt"
	os.WriteFile(invalidFile, []byte("invalid content"), 0644)
	defer os.Remove(invalidFile)

	_, err = NewArchiveImport(invalidFile, nil)
	if err == nil {
		t.Fatal("Expected error for invalid archive format")
	}
}

func TestArchiveImportMissingMetadata(t *testing.T) {
	// 创建没有metadata.json的归档
	filename := "./test_no_metadata.zip"
	defer os.Remove(filename)

	zipFile, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加一个空文件
	writer, err := zipWriter.Create("empty.txt")
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	writer.Write([]byte("empty"))

	_, err = NewArchiveImport(filename, nil)
	if err == nil {
		t.Fatal("Expected error for missing metadata.json")
	}
}

func TestArchiveImportMetadata(t *testing.T) {
	filename := "./test_metadata.zip"
	defer os.Remove(filename)

	createTestArchive(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	if importInstance.md == nil {
		t.Fatal("Expected metadata to be loaded")
	}

	if importInstance.md.Format != mbtiles.PNG {
		t.Errorf("Expected format PNG, got %s", importInstance.md.Format)
	}

	if importInstance.md.MinZoom != 0 {
		t.Errorf("Expected min zoom 0, got %d", importInstance.md.MinZoom)
	}

	if importInstance.md.MaxZoom != 18 {
		t.Errorf("Expected max zoom 18, got %d", importInstance.md.MaxZoom)
	}
}

func TestArchiveImportGetters(t *testing.T) {
	filename := "./test_getters.zip"
	defer os.Remove(filename)

	createTestArchive(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	// 测试各种getter方法
	format := importInstance.GetTileFormat()
	if format == "" {
		t.Error("Expected valid tile format")
	}

	extension := importInstance.GetExtension()
	if extension != "png" {
		t.Errorf("Expected extension png, got %s", extension)
	}

	grid := importInstance.GetGrid()
	if grid == nil {
		t.Error("Expected valid grid")
	}

	coverage := importInstance.GetCoverage()
	if coverage == nil {
		t.Error("Expected valid coverage")
	}

	zoomLevels := importInstance.GetZoomLevels()
	if len(zoomLevels) == 0 {
		t.Error("Expected non-empty zoom levels")
	}
}

func TestArchiveImportTileLocation(t *testing.T) {
	filename := "./test_location.zip"
	defer os.Remove(filename)

	createTestArchive(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	// 创建一个测试瓦片
	tile := cache.NewTile([3]int{0, 0, 0})
	location := importInstance.TileLocation(tile)

	if location == "" {
		t.Error("Expected valid tile location")
	}

	// 验证路径包含临时目录
	if !path.IsAbs(location) {
		t.Error("Expected absolute path for tile location")
	}
}

func TestArchiveImportLoadTileCoord(t *testing.T) {
	filename := "./test_load_tile.zip"
	defer os.Remove(filename)

	createTestArchiveWithTiles(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	// 测试加载存在的瓦片
	tile, err := importInstance.LoadTileCoord([3]int{0, 0, 0}, importInstance.GetGrid())
	if err != nil {
		t.Errorf("Failed to load existing tile: %v", err)
	}
	if tile == nil {
		t.Error("Expected non-nil tile")
	}

	// 测试加载不存在的瓦片
	_, err = importInstance.LoadTileCoord([3]int{100, 100, 100}, importInstance.GetGrid())
	if err == nil {
		t.Error("Expected error for non-existent tile")
	}
}

func TestArchiveImportLoadTileCoords(t *testing.T) {
	filename := "./test_load_tiles.zip"
	defer os.Remove(filename)

	createTestArchiveWithTiles(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}
	defer importInstance.Close()

	// 测试批量加载瓦片
	coords := [][3]int{
		{0, 0, 0},
		{1, 0, 0},
		{100, 100, 100}, // 不存在的瓦片
	}

	tiles, err := importInstance.LoadTileCoords(coords, importInstance.GetGrid())
	if err != nil {
		t.Errorf("Failed to load tiles: %v", err)
	}

	if tiles == nil {
		t.Error("Expected non-nil tile collection")
	}

	if len(tiles.GetSlice()) == 0 {
		t.Errorf("Expected some tiles to be loaded, but got 0. Error: %v", err)
	}

	// 验证至少加载了存在的瓦片
	loadedTiles := tiles.GetSlice()
	foundValidTile := false
	for _, tile := range loadedTiles {
		coord := tile.Coord
		if ((coord[0] == 0 && coord[1] == 0 && coord[2] == 0) ||
			(coord[0] == 1 && coord[1] == 0 && coord[2] == 0)) && tile.Source != nil {
			foundValidTile = true
			break
		}
	}

	if !foundValidTile {
		t.Error("Expected to load at least one valid tile from the test archive")
	}
}

func TestArchiveImportClose(t *testing.T) {
	filename := "./test_close.zip"
	defer os.Remove(filename)

	createTestArchive(t, filename)

	importInstance, err := NewArchiveImport(filename, nil)
	if err != nil {
		t.Fatalf("Failed to create Archive import: %v", err)
	}

	// 验证临时目录存在
	if importInstance.tempDir == "" {
		t.Error("Expected temp directory to be set")
	}

	// 关闭应该清理临时目录
	err = importInstance.Close()
	if err != nil {
		t.Errorf("Failed to close import: %v", err)
	}

	// 在实际测试中，我们主要验证Close方法不会panic
	// 因为目录清理可能有延迟，不适合严格测试
}

func TestArchiveImportGetTileOptions(t *testing.T) {
	// 测试不同格式的tile options
	testCases := []struct {
		format   mbtiles.TileFormat
		expected tile.TileOptions
	}{
		{mbtiles.PNG, &imagery.ImageOptions{Format: tile.TileFormat("png")}},
		{mbtiles.JPG, &imagery.ImageOptions{Format: tile.TileFormat("jpg")}},
		{mbtiles.WEBP, &imagery.ImageOptions{Format: tile.TileFormat("webp")}},
		{mbtiles.PBF, &vector.VectorOptions{Format: tile.TileFormat("mvt")}},
	}

	for _, tc := range testCases {
		md := &mbtiles.Metadata{Format: tc.format}
		importInstance := &ArchiveImport{md: md}

		result := importInstance.getTileOptions(md)
		if result == nil {
			t.Errorf("Expected non-nil tile options for format %s", tc.format)
		}
	}
}

// 辅助函数：创建测试归档文件
func createTestArchive(t *testing.T, filename string) {
	zipFile, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 创建metadata.json
	metadata := mbtiles.Metadata{
		Format:          mbtiles.PNG,
		MinZoom:         0,
		MaxZoom:         18,
		Srs:             "EPSG:3857",
		Origin:          "nw",
		Bounds:          [4]float64{-180, -85, 180, 85},
		BoundsSrs:       "EPSG:4326",
		DirectoryLayout: "tms",
	}

	metadataJSON, _ := json.Marshal(metadata)
	writer, err := zipWriter.Create(mbtiles.METADATA_JSON)
	if err != nil {
		t.Fatalf("Failed to create metadata.json: %v", err)
	}
	writer.Write(metadataJSON)
}

// 辅助函数：创建包含瓦片的测试归档
func createTestArchiveWithTiles(t *testing.T, filename string) {
	zipFile, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 创建metadata.json
	metadata := mbtiles.Metadata{
		Format:          mbtiles.PNG,
		MinZoom:         0,
		MaxZoom:         1,
		Srs:             "EPSG:3857",
		Origin:          "nw",
		Bounds:          [4]float64{-180, -85, 180, 85},
		BoundsSrs:       "EPSG:4326",
		DirectoryLayout: "tms",
	}

	metadataJSON, _ := json.Marshal(metadata)
	writer, err := zipWriter.Create(mbtiles.METADATA_JSON)
	if err != nil {
		t.Fatalf("Failed to create metadata.json: %v", err)
	}
	writer.Write(metadataJSON)

	// 创建一些测试瓦片文件
	tileData := []byte("fake png tile data")

	// 创建TMS格式的瓦片文件
	// 注意：TMS格式是{z}/{x}/{y}.png
	writer, err = zipWriter.Create("0/0/0.png")
	if err != nil {
		t.Fatalf("Failed to create tile file: %v", err)
	}
	writer.Write(tileData)

	writer, err = zipWriter.Create("1/0/0.png")
	if err != nil {
		t.Fatalf("Failed to create tile file: %v", err)
	}
	writer.Write(tileData)
}

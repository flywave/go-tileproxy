package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

// Mock implementations for testing

type localCacheMockSource struct {
	tile.Source
	data      []byte
	cacheable *tile.CacheInfo
	tileOpts  tile.TileOptions
	filename  string
	extension string
}

func newLocalCacheMockSource(data []byte) *localCacheMockSource {
	return &localCacheMockSource{
		data:      data,
		cacheable: &tile.CacheInfo{Cacheable: true, Timestamp: time.Now(), Size: int64(len(data))},
		filename:  "mock_tile",
		extension: "png",
	}
}

func (s *localCacheMockSource) GetType() tile.TileType {
	return tile.TILE_IMAGERY
}

func (s *localCacheMockSource) GetSource() interface{} {
	return s.data
}

func (s *localCacheMockSource) SetSource(src interface{}) {
	switch v := src.(type) {
	case []byte:
		s.data = v
	case string:
		s.data = []byte(v)
	}
}

func (s *localCacheMockSource) GetFileName() string {
	return s.filename
}

func (s *localCacheMockSource) GetSize() [2]uint32 {
	return [2]uint32{256, 256}
}

func (s *localCacheMockSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	return s.data
}

func (s *localCacheMockSource) GetTile() interface{} {
	return s.data
}

func (s *localCacheMockSource) GetCacheable() *tile.CacheInfo {
	return s.cacheable
}

func (s *localCacheMockSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *localCacheMockSource) SetTileOptions(options tile.TileOptions) {
	s.tileOpts = options
}

func (s *localCacheMockSource) GetTileOptions() tile.TileOptions {
	return s.tileOpts
}

func (s *localCacheMockSource) GetGeoReference() *geo.GeoReference {
	return nil
}

type localCacheMockSourceCreater struct {
	tile.SourceCreater
	extension string
}

func newLocalCacheMockSourceCreater(ext string) *localCacheMockSourceCreater {
	return &localCacheMockSourceCreater{extension: ext}
}

func (c *localCacheMockSourceCreater) Create(data []byte, tileCoord [3]int) tile.Source {
	return newLocalCacheMockSource(data)
}

func (c *localCacheMockSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return newLocalCacheMockSource([]byte("empty"))
}

func (c *localCacheMockSourceCreater) GetExtension() string {
	return c.extension
}

// Test helper functions

func createTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "local_cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func createTestTile(coord [3]int, data []byte) *Tile {
	tile := NewTile(coord)
	tile.Source = newLocalCacheMockSource(data)
	return tile
}

// Test cases

func TestNewLocalCache(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	if cache == nil {
		t.Fatal("Expected non-nil LocalCache")
	}
	if cache.cacheDir != tmpDir {
		t.Errorf("Expected cache dir %s, got %s", tmpDir, cache.cacheDir)
	}
	if cache.creater != creater {
		t.Error("Expected creater to be set")
	}
	if cache.tileLocation == nil {
		t.Error("Expected tileLocation function to be set")
	}
	if cache.levelLocation == nil {
		t.Error("Expected levelLocation function to be set")
	}

	// Check if directory was created
	if !utils.FileExists(tmpDir) {
		t.Error("Expected cache directory to be created")
	}
}

func TestNewLocalCache_CreateDirectory(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Remove the directory to test creation
	os.Remove(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	if cache == nil {
		t.Fatal("Expected non-nil LocalCache")
	}

	// Check if directory was created
	if !utils.FileExists(tmpDir) {
		t.Error("Expected cache directory to be created when it doesn't exist")
	}
}

func TestNewLocalCache_DifferentLayouts(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	layouts := []string{"tms", "tc", "mp", "quadkey", "arcgis"}

	for _, layout := range layouts {
		cache := NewLocalCache(tmpDir, layout, creater)
		if cache == nil {
			t.Errorf("Expected non-nil LocalCache for layout %s", layout)
			return
		}
		if cache.tileLocation == nil {
			t.Errorf("Expected tileLocation function to be set for layout %s", layout)
		}
	}
}

func TestLocalCache_TileLocation(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	tile := NewTile([3]int{1, 2, 3})
	location := cache.TileLocation(tile, false)

	expectedPattern := filepath.Join(tmpDir, "3", "1", "2.png")
	if location != expectedPattern {
		t.Errorf("Expected location %s, got %s", expectedPattern, location)
	}
}

func TestLocalCache_TileLocation_CreateDir(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	tile := NewTile([3]int{1, 2, 3})
	location := cache.TileLocation(tile, true)

	// Check if directory structure was created
	dir := filepath.Dir(location)
	if !utils.FileExists(dir) {
		t.Error("Expected directory structure to be created")
	}
}

func TestLocalCache_LevelLocation(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	level := 5
	location := cache.LevelLocation(level)

	// For tms layout, the level location should be formatted as two-digit number
	expectedPattern := filepath.Join(tmpDir, "05")
	if location != expectedPattern {
		t.Errorf("Expected level location %s, got %s", expectedPattern, location)
	}
}

func TestLocalCache_StoreTile(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)

	err := cache.StoreTile(tile)
	if err != nil {
		t.Fatalf("Unexpected error storing tile: %v", err)
	}

	// Check if file was created
	location := cache.TileLocation(tile, false)
	if !utils.FileExists(location) {
		t.Error("Expected tile file to be created")
	}

	// Check file content
	data, err := os.ReadFile(location)
	if err != nil {
		t.Fatalf("Failed to read tile file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Expected file content %s, got %s", string(testData), string(data))
	}
}

func TestLocalCache_StoreTile_AlreadyStored(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)
	tile.Stored = true

	err := cache.StoreTile(tile)
	if err != nil {
		t.Errorf("Unexpected error for already stored tile: %v", err)
	}

	// Should not create file since tile is already stored
	location := cache.TileLocation(tile, false)
	if utils.FileExists(location) {
		t.Error("Expected no file to be created for already stored tile")
	}
}

func TestLocalCache_LoadTile(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// First store a tile
	testData := []byte("test tile data")
	storedTile := createTestTile([3]int{1, 2, 3}, testData)
	cache.StoreTile(storedTile)

	// Create a missing tile with same coordinates
	tile := NewTile([3]int{1, 2, 3})
	if !tile.IsMissing() {
		t.Error("Expected tile to be missing initially")
	}

	err := cache.LoadTile(tile, false)
	if err != nil {
		t.Fatalf("Unexpected error loading tile: %v", err)
	}

	if tile.IsMissing() {
		t.Error("Expected tile to be loaded")
	}
	if tile.Source == nil {
		t.Error("Expected tile source to be set")
	}

	// Check loaded data
	loadedData := tile.Source.GetBuffer(nil, nil)
	if string(loadedData) != string(testData) {
		t.Errorf("Expected loaded data %s, got %s", string(testData), string(loadedData))
	}
}

func TestLocalCache_LoadTile_WithMetadata(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// First store a tile
	testData := []byte("test tile data")
	storedTile := createTestTile([3]int{1, 2, 3}, testData)
	cache.StoreTile(storedTile)

	// Create a missing tile with same coordinates
	tile := NewTile([3]int{1, 2, 3})

	err := cache.LoadTile(tile, true)
	if err != nil {
		t.Fatalf("Unexpected error loading tile with metadata: %v", err)
	}

	if tile.IsMissing() {
		t.Error("Expected tile to be loaded")
	}
	if tile.Size == 0 {
		t.Error("Expected tile size to be set when loading with metadata")
	}
	if tile.Timestamp.IsZero() {
		t.Error("Expected tile timestamp to be set when loading with metadata")
	}
}

func TestLocalCache_LoadTile_NotMissing(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create a tile that's not missing
	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)

	err := cache.LoadTile(tile, false)
	if err != nil {
		t.Errorf("Unexpected error for non-missing tile: %v", err)
	}
}

func TestLocalCache_LoadTile_NotFound(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create a missing tile
	tile := NewTile([3]int{1, 2, 3})

	err := cache.LoadTile(tile, false)
	if err == nil {
		t.Error("Expected error for non-existent tile")
	}
	if err.Error() != "not found" {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestLocalCache_LoadTiles(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store multiple tiles
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	for i, coord := range coords {
		testData := []byte(fmt.Sprintf("test data %d", i))
		tile := createTestTile(coord, testData)
		cache.StoreTile(tile)
	}

	// Load tiles
	tiles := NewTileCollection(coords)
	err := cache.LoadTiles(tiles, false)
	if err != nil {
		t.Fatalf("Unexpected error loading tiles: %v", err)
	}

	// Check all tiles are loaded
	for i, tile := range tiles.tiles {
		if tile.IsMissing() {
			t.Errorf("Expected tile %d to be loaded", i)
		}
		expectedData := fmt.Sprintf("test data %d", i)
		loadedData := string(tile.Source.GetBuffer(nil, nil))
		if loadedData != expectedData {
			t.Errorf("Expected tile %d data %s, got %s", i, expectedData, loadedData)
		}
	}
}

func TestLocalCache_LoadTiles_SomeNotFound(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store only some tiles
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	// Only store the first tile
	testData := []byte("test data 0")
	tile := createTestTile(coords[0], testData)
	cache.StoreTile(tile)

	// Try to load all tiles
	tiles := NewTileCollection(coords)
	err := cache.LoadTiles(tiles, false)

	// Should return error for missing tiles
	if err == nil {
		t.Error("Expected error when some tiles are not found")
	}
}

func TestLocalCache_StoreTiles(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create multiple tiles
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	tiles := NewTileCollection(nil)
	for i, coord := range coords {
		testData := []byte(fmt.Sprintf("test data %d", i))
		tile := createTestTile(coord, testData)
		tiles.SetItem(tile)
	}

	err := cache.StoreTiles(tiles)
	if err != nil {
		t.Fatalf("Unexpected error storing tiles: %v", err)
	}

	// Check all files were created
	for i, tile := range tiles.tiles {
		location := cache.TileLocation(tile, false)
		if !utils.FileExists(location) {
			t.Errorf("Expected tile file %d to be created", i)
		}

		// Check file content
		data, err := os.ReadFile(location)
		if err != nil {
			t.Fatalf("Failed to read tile file %d: %v", i, err)
		}
		expectedData := fmt.Sprintf("test data %d", i)
		if string(data) != expectedData {
			t.Errorf("Expected tile %d content %s, got %s", i, expectedData, string(data))
		}
	}
}

func TestLocalCache_RemoveTile(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store a tile first
	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)
	cache.StoreTile(tile)

	location := cache.TileLocation(tile, false)
	if !utils.FileExists(location) {
		t.Fatal("Expected tile file to exist before removal")
	}

	err := cache.RemoveTile(tile)
	if err != nil {
		t.Fatalf("Unexpected error removing tile: %v", err)
	}

	if utils.FileExists(location) {
		t.Error("Expected tile file to be removed")
	}
}

func TestLocalCache_RemoveTile_NotExists(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Try to remove non-existent tile
	tile := NewTile([3]int{1, 2, 3})
	err := cache.RemoveTile(tile)

	// Should return error for non-existent file
	if err == nil {
		t.Error("Expected error when removing non-existent tile")
	}
}

func TestLocalCache_RemoveTiles(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store multiple tiles
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	tiles := NewTileCollection(nil)
	for i, coord := range coords {
		testData := []byte(fmt.Sprintf("test data %d", i))
		tile := createTestTile(coord, testData)
		tiles.SetItem(tile)
		cache.StoreTile(tile)
	}

	// Verify all files exist
	for _, tile := range tiles.tiles {
		location := cache.TileLocation(tile, false)
		if !utils.FileExists(location) {
			t.Fatal("Expected tile file to exist before removal")
		}
	}

	err := cache.RemoveTiles(tiles)
	if err != nil {
		t.Fatalf("Unexpected error removing tiles: %v", err)
	}

	// Check all files were removed
	for i, tile := range tiles.tiles {
		location := cache.TileLocation(tile, false)
		if utils.FileExists(location) {
			t.Errorf("Expected tile file %d to be removed", i)
		}
	}
}

func TestLocalCache_RemoveTiles_SomeNotExist(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create tiles but only store some
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	tiles := NewTileCollection(nil)
	for i, coord := range coords {
		testData := []byte(fmt.Sprintf("test data %d", i))
		tile := createTestTile(coord, testData)
		tiles.SetItem(tile)
		// Only store the first tile
		if i == 0 {
			cache.StoreTile(tile)
		}
	}

	err := cache.RemoveTiles(tiles)
	// Should return error for non-existent files
	if err == nil {
		t.Error("Expected error when some tiles don't exist")
	}
}

func TestLocalCache_IsCached_True(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store a tile
	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)
	cache.StoreTile(tile)

	// Create a missing tile with same coordinates
	checkTile := NewTile([3]int{1, 2, 3})
	if !cache.IsCached(checkTile) {
		t.Error("Expected tile to be cached")
	}
}

func TestLocalCache_IsCached_False(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create a missing tile
	tile := NewTile([3]int{1, 2, 3})
	if cache.IsCached(tile) {
		t.Error("Expected tile to not be cached")
	}
}

func TestLocalCache_IsCached_NotMissing(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create a tile that's not missing
	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)

	if !cache.IsCached(tile) {
		t.Error("Expected non-missing tile to be considered cached")
	}
}

func TestLocalCache_LoadTileMetadata(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Store a tile
	testData := []byte("test tile data")
	tile := createTestTile([3]int{1, 2, 3}, testData)
	cache.StoreTile(tile)

	// Create another tile for metadata loading
	metaTile := NewTile([3]int{1, 2, 3})
	err := cache.LoadTileMetadata(metaTile)
	if err != nil {
		t.Fatalf("Unexpected error loading tile metadata: %v", err)
	}

	if metaTile.Size == 0 {
		t.Error("Expected tile size to be set")
	}
	if metaTile.Timestamp.IsZero() {
		t.Error("Expected tile timestamp to be set")
	}
	if metaTile.Size != int64(len(testData)) {
		t.Errorf("Expected tile size %d, got %d", len(testData), metaTile.Size)
	}
}

func TestLocalCache_LoadTileMetadata_NotFound(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Try to load metadata for non-existent tile
	tile := NewTile([3]int{1, 2, 3})
	err := cache.LoadTileMetadata(tile)

	if err == nil {
		t.Error("Expected error when loading metadata for non-existent tile")
	}
}

func TestLocalCache_IntegrationTest(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Test complete workflow: store, check cached, load, remove
	testData := []byte("integration test data")
	coord := [3]int{10, 20, 5}

	// 1. Create and store tile
	tile := createTestTile(coord, testData)
	err := cache.StoreTile(tile)
	if err != nil {
		t.Fatalf("Failed to store tile: %v", err)
	}

	// 2. Check if cached
	checkTile := NewTile(coord)
	if !cache.IsCached(checkTile) {
		t.Error("Expected tile to be cached after storing")
	}

	// 3. Load tile
	loadTile := NewTile(coord)
	err = cache.LoadTile(loadTile, true)
	if err != nil {
		t.Fatalf("Failed to load tile: %v", err)
	}

	// Verify loaded data
	loadedData := loadTile.Source.GetBuffer(nil, nil)
	if string(loadedData) != string(testData) {
		t.Errorf("Expected loaded data %s, got %s", string(testData), string(loadedData))
	}

	// Verify metadata
	if loadTile.Size == 0 {
		t.Error("Expected tile size to be set")
	}
	if loadTile.Timestamp.IsZero() {
		t.Error("Expected tile timestamp to be set")
	}

	// 4. Remove tile
	err = cache.RemoveTile(loadTile)
	if err != nil {
		t.Fatalf("Failed to remove tile: %v", err)
	}

	// 5. Check if no longer cached
	if cache.IsCached(NewTile(coord)) {
		t.Error("Expected tile to not be cached after removal")
	}
}

func TestLocalCache_Store_ReplaceSymlink(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	creater := newLocalCacheMockSourceCreater("png")
	cache := NewLocalCache(tmpDir, "tms", creater)

	// Create a test tile
	tile := createTestTile([3]int{1, 2, 3}, []byte("test data"))
	location := cache.TileLocation(tile, true)

	// Create a symlink at the tile location
	targetFile := filepath.Join(tmpDir, "target.png")
	err := os.WriteFile(targetFile, []byte("target data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	err = os.Symlink(targetFile, location)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Store the tile (should replace symlink)
	err = cache.StoreTile(tile)
	if err != nil {
		t.Fatalf("Failed to store tile over symlink: %v", err)
	}

	// Verify the file contains the new data, not the symlink target
	data, err := os.ReadFile(location)
	if err != nil {
		t.Fatalf("Failed to read stored file: %v", err)
	}
	if string(data) != "test data" {
		t.Errorf("Expected stored data 'test data', got %s", string(data))
	}
}

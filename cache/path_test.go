package cache

import (
	"os"
	"path/filepath"
	"testing"
)

// Test helper functions

func createTestTileForPath(coord [3]int) *Tile {
	return NewTile(coord)
}

func createTempTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "path_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

// Tests for level location functions

func TestLevelLocation(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		cacheDir string
		expected string
	}{
		{"Level 0", 0, "/cache", "/cache/00"},
		{"Level 5", 5, "/cache", "/cache/05"},
		{"Level 15", 15, "/cache", "/cache/15"},
		{"Level 25", 25, "/cache", "/cache/25"},
		{"Different cache dir", 10, "/tmp/tiles", "/tmp/tiles/10"},
		{"Empty cache dir", 3, "", "03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := level_location(tt.level, tt.cacheDir)
			expected := filepath.FromSlash(tt.expected)
			if result != expected {
				t.Errorf("level_location(%d, %q) = %q, expected %q", tt.level, tt.cacheDir, result, expected)
			}
		})
	}
}

func TestNoLevelLocation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("no_level_location should panic")
		}
	}()
	no_level_location(5, "/cache")
}

func TestLevelLocationArcgisCache(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		cacheDir string
		expected string
	}{
		{"Level 0", 0, "/cache", "/cache/L00"},
		{"Level 5", 5, "/cache", "/cache/L05"},
		{"Level 15", 15, "/cache", "/cache/L15"},
		{"Level 25", 25, "/cache", "/cache/L25"},
		{"Different cache dir", 10, "/arcgis/tiles", "/arcgis/tiles/L10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := level_location_arcgiscache(tt.level, tt.cacheDir)
			expected := filepath.FromSlash(tt.expected)
			if result != expected {
				t.Errorf("level_location_arcgiscache(%d, %q) = %q, expected %q", tt.level, tt.cacheDir, result, expected)
			}
		})
	}
}

// Tests for level_part function

func TestLevelPart(t *testing.T) {
	tests := []struct {
		name     string
		level    interface{}
		expected string
	}{
		{"String level", "05", "05"},
		{"String level single digit", "5", "5"},
		{"Int level single digit", 5, "05"},
		{"Int level double digit", 15, "15"},
		{"Int level zero", 0, "00"},
		{"Nil level", nil, ""},
		{"Float level (unsupported)", 5.5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := level_part(tt.level)
			if result != tt.expected {
				t.Errorf("level_part(%v) = %q, expected %q", tt.level, result, tt.expected)
			}
		})
	}
}

// Tests for ensure_directory function

func TestEnsureDirectory(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "sub", "nested", "directory")

	// Directory should not exist initially
	if _, err := os.Stat(testDir); err == nil {
		t.Error("Test directory should not exist initially")
	}

	// Create directory
	ensure_directory(testDir)

	// Directory should now exist
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory should exist after ensure_directory call")
	}

	// Calling again should not cause error
	ensure_directory(testDir)
}

// Tests for tile location functions

func TestTileLocationTc(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Simple coordinates",
			[3]int{12345, 67890, 2},
			"png",
			false,
			"02/000/012/345/000/067/890.png",
		},
		{
			"Zero coordinates",
			[3]int{0, 0, 0},
			"jpg",
			false,
			"00/000/000/000/000/000/000.jpg",
		},
		{
			"Large coordinates",
			[3]int{12345678, 87654321, 15},
			"webp",
			false,
			"15/012/345/678/087/654/321.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_tc(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, filepath.FromSlash(tt.expected))
			if result != expectedPath {
				t.Errorf("tile_location_tc() = %q, expected %q", result, expectedPath)
			}

			// Check if directory creation works
			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

func TestTileLocationMp(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Simple coordinates",
			[3]int{12345, 67890, 2},
			"png",
			false,
			"02/0001/2345/0006/7890.png",
		},
		{
			"Zero coordinates",
			[3]int{0, 0, 0},
			"jpg",
			false,
			"00/0000/0000/0000/0000.jpg",
		},
		{
			"Large coordinates",
			[3]int{123456789, 987654321, 12},
			"webp",
			false,
			"12/12345/6789/98765/4321.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_mp(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, filepath.FromSlash(tt.expected))
			if result != expectedPath {
				t.Errorf("tile_location_mp() = %q, expected %q", result, expectedPath)
			}

			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

func TestTileLocationTms(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Simple coordinates",
			[3]int{1, 2, 3},
			"png",
			false,
			"3/1/2.png",
		},
		{
			"Zero coordinates",
			[3]int{0, 0, 0},
			"jpg",
			false,
			"0/0/0.jpg",
		},
		{
			"Large coordinates",
			[3]int{12345, 67890, 12},
			"webp",
			false,
			"12/12345/67890.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_tms(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, filepath.FromSlash(tt.expected))
			if result != expectedPath {
				t.Errorf("tile_location_tms() = %q, expected %q", result, expectedPath)
			}

			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

func TestTileLocationReverseTms(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Simple coordinates",
			[3]int{1, 2, 3},
			"png",
			false,
			"2/1/3.png",
		},
		{
			"Zero coordinates",
			[3]int{0, 0, 0},
			"jpg",
			false,
			"0/0/0.jpg",
		},
		{
			"Large coordinates",
			[3]int{12345, 67890, 12},
			"webp",
			false,
			"67890/12345/12.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_reverse_tms(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, filepath.FromSlash(tt.expected))
			if result != expectedPath {
				t.Errorf("tile_location_reverse_tms() = %q, expected %q", result, expectedPath)
			}

			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

func TestTileLocationQuadkey(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Level 0 center",
			[3]int{0, 0, 0},
			"png",
			false,
			".png",
		},
		{
			"Level 1 quadrant 0",
			[3]int{0, 0, 1},
			"png",
			false,
			"0.png",
		},
		{
			"Level 1 quadrant 3",
			[3]int{1, 1, 1},
			"png",
			false,
			"3.png",
		},
		{
			"Complex quadkey",
			[3]int{12345, 67890, 12},
			"png",
			false,
			"200200331021.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_quadkey(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, tt.expected)
			if result != expectedPath {
				t.Errorf("tile_location_quadkey() = %q, expected %q", result, expectedPath)
			}

			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

func TestTileLocationArcgisCache(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		coord     [3]int
		fileExt   string
		createDir bool
		expected  string
	}{
		{
			"Simple coordinates",
			[3]int{1, 2, 3},
			"png",
			false,
			"L03/R00000002/C00000001.png",
		},
		{
			"Hex coordinates",
			[3]int{9, 2, 3},
			"png",
			false,
			"L03/R00000002/C00000009.png",
		},
		{
			"Hex coordinates with letters",
			[3]int{10, 2, 3},
			"png",
			false,
			"L03/R00000002/C0000000a.png",
		},
		{
			"Large coordinates",
			[3]int{12345, 67890, 12},
			"png",
			false,
			"L12/R00010932/C00003039.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := createTestTileForPath(tt.coord)
			result, _ := tile_location_arcgiscache(tile, tmpDir, tt.fileExt, tt.createDir)

			expectedPath := filepath.Join(tmpDir, filepath.FromSlash(tt.expected))
			if result != expectedPath {
				t.Errorf("tile_location_arcgiscache() = %q, expected %q", result, expectedPath)
			}

			if tt.createDir {
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}
		})
	}
}

// Test for cached location behavior

func TestTileLocationCaching(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	tile := createTestTileForPath([3]int{1, 2, 3})

	// First call should set the location
	result1, _ := tile_location_tms(tile, tmpDir, "png", false)

	// Second call should return cached location
	result2, _ := tile_location_tms(tile, tmpDir, "jpg", false) // Different extension

	if result1 != result2 {
		t.Errorf("Expected cached location to be returned. First: %q, Second: %q", result1, result2)
	}

	// The extension should still be png (from first call)
	expectedPath := filepath.Join(tmpDir, "3/1/2.png")
	if result2 != expectedPath {
		t.Errorf("Expected cached location with original extension. Got: %q, Expected: %q", result2, expectedPath)
	}
}

// Tests for LocationPaths function

func TestLocationPaths(t *testing.T) {
	tests := []struct {
		name            string
		layout          string
		shouldError     bool
		expectTileFunc  bool
		expectLevelFunc bool
		expectPanic     bool
	}{
		{"TC layout", "tc", false, true, true, false},
		{"MP layout", "mp", false, true, true, false},
		{"TMS layout", "tms", false, true, true, false},
		{"Reverse TMS layout", "reverse_tms", false, true, false, false},
		{"Quadkey layout", "quadkey", false, true, true, true},
		{"ArcGIS layout", "arcgis", false, true, true, false},
		{"Unknown layout", "unknown", true, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tileFunc, levelFunc, err := LocationPaths(tt.layout)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error for unknown layout")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectTileFunc && tileFunc == nil {
				t.Error("Expected tile function to be returned")
			}

			if !tt.expectTileFunc && tileFunc != nil {
				t.Error("Expected tile function to be nil")
			}

			if tt.expectLevelFunc && levelFunc == nil {
				t.Error("Expected level function to be returned")
			}

			if !tt.expectLevelFunc && levelFunc != nil {
				t.Error("Expected level function to be nil")
			}

			// Test quadkey panic case
			if tt.expectPanic && levelFunc != nil {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected level function to panic for quadkey layout")
					}
				}()
				levelFunc(5, "/cache")
			}
		})
	}
}

// Integration tests

func TestLocationPathsIntegration(t *testing.T) {
	tmpDir := createTempTestDir(t)
	defer os.RemoveAll(tmpDir)

	layouts := []string{"tc", "mp", "tms", "reverse_tms", "quadkey", "arcgis"}
	coord := [3]int{12345, 67890, 5}

	for _, layout := range layouts {
		t.Run("Integration_"+layout, func(t *testing.T) {
			tileFunc, levelFunc, err := LocationPaths(layout)
			if err != nil {
				t.Fatalf("Unexpected error for layout %s: %v", layout, err)
			}

			// Test tile function
			if tileFunc != nil {
				tile := createTestTileForPath(coord)
				result, _ := tileFunc(tile, tmpDir, "png", true)

				if result == "" {
					t.Error("Tile function returned empty string")
				}

				// Check if directory was created
				dir := filepath.Dir(result)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory %q should have been created", dir)
				}
			}

			// Test level function (skip quadkey as it panics)
			if levelFunc != nil && layout != "quadkey" {
				levelResult := levelFunc(coord[2], tmpDir)
				if levelResult == "" {
					t.Error("Level function returned empty string")
				}
			}
		})
	}
}

// Benchmark tests

func BenchmarkTileLocationTms(b *testing.B) {
	tile := createTestTileForPath([3]int{12345, 67890, 12})
	cacheDir := "/tmp/cache"
	fileExt := "png"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset location to ensure function execution
		tile.Location = ""
		tile_location_tms(tile, cacheDir, fileExt, false)
	}
}

func BenchmarkTileLocationTc(b *testing.B) {
	tile := createTestTileForPath([3]int{12345, 67890, 12})
	cacheDir := "/tmp/cache"
	fileExt := "png"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tile.Location = ""
		tile_location_tc(tile, cacheDir, fileExt, false)
	}
}

func BenchmarkTileLocationQuadkey(b *testing.B) {
	tile := createTestTileForPath([3]int{12345, 67890, 12})
	cacheDir := "/tmp/cache"
	fileExt := "png"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tile.Location = ""
		tile_location_quadkey(tile, cacheDir, fileExt, false)
	}
}

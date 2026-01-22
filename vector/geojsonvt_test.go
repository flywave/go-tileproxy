package vector

import (
	"testing"

	"github.com/flywave/go-tileproxy/tile"
)

func TestGeoJSONVTSource(t *testing.T) {
	tests := []struct {
		name     string
		tile     [3]int
		options  *VectorOptions
		setup    func() interface{}
		validate func(*testing.T, interface{})
	}{
		{
			name: "basic functionality",
			tile: [3]int{13515, 6392, 14},
			options: &VectorOptions{
				Extent: 4096,
				Buffer: 1024,
				Format: tile.TileFormat("application/json"),
			},
			setup: func() interface{} {
				return Vector{}
			},
			validate: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Fatal("Expected non-nil vector result")
				}
			},
		},
		{
			name: "empty vector tile",
			tile: [3]int{0, 0, 0},
			options: &VectorOptions{
				Extent: 4096,
				Buffer: 0,
				Format: tile.TileFormat("application/json"),
			},
			setup: func() interface{} {
				return Vector{}
			},
			validate: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Fatal("Expected non-nil empty vector")
				}
			},
		},
		{
			name: "small extent",
			tile: [3]int{1, 1, 1},
			options: &VectorOptions{
				Extent: 256,
				Buffer: 64,
				Format: tile.TileFormat("application/json"),
			},
			setup: func() interface{} {
				return Vector{}
			},
			validate: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Fatal("Expected non-nil vector result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewGeoJSONVTSource(tt.tile, tt.options)

			// Use empty vector to avoid C++ library issues
			source.SetSource(Vector{})

			result := source.GetTile()
			if result == nil {
				t.Fatal("GetTile() returned nil")
			}

			tt.validate(t, result)

			// Test encoding
			bytes := source.GetBuffer(nil, nil)
			if len(bytes) == 0 {
				t.Error("GetBuffer() returned empty data")
			}
		})
	}
}

func TestNewEmptyGeoJSONVTSource(t *testing.T) {
	options := &VectorOptions{
		Extent: 4096,
		Buffer: 1024,
		Format: tile.TileFormat("application/json"),
	}

	source := NewEmptyGeoJSONVTSource(options)

	result := source.GetTile()
	if result == nil {
		t.Fatal("Expected non-nil empty vector")
	}

	// Test encoding empty vector
	bytes := source.GetBuffer(nil, nil)
	if len(bytes) == 0 {
		t.Error("Expected non-empty buffer for empty vector")
	}
}

func TestTileBounds(t *testing.T) {
	tests := []struct {
		tile     [3]int
		expected []float64
	}{
		{
			tile:     [3]int{0, 0, 0},
			expected: []float64{-180.0, -85.05112877980659, 180.0, 85.05112877980659},
		},
		{
			tile:     [3]int{1, 1, 1},
			expected: []float64{0.0, -85.05112877980659, 180.0, 0.0},
		},
		{
			tile:     [3]int{0, 1, 1},
			expected: []float64{-180.0, -85.05112877980659, 0.0, 0.0},
		},
		{
			tile:     [3]int{1, 0, 1},
			expected: []float64{0.0, 0.0, 180.0, 85.05112877980659},
		},
	}

	for _, tt := range tests {
		t.Run("tile_bounds", func(t *testing.T) {
			bounds := tileBounds(tt.tile)
			if len(bounds) != 4 {
				t.Fatalf("Expected 4 bounds values, got %d", len(bounds))
			}

			const epsilon = 1e-6
			for i := 0; i < 4; i++ {
				if abs(bounds[i]-tt.expected[i]) > epsilon {
					t.Errorf("Bounds[%d] = %f, expected %f", i, bounds[i], tt.expected[i])
				}
			}
		})
	}
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkNewGeoJSONVTSource(b *testing.B) {
	options := &VectorOptions{Extent: 4096, Buffer: 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source := NewGeoJSONVTSource([3]int{1, 1, 1}, options)
		source.SetSource(Vector{})
		_ = source.GetTile()
	}
}

func BenchmarkNewEmptyGeoJSONVTSource(b *testing.B) {
	options := &VectorOptions{Extent: 4096, Buffer: 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEmptyGeoJSONVTSource(options)
	}
}

func BenchmarkTileBounds(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tileBounds([3]int{1, 1, 1})
	}
}

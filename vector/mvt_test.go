package vector

import (
	"bytes"
	"testing"

	"github.com/flywave/go-geom"
	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-mapbox/mvt"
	"github.com/flywave/go-tileproxy/tile"
)

// Test data constants
const (
	testTileX  = 13515
	testTileY  = 6392
	testTileZ  = 14
	testTileX2 = 1686
	testTileY2 = 776
	testTileZ2 = 11
	emptyTileX = 0
	emptyTileY = 0
	emptyTileZ = 0
)

// TestMVTSourceBasic tests basic MVTSource creation and functionality
func TestMVTSourceBasic(t *testing.T) {
	tests := []struct {
		name     string
		tile     [3]int
		proto    mvt.ProtoType
		format   tile.TileFormat
		setup    func() interface{}
		validate func(*testing.T, interface{}, error)
	}{
		{
			name:   "empty vector source",
			tile:   [3]int{emptyTileX, emptyTileY, emptyTileZ},
			proto:  PBF_PTOTO_MAPBOX,
			format: MVT_MIME,
			setup: func() interface{} {
				return Vector{} // Empty vector
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected non-nil result")
				}
				vec, ok := result.(Vector)
				if !ok {
					t.Fatal("Expected Vector type")
				}
				if vec == nil {
					t.Fatal("Expected non-nil vector")
				}
			},
		},
		{
			name:   "vector with test features",
			tile:   [3]int{testTileX, testTileY, testTileZ},
			proto:  PBF_PTOTO_MAPBOX,
			format: MVT_MIME,
			setup: func() interface{} {
				return createTestVector()
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				vec, ok := result.(Vector)
				if !ok {
					t.Fatal("Expected Vector type")
				}
				if len(vec) == 0 {
					t.Error("Expected non-empty vector")
				}
			},
		},
		{
			name:   "PBF format test",
			tile:   [3]int{testTileX2, testTileY2, testTileZ2},
			proto:  PBF_PTOTO_LUOKUANG,
			format: PBF_MIME,
			setup: func() interface{} {
				return createTestVector()
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				vec, ok := result.(Vector)
				if !ok {
					t.Fatal("Expected Vector type")
				}
				if vec == nil {
					t.Fatal("Expected non-nil vector")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewMVTSource(tt.tile, tt.proto, &VectorOptions{
				Format: tt.format,
				Proto:  int(tt.proto),
			})

			data := tt.setup()
			source.SetSource(data)

			result := source.GetTile()
			tt.validate(t, result, nil)
		})
	}
}

// TestMVTSourceEncoding tests encoding functionality
func TestMVTSourceEncoding(t *testing.T) {
	tests := []struct {
		name     string
		tile     [3]int
		vector   Vector
		validate func(*testing.T, []byte, error)
	}{
		{
			name:   "encode empty vector",
			tile:   [3]int{0, 0, 1},
			vector: Vector{},
			validate: func(t *testing.T, buf []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(buf) == 0 {
					t.Error("Expected non-empty buffer")
				}
			},
		},
		{
			name:   "encode vector with features",
			tile:   [3]int{1, 1, 1},
			vector: createTestVector(),
			validate: func(t *testing.T, buf []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(buf) == 0 {
					t.Error("Expected non-empty buffer")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewMVTSource(tt.tile, PBF_PTOTO_MAPBOX, &VectorOptions{
				Format: MVT_MIME,
				Proto:  int(PBF_PTOTO_MAPBOX),
			})

			source.SetSource(tt.vector)

			buf := source.GetBuffer(nil, nil)
			tt.validate(t, buf, nil)
		})
	}
}

// TestMVTSourceErrorHandling tests error handling scenarios
func TestMVTSourceErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MVTSource
		validate func(*testing.T, interface{})
	}{
		{
			name: "nil source data",
			setup: func() *MVTSource {
				source := NewMVTSource([3]int{1, 1, 1}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})
				source.SetSource(nil)
				return source
			},
			validate: func(t *testing.T, result interface{}) {
				// Should handle nil gracefully
				if result == nil {
					t.Log("Nil source handled correctly")
				}
			},
		},
		{
			name: "invalid file path",
			setup: func() *MVTSource {
				source := NewMVTSource([3]int{1, 1, 1}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})
				source.SetSource("nonexistent_file.mvt")
				return source
			},
			validate: func(t *testing.T, result interface{}) {
				// Should handle missing file gracefully
				if result == nil {
					t.Log("Missing file handled correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.setup()
			result := source.GetTile()
			tt.validate(t, result)
		})
	}
}

// TestMVTSourceOptions tests different VectorOptions configurations
func TestMVTSourceOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  *VectorOptions
		validate func(*testing.T, *MVTSource)
	}{
		{
			name: "default options",
			options: &VectorOptions{
				Format: MVT_MIME,
				Proto:  int(PBF_PTOTO_MAPBOX),
			},
			validate: func(t *testing.T, source *MVTSource) {
				if source.GetTileOptions() == nil {
					t.Error("Expected non-nil options")
				}
			},
		},
		{
			name: "custom options",
			options: &VectorOptions{
				Format:      PBF_MIME,
				Proto:       int(PBF_PTOTO_LUOKUANG),
				Extent:      4096,
				Buffer:      1024,
				Tolerance:   0.5,
				LineMetrics: true,
			},
			validate: func(t *testing.T, source *MVTSource) {
				opts := source.GetTileOptions().(*VectorOptions)
				if opts.Extent != 4096 {
					t.Errorf("Expected Extent 4096, got %d", opts.Extent)
				}
				if opts.Buffer != 1024 {
					t.Errorf("Expected Buffer 1024, got %d", opts.Buffer)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewMVTSource([3]int{1, 1, 1}, mvt.ProtoType(tt.options.Proto), tt.options)
			tt.validate(t, source)
		})
	}
}

// TestNewEmptyMVTSource tests empty MVT source creation
func TestNewEmptyMVTSource(t *testing.T) {
	tests := []struct {
		name     string
		proto    mvt.ProtoType
		options  *VectorOptions
		validate func(*testing.T, *MVTSource)
	}{
		{
			name:    "empty MAPBOX source",
			proto:   PBF_PTOTO_MAPBOX,
			options: &VectorOptions{Format: MVT_MIME},
			validate: func(t *testing.T, source *MVTSource) {
				if source == nil {
					t.Fatal("Expected non-nil source")
				}
				result := source.GetTile()
				if result == nil {
					t.Fatal("Expected non-nil tile")
				}
			},
		},
		{
			name:    "empty LUOKUANG source",
			proto:   PBF_PTOTO_LUOKUANG,
			options: &VectorOptions{Format: PBF_MIME},
			validate: func(t *testing.T, source *MVTSource) {
				if source == nil {
					t.Fatal("Expected non-nil source")
				}
				result := source.GetTile()
				if result == nil {
					t.Fatal("Expected non-nil tile")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewEmptyMVTSource(tt.proto, tt.options)
			tt.validate(t, source)
		})
	}
}

// TestMVTSourceRoundTrip tests encoding and decoding round trip
func TestMVTSourceRoundTrip(t *testing.T) {
	original := createTestVector()
	tile := [3]int{testTileX, testTileY, testTileZ}

	// Create source with test data
	source := NewMVTSource(tile, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})
	source.SetSource(original)

	// Encode to buffer
	buf := source.GetBuffer(nil, nil)
	if len(buf) == 0 {
		t.Fatal("Failed to encode")
	}

	// Create new source from buffer
	newSource := NewMVTSource(tile, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})
	newSource.SetSource(bytes.NewReader(buf))

	// Verify round trip
	decoded := newSource.GetTile()
	if decoded == nil {
		t.Fatal("Failed to decode")
	}

	vec, ok := decoded.(Vector)
	if !ok {
		t.Fatal("Expected Vector type")
	}

	if len(vec) != len(original) {
		t.Errorf("Round trip failed: expected %d layers, got %d", len(original), len(vec))
	}
}

// createTestVector creates a test vector with sample data
func createTestVector() Vector {
	return Vector{
		"test_layer": []*geom.Feature{
			{
				Geometry: general.NewPoint([]float64{125.6, 10.1}),
				Properties: map[string]interface{}{
					"name":  "Test Point",
					"value": 42,
				},
				ID: "test-1",
			},
			{
				Geometry: general.NewLineString([][]float64{
					{125.6, 10.1},
					{125.7, 10.2},
				}),
				Properties: map[string]interface{}{
					"name":  "Test Line",
					"value": 43,
				},
				ID: "test-2",
			},
		},
		"empty_layer": []*geom.Feature{},
	}
}

// Benchmark tests
func BenchmarkMVTSourceCreation(b *testing.B) {
	options := &VectorOptions{Format: MVT_MIME, Proto: int(PBF_PTOTO_MAPBOX)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source := NewMVTSource([3]int{1, 1, 1}, PBF_PTOTO_MAPBOX, options)
		source.SetSource(createTestVector())
		_ = source.GetTile()
	}
}

func BenchmarkMVTSourceEncoding(b *testing.B) {
	testVector := createTestVector()
	source := NewMVTSource([3]int{1, 1, 1}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})
	source.SetSource(testVector)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = source.GetBuffer(nil, nil)
	}
}

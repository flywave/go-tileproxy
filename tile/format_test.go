package tile

import "testing"

func TestTileFormat_MimeType(t *testing.T) {
	tests := []struct {
		name     string
		format   TileFormat
		expected string
	}{
		{"mvt extension", "mvt", "application/vnd.mapbox-vector-tile"},
		{"pbf extension", "pbf", "application/x-protobuf"},
		{"omv extension", "omv", "application/x-protobuf"},
		{"png extension", "png", "image/png"},
		{"jpg extension", "jpg", "image/jpeg"},
		{"jpeg extension", "jpeg", "image/jpeg"},
		{"tiff extension", "tiff", "image/tiff"},
		{"webp extension", "webp", "image/webp"},
		{"json extension", "json", "application/json"},
		{"geojson extension", "geojson", "application/json"},
		{"terrain extension", "terrain", "application/vnd.quantized-mesh,application/octet-stream;q=1.0"},
		{"full mimetype", "application/vnd.mapbox-vector-tile", "application/vnd.mapbox-vector-tile"},
		{"full image mimetype", "image/png", "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.MimeType(); got != tt.expected {
				t.Errorf("MimeType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTileFormat_Extension(t *testing.T) {
	tests := []struct {
		name     string
		format   TileFormat
		expected string
	}{
		{"mvt mimetype", "application/vnd.mapbox-vector-tile", "mvt"},
		{"pbf mimetype", "application/x-protobuf", "pbf"},
		{"omv mimetype", "application/x-protobuf", "pbf"},
		{"png mimetype", "image/png", "png"},
		{"jpg mimetype", "image/jpeg", "jpeg"},
		{"jpeg mimetype", "image/jpeg", "jpeg"},
		{"tiff mimetype", "image/tiff", "tiff"},
		{"webp mimetype", "image/webp", "webp"},
		{"json mimetype", "application/json", "geojson"},
		{"terrain mimetype", "application/vnd.quantized-mesh", "terrain"},
		{"with parameters", "application/vnd.mapbox-vector-tile; charset=utf-8", "mvt"},
		{"with spaces", " image/png ", "png"},
		{"raw extension", "png", "png"},
		{"raw mvt", "mvt", "mvt"},
		{"raw pbf", "pbf", "pbf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.Extension(); got != tt.expected {
				t.Errorf("Extension() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormat_EdgeCases(t *testing.T) {
	// Test empty format
	f := TileFormat("")
	if f.MimeType() != "" {
		t.Errorf("Empty format MimeType() should return empty string, got %v", f.MimeType())
	}
	if f.Extension() != "" {
		t.Errorf("Empty format Extension() should return empty string, got %v", f.Extension())
	}

	// Test unknown format
	f = TileFormat("unknown")
	if f.MimeType() != "unknown" {
		t.Errorf("Unknown format MimeType() should return original string, got %v", f.MimeType())
	}
	if f.Extension() != "unknown" {
		t.Errorf("Unknown format Extension() should return original string, got %v", f.Extension())
	}

	// Test format with slash but no known extension
	f = TileFormat("custom/type")
	if f.Extension() != "type" {
		t.Errorf("Custom format Extension() should extract type part, got %v", f.Extension())
	}
}

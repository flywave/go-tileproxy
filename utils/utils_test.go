package utils

import (
	"os"
	"path/filepath"
	"testing"

	colorful "github.com/lucasb-eyer/go-colorful"
)

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "existing file",
			setup: func() string {
				file := filepath.Join(tempDir, "testfile.txt")
				os.WriteFile(file, []byte("test content"), 0644)
				return file
			},
			expected: true,
		},
		{
			name: "non-existing file",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.txt")
			},
			expected: false,
		},
		{
			name: "existing directory",
			setup: func() string {
				dir := filepath.Join(tempDir, "testdir")
				os.Mkdir(dir, 0755)
				return dir
			},
			expected: true,
		},
		{
			name: "empty path",
			setup: func() string {
				return ""
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := FileExists(path)
			if result != tt.expected {
				t.Errorf("FileExists() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected bool
		wantErr  bool
	}{
		{
			name: "existing directory",
			setup: func() string {
				dir := filepath.Join(tempDir, "testdir")
				os.Mkdir(dir, 0755)
				return dir
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "existing file",
			setup: func() string {
				file := filepath.Join(tempDir, "testfile.txt")
				os.WriteFile(file, []byte("test content"), 0644)
				return file
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "non-existing path",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result, err := IsDir(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("IsDir() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSymlink(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected bool
		wantErr  bool
	}{
		{
			name: "regular file",
			setup: func() string {
				file := filepath.Join(tempDir, "testfile.txt")
				os.WriteFile(file, []byte("test content"), 0644)
				return file
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "symlink to file",
			setup: func() string {
				file := filepath.Join(tempDir, "testfile.txt")
				link := filepath.Join(tempDir, "testlink.txt")
				os.WriteFile(file, []byte("test content"), 0644)
				os.Symlink(file, link)
				return link
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "symlink to directory",
			setup: func() string {
				dir := filepath.Join(tempDir, "testdir")
				link := filepath.Join(tempDir, "testlinkdir")
				os.Mkdir(dir, 0755)
				os.Symlink(dir, link)
				return link
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "non-existing path",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result, err := IsSymlink(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSymlink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("IsSymlink() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHexColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected colorful.Color
	}{
		{
			name:     "valid hex color",
			input:    "#FF0000",
			expected: colorful.Color{R: 1, G: 0, B: 0},
		},
		{
			name:     "valid short hex color",
			input:    "#F00",
			expected: colorful.Color{R: 1, G: 0, B: 0},
		},
		{
			name:     "valid hex with lowercase",
			input:    "#00ff00",
			expected: colorful.Color{R: 0, G: 1, B: 0},
		},
		{
			name:     "valid hex with alpha",
			input:    "#FF0000FF",
			expected: colorful.Color{R: 1, G: 0, B: 0},
		},
		{
			name:     "invalid hex color",
			input:    "invalid",
			expected: colorful.Color{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: colorful.Color{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HexColor(tt.input)

			// 比较RGB值，允许小的浮点误差
			if !colorsEqual(result, tt.expected) {
				t.Errorf("HexColor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prefix         string
		expectedBool   bool
		expectedString string
	}{
		{
			name:           "prefix exists",
			input:          "hello world",
			prefix:         "hello",
			expectedBool:   true,
			expectedString: " world",
		},
		{
			name:           "prefix does not exist",
			input:          "hello world",
			prefix:         "goodbye",
			expectedBool:   false,
			expectedString: "hello world",
		},
		{
			name:           "empty prefix",
			input:          "hello world",
			prefix:         "",
			expectedBool:   true,
			expectedString: "hello world",
		},
		{
			name:           "empty input",
			input:          "",
			prefix:         "hello",
			expectedBool:   false,
			expectedString: "",
		},
		{
			name:           "prefix equals input",
			input:          "hello",
			prefix:         "hello",
			expectedBool:   true,
			expectedString: "",
		},
		{
			name:           "case sensitive",
			input:          "Hello World",
			prefix:         "hello",
			expectedBool:   false,
			expectedString: "Hello World",
		},
		{
			name:           "unicode prefix",
			input:          "你好世界",
			prefix:         "你好",
			expectedBool:   true,
			expectedString: "世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boolResult, stringResult := HasPrefix(tt.input, tt.prefix)
			if boolResult != tt.expectedBool {
				t.Errorf("HasPrefix() bool = %v, want %v", boolResult, tt.expectedBool)
			}
			if stringResult != tt.expectedString {
				t.Errorf("HasPrefix() string = %q, want %q", stringResult, tt.expectedString)
			}
		})
	}
}

// 辅助函数：比较两个颜色是否相等（允许小的浮点误差）
func colorsEqual(c1, c2 colorful.Color) bool {
	const epsilon = 1e-6
	return abs(c1.R-c2.R) < epsilon &&
		abs(c1.G-c2.G) < epsilon &&
		abs(c1.B-c2.B) < epsilon
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkFileExists(b *testing.B) {
	tempDir := b.TempDir()
	file := filepath.Join(tempDir, "testfile.txt")
	os.WriteFile(file, []byte("test content"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(file)
	}
}

func BenchmarkIsDir(b *testing.B) {
	tempDir := b.TempDir()
	dir := filepath.Join(tempDir, "testdir")
	os.Mkdir(dir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsDir(dir)
	}
}

func BenchmarkHexColor(b *testing.B) {
	color := "#FF0000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HexColor(color)
	}
}

func BenchmarkHasPrefix(b *testing.B) {
	input := "hello world"
	prefix := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HasPrefix(input, prefix)
	}
}

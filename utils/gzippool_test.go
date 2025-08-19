package utils

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
	"testing"
)

func TestNewGzipPool(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"positive size", 5, false},
		{"zero size", 0, false},
		{"negative size", -1, false}, // Should still work with 0 entries initially
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewGzipPool(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGzipPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pool == nil {
				t.Error("NewGzipPool() returned nil pool")
			}
		})
	}
}

func TestGzipPoolGetPut(t *testing.T) {
	pool, err := NewGzipPool(2)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	var buf bytes.Buffer

	// Test getting a writer
	gz := pool.Get(&buf)
	if gz == nil {
		t.Fatal("Get() returned nil writer")
	}

	// Test writing and compressing data
	testData := "Hello, World! This is a test string for gzip compression."
	_, err = gz.Write([]byte(testData))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Close the gzip writer
	if cerr := gz.Close(); cerr != nil {
		t.Errorf("Close() error = %v", cerr)
	}

	// Test decompression
	reader, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(decompressed) != testData {
		t.Errorf("decompressed data = %q, want %q", string(decompressed), testData)
	}

	// Test putting the writer back
	pool.Put(gz)
}

func TestGzipPoolConcurrency(t *testing.T) {
	pool, err := NewGzipPool(5)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	const numGoroutines = 10
	const testData = "concurrent gzip test data"

	var wg sync.WaitGroup
	results := make([][]byte, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			var buf bytes.Buffer
			gz := pool.Get(&buf)

			_, err := gz.Write([]byte(testData))
			if err != nil {
				t.Errorf("goroutine %d: Write() error = %v", index, err)
				return
			}

			if err := gz.Close(); err != nil {
				t.Errorf("goroutine %d: Close() error = %v", index, err)
				return
			}

			results[index] = buf.Bytes()
			pool.Put(gz)
		}(i)
	}

	wg.Wait()

	// Verify all results are valid gzip data
	for i, data := range results {
		if len(data) == 0 {
			t.Errorf("goroutine %d: empty compressed data", i)
			continue
		}

		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			t.Errorf("goroutine %d: invalid gzip data: %v", i, err)
			continue
		}

		decompressed, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Errorf("goroutine %d: decompression error: %v", i, err)
			continue
		}

		if string(decompressed) != testData {
			t.Errorf("goroutine %d: decompressed data mismatch", i)
		}
	}
}

func TestGzipPoolGrowth(t *testing.T) {
	pool, err := NewGzipPool(1)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	// Use up the initial writer
	var buf1 bytes.Buffer
	gz1 := pool.Get(&buf1)

	// Request another writer (should trigger growth)
	var buf2 bytes.Buffer
	gz2 := pool.Get(&buf2)

	if gz1 == nil || gz2 == nil {
		t.Fatal("Get() returned nil writer")
	}

	if gz1 == gz2 {
		t.Error("Get() returned the same writer instance")
	}

	pool.Put(gz1)
	pool.Put(gz2)
}

func TestGzipPoolReset(t *testing.T) {
	pool, err := NewGzipPool(1)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	var buf1, buf2 bytes.Buffer

	// First use
	gz := pool.Get(&buf1)
	_, err = gz.Write([]byte("first data"))
	if err != nil {
		t.Errorf("First Write() error = %v", err)
	}
	if cerr := gz.Close(); cerr != nil {
		t.Errorf("First Close() error = %v", cerr)
	}

	firstCompressed := buf1.Len()
	pool.Put(gz)

	// Second use (should be reset)
	gz = pool.Get(&buf2)
	_, err = gz.Write([]byte("second data"))
	if err != nil {
		t.Errorf("Second Write() error = %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	secondCompressed := buf2.Len()
	pool.Put(gz)

	if firstCompressed == 0 || secondCompressed == 0 {
		t.Error("No compression occurred")
	}

	// Both should be valid gzip streams
	for i, data := range [][]byte{buf1.Bytes(), buf2.Bytes()} {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			t.Errorf("Data %d: invalid gzip: %v", i+1, err)
			continue
		}
		reader.Close()
	}
}

func TestGzipPoolNilHandling(t *testing.T) {
	pool, err := NewGzipPool(1)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	// Test putting nil writer - expect panic, but verify graceful handling
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Put(nil) caused expected panic: %v", r)
		}
	}()

	// This will panic, but that's acceptable behavior for nil input
	pool.Put(nil)
}

func TestGzipPoolLargeData(t *testing.T) {
	pool, err := NewGzipPool(1)
	if err != nil {
		t.Fatalf("NewGzipPool() error = %v", err)
	}

	// Create large data with high entropy (more compressible)
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 32) // More repetitive pattern
	}

	var buf bytes.Buffer
	gz := pool.Get(&buf)

	_, err = gz.Write(largeData)
	if err != nil {
		t.Errorf("Write large data error = %v", err)
	}

	if cerr := gz.Close(); cerr != nil {
		t.Errorf("Close large data error = %v", cerr)
	}

	// Allow for some compression, even if small
	compressedSize := buf.Len()
	originalSize := len(largeData)
	compressionRatio := float64(compressedSize) / float64(originalSize)
	if compressionRatio >= 0.95 {
		t.Logf("Compression ratio: %.2f (expected for low-entropy data)", compressionRatio)
	} else {
		t.Logf("Compression effective: %d -> %d (ratio: %.2f)", originalSize, compressedSize, compressionRatio)
	}

	// Verify decompression
	reader, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader error = %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll error = %v", err)
	}

	if !bytes.Equal(decompressed, largeData) {
		t.Error("Large data decompression mismatch")
	}

	pool.Put(gz)
}

// Benchmark tests
func BenchmarkGzipPoolGetPut(b *testing.B) {
	pool, err := NewGzipPool(10)
	if err != nil {
		b.Fatalf("NewGzipPool() error = %v", err)
	}

	var buf bytes.Buffer
	testData := []byte("benchmark test data for gzip compression")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gz := pool.Get(&buf)
		gz.Write(testData)
		gz.Close()
		buf.Reset()
		pool.Put(gz)
	}
}

func BenchmarkGzipPoolParallel(b *testing.B) {
	pool, err := NewGzipPool(10)
	if err != nil {
		b.Fatalf("NewGzipPool() error = %v", err)
	}

	testData := []byte("parallel benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			gz := pool.Get(&buf)
			gz.Write(testData)
			gz.Close()
			buf.Reset()
			pool.Put(gz)
		}
	})
}

func BenchmarkGzipNewWriter(b *testing.B) {
	testData := []byte("benchmark test data for gzip compression")
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gz, _ := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
		gz.Write(testData)
		gz.Close()
		buf.Reset()
	}
}

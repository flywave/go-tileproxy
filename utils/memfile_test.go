package utils

import (
	"bytes"
	"io"
	"testing"
)

func TestNewMemFile(t *testing.T) {
	data := []byte("hello world")
	memFile := NewMemFile(data)

	if memFile == nil {
		t.Fatal("NewMemFile() returned nil")
	}

	if !bytes.Equal(memFile.Bytes(), data) {
		t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), data)
	}
}

func TestMemFileRead(t *testing.T) {
	data := []byte("hello world")
	memFile := NewMemFile(data)

	t.Run("read entire content", func(t *testing.T) {
		buf := make([]byte, len(data))
		n, err := memFile.Read(buf)
		if err != nil && err != io.EOF {
			t.Errorf("Read() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Read() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(buf, data) {
			t.Errorf("Read() buf = %v, want %v", buf, data)
		}
	})

	t.Run("read partial", func(t *testing.T) {
		memFile := NewMemFile(data) // Reset position
		buf := make([]byte, 5)
		n, err := memFile.Read(buf)
		if err != nil {
			t.Errorf("Read() error = %v", err)
		}
		if n != 5 {
			t.Errorf("Read() n = %d, want 5", n)
		}
		if !bytes.Equal(buf, data[:5]) {
			t.Errorf("Read() buf = %v, want %v", buf, data[:5])
		}
	})

	t.Run("read at EOF", func(t *testing.T) {
		memFile := NewMemFile(data)
		// Read all content first
		buf := make([]byte, len(data))
		memFile.Read(buf)

		// Try to read more
		buf = make([]byte, 1)
		n, err := memFile.Read(buf)
		if err != io.EOF {
			t.Errorf("Read() at EOF error = %v, want EOF", err)
		}
		if n != 0 {
			t.Errorf("Read() at EOF n = %d, want 0", n)
		}
	})
}

func TestMemFileReadAt(t *testing.T) {
	data := []byte("hello world")
	memFile := NewMemFile(data)

	t.Run("read from middle", func(t *testing.T) {
		buf := make([]byte, 5)
		n, err := memFile.ReadAt(buf, 6)
		if err != nil {
			t.Errorf("ReadAt() error = %v", err)
		}
		if n != 5 {
			t.Errorf("ReadAt() n = %d, want 5", n)
		}
		if !bytes.Equal(buf, []byte("world")) {
			t.Errorf("ReadAt() buf = %v, want %v", buf, []byte("world"))
		}
	})

	t.Run("read at start", func(t *testing.T) {
		buf := make([]byte, 5)
		n, err := memFile.ReadAt(buf, 0)
		if err != nil {
			t.Errorf("ReadAt() error = %v", err)
		}
		if n != 5 {
			t.Errorf("ReadAt() n = %d, want 5", n)
		}
		if !bytes.Equal(buf, []byte("hello")) {
			t.Errorf("ReadAt() buf = %v, want %v", buf, []byte("hello"))
		}
	})

	t.Run("read beyond end", func(t *testing.T) {
		buf := make([]byte, 10)
		n, err := memFile.ReadAt(buf, 8)
		if err != io.EOF {
			t.Errorf("ReadAt() beyond end error = %v, want EOF", err)
		}
		if n != 3 {
			t.Errorf("ReadAt() beyond end n = %d, want 3", n)
		}
		if !bytes.Equal(buf[:3], []byte("rld")) {
			t.Errorf("ReadAt() beyond end buf = %v, want %v", buf[:3], []byte("rld"))
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		buf := make([]byte, 1)
		_, err := memFile.ReadAt(buf, -1)
		if err != errInvalid {
			t.Errorf("ReadAt() with invalid offset error = %v, want errInvalid", err)
		}
	})
}

func TestMemFileWrite(t *testing.T) {
	t.Run("write to empty", func(t *testing.T) {
		memFile := NewMemFile(nil)
		data := []byte("hello")

		n, err := memFile.Write(data)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Write() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(memFile.Bytes(), data) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), data)
		}
	})

	t.Run("write append", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))
		data := []byte(" world")

		// Seek to end
		memFile.Seek(0, io.SeekEnd)

		n, err := memFile.Write(data)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Write() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(memFile.Bytes(), []byte("hello world")) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), []byte("hello world"))
		}
	})

	t.Run("write at position", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello world"))
		data := []byte("WORLD")

		// Seek to position 6
		memFile.Seek(6, io.SeekStart)

		n, err := memFile.Write(data)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Write() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(memFile.Bytes(), []byte("hello WORLD")) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), []byte("hello WORLD"))
		}
	})
}

func TestMemFileWriteAt(t *testing.T) {
	t.Run("write at start", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello world"))
		data := []byte("HELLO")

		n, err := memFile.WriteAt(data, 0)
		if err != nil {
			t.Errorf("WriteAt() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("WriteAt() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(memFile.Bytes(), []byte("HELLO world")) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), []byte("HELLO world"))
		}
	})

	t.Run("write at middle", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello world"))
		data := []byte("WORLD")

		n, err := memFile.WriteAt(data, 6)
		if err != nil {
			t.Errorf("WriteAt() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("WriteAt() n = %d, want %d", n, len(data))
		}
		if !bytes.Equal(memFile.Bytes(), []byte("hello WORLD")) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), []byte("hello WORLD"))
		}
	})

	t.Run("write beyond end", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))
		data := []byte("world")

		n, err := memFile.WriteAt(data, 10)
		if err != nil {
			t.Errorf("WriteAt() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("WriteAt() n = %d, want %d", n, len(data))
		}

		expected := []byte("hello\x00\x00\x00\x00\x00world")
		if !bytes.Equal(memFile.Bytes(), expected) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), expected)
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))
		data := []byte("world")

		_, err := memFile.WriteAt(data, -1)
		if err != errInvalid {
			t.Errorf("WriteAt() with invalid offset error = %v, want errInvalid", err)
		}
	})
}

func TestMemFileSeek(t *testing.T) {
	data := []byte("hello world")
	memFile := NewMemFile(data)

	t.Run("seek start", func(t *testing.T) {
		pos, err := memFile.Seek(6, io.SeekStart)
		if err != nil {
			t.Errorf("Seek() error = %v", err)
		}
		if pos != 6 {
			t.Errorf("Seek() pos = %d, want 6", pos)
		}

		buf := make([]byte, 5)
		n, err := memFile.Read(buf)
		if err != nil {
			t.Errorf("Read() after seek error = %v", err)
		}
		if n != 5 {
			t.Errorf("Read() after seek n = %d, want 5", n)
		}
		if !bytes.Equal(buf, []byte("world")) {
			t.Errorf("Read() after seek buf = %v, want %v", buf, []byte("world"))
		}
	})

	t.Run("seek current", func(t *testing.T) {
		memFile := NewMemFile(data) // Reset
		
		// Read 6 bytes first
		buf := make([]byte, 6)
		memFile.Read(buf)
		
		// Seek relative to current position
		pos, err := memFile.Seek(-1, io.SeekCurrent)
		if err != nil {
			t.Errorf("Seek() error = %v", err)
		}
		if pos != 5 {
			t.Errorf("Seek() pos = %d, want 5", pos)
		}
		
		buf = make([]byte, 6)
		n, err := memFile.Read(buf)
		if err != nil {
			t.Errorf("Read() after seek error = %v", err)
		}
		if n != 6 {
			t.Errorf("Read() after seek n = %d, want 6", n)
		}
		if !bytes.Equal(buf, []byte(" world")) {
			t.Errorf("Read() after seek buf = %v, want %v", buf, []byte(" world"))
		}
	})

	t.Run("seek end", func(t *testing.T) {
		memFile := NewMemFile(data) // Reset

		pos, err := memFile.Seek(-5, io.SeekEnd)
		if err != nil {
			t.Errorf("Seek() error = %v", err)
		}
		if pos != 6 {
			t.Errorf("Seek() pos = %d, want 6", pos)
		}

		buf := make([]byte, 5)
		n, err := memFile.Read(buf)
		if err != nil {
			t.Errorf("Read() after seek error = %v", err)
		}
		if n != 5 {
			t.Errorf("Read() after seek n = %d, want 5", n)
		}
		if !bytes.Equal(buf, []byte("world")) {
			t.Errorf("Read() after seek buf = %v, want %v", buf, []byte("world"))
		}
	})

	t.Run("invalid whence", func(t *testing.T) {
		memFile := NewMemFile(data)
		_, err := memFile.Seek(0, 99)
		if err != errInvalid {
			t.Errorf("Seek() with invalid whence error = %v, want errInvalid", err)
		}
	})

	t.Run("negative position", func(t *testing.T) {
		memFile := NewMemFile(data)
		_, err := memFile.Seek(-100, io.SeekStart)
		if err != errInvalid {
			t.Errorf("Seek() with negative position error = %v, want errInvalid", err)
		}
	})
}

func TestMemFileTruncate(t *testing.T) {
	t.Run("truncate shorter", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello world"))

		err := memFile.Truncate(5)
		if err != nil {
			t.Errorf("Truncate() error = %v", err)
		}
		if !bytes.Equal(memFile.Bytes(), []byte("hello")) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), []byte("hello"))
		}
	})

	t.Run("truncate longer", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))

		err := memFile.Truncate(10)
		if err != nil {
			t.Errorf("Truncate() error = %v", err)
		}
		expected := []byte("hello\x00\x00\x00\x00\x00")
		if !bytes.Equal(memFile.Bytes(), expected) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), expected)
		}
	})

	t.Run("truncate same length", func(t *testing.T) {
		data := []byte("hello")
		memFile := NewMemFile(data)

		err := memFile.Truncate(5)
		if err != nil {
			t.Errorf("Truncate() error = %v", err)
		}
		if !bytes.Equal(memFile.Bytes(), data) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), data)
		}
	})

	t.Run("truncate empty", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))

		err := memFile.Truncate(0)
		if err != nil {
			t.Errorf("Truncate() error = %v", err)
		}
		if len(memFile.Bytes()) != 0 {
			t.Errorf("Bytes() length = %d, want 0", len(memFile.Bytes()))
		}
	})

	t.Run("invalid truncate", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))

		err := memFile.Truncate(-1)
		if err != errInvalid {
			t.Errorf("Truncate() with invalid size error = %v, want errInvalid", err)
		}
	})
}

func TestMemFileBytes(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		memFile := NewMemFile(nil)
		if len(memFile.Bytes()) != 0 {
			t.Errorf("Bytes() length = %d, want 0", len(memFile.Bytes()))
		}
	})

	t.Run("non-empty file", func(t *testing.T) {
		data := []byte("test data")
		memFile := NewMemFile(data)
		if !bytes.Equal(memFile.Bytes(), data) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), data)
		}
	})

	t.Run("after modifications", func(t *testing.T) {
		memFile := NewMemFile([]byte("hello"))
		// Reset position to end before writing
		memFile.Seek(0, io.SeekEnd)
		memFile.Write([]byte(" world"))

		expected := []byte("hello world")
		if !bytes.Equal(memFile.Bytes(), expected) {
			t.Errorf("Bytes() = %v, want %v", memFile.Bytes(), expected)
		}
	})
}

func TestMemFileConcurrentAccess(t *testing.T) {
	data := []byte("concurrent test data")
	memFile := NewMemFile(data)

	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			buf := make([]byte, len(data))
			memFile.ReadAt(buf, 0)
			done <- true
		}()
	}

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func() {
			memFile.Write([]byte("write"))
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify no panic occurred
	_ = memFile.Bytes()
}

func BenchmarkMemFileRead(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	memFile := NewMemFile(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memFile.ReadAt(data, 0)
	}
}

func BenchmarkMemFileWrite(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memFile := NewMemFile(nil)
		memFile.Write(data)
	}
}

func BenchmarkMemFileConcurrent(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	memFile := NewMemFile(data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		buf := make([]byte, 1024)
		for pb.Next() {
			memFile.ReadAt(buf, 0)
			memFile.Write([]byte("test"))
		}
	})
}

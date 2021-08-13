package utils

import (
	"errors"
	"io"
	"sync"
)

var errInvalid = errors.New("invalid argument")

type MemFile struct {
	m sync.Mutex
	b []byte
	i int
}

func NewMemFile(b []byte) *MemFile {
	return &MemFile{b: b}
}

func (fb *MemFile) Read(b []byte) (int, error) {
	fb.m.Lock()
	defer fb.m.Unlock()

	n, err := fb.readAt(b, int64(fb.i))
	fb.i += n
	return n, err
}

func (fb *MemFile) ReadAt(b []byte, offset int64) (int, error) {
	fb.m.Lock()
	defer fb.m.Unlock()
	return fb.readAt(b, offset)
}

func (fb *MemFile) readAt(b []byte, off int64) (int, error) {
	if off < 0 || int64(int(off)) < off {
		return 0, errInvalid
	}
	if off > int64(len(fb.b)) {
		return 0, io.EOF
	}
	n := copy(b, fb.b[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (fb *MemFile) Write(b []byte) (int, error) {
	fb.m.Lock()
	defer fb.m.Unlock()

	n, err := fb.writeAt(b, int64(fb.i))
	fb.i += n
	return n, err
}

func (fb *MemFile) WriteAt(b []byte, offset int64) (int, error) {
	fb.m.Lock()
	defer fb.m.Unlock()
	return fb.writeAt(b, offset)
}

func (fb *MemFile) writeAt(b []byte, off int64) (int, error) {
	if off < 0 || int64(int(off)) < off {
		return 0, errInvalid
	}
	if off > int64(len(fb.b)) {
		fb.truncate(off)
	}
	n := copy(fb.b[off:], b)
	fb.b = append(fb.b, b[n:]...)
	return len(b), nil
}

func (fb *MemFile) Seek(offset int64, whence int) (int64, error) {
	fb.m.Lock()
	defer fb.m.Unlock()

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(fb.i) + offset
	case io.SeekEnd:
		abs = int64(len(fb.b)) + offset
	default:
		return 0, errInvalid
	}
	if abs < 0 {
		return 0, errInvalid
	}
	fb.i = int(abs)
	return abs, nil
}

func (fb *MemFile) Truncate(n int64) error {
	fb.m.Lock()
	defer fb.m.Unlock()
	return fb.truncate(n)
}

func (fb *MemFile) truncate(n int64) error {
	switch {
	case n < 0 || int64(int(n)) < n:
		return errInvalid
	case n <= int64(len(fb.b)):
		fb.b = fb.b[:n]
		return nil
	default:
		fb.b = append(fb.b, make([]byte, int(n)-len(fb.b))...)
		return nil
	}
}

func (fb *MemFile) Bytes() []byte {
	fb.m.Lock()
	defer fb.m.Unlock()
	return fb.b
}

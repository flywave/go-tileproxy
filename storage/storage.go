package storage

import (
	"io"
)

type ReadCloseSeeker interface {
	io.ReadCloser
	io.Seeker
}

type Storage interface {
	TestConnection() error

	Reader(path string) (ReadCloseSeeker, error)
	ReadFile(path string) ([]byte, error)
	FileExists(path string) (bool, error)
	CopyFile(oldPath, newPath string) error
	MoveFile(oldPath, newPath string) error
	WriteFile(fr io.Reader, path string) (int64, error)
	RemoveFile(path string) error

	ListDirectory(path string) (*[]string, error)
	RemoveDirectory(path string) error
}

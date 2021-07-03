package storage

import (
	"io"
)

type Storage interface {
	ReadFile(path string) ([]byte, error)
	FileExists(path string) (bool, error)
	CopyFile(oldPath, newPath string) error
	MoveFile(oldPath, newPath string) error
	WriteFile(fr io.Reader, path string) (int64, error)
	RemoveFile(path string) error

	ListDirectory(path string) (*[]string, error)
	RemoveDirectory(path string) error
}

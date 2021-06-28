package storage

import (
	"errors"
	"io"
)

const (
	STORAGE_DRIVER_LOCAL = "local"
	STORAGE_DRIVER_S3    = "amazons3"
)

type StorageSettings struct {
	EnableFileAttachments   *bool
	EnableMobileUpload      *bool
	EnableMobileDownload    *bool
	MaxFileSize             *int64
	DriverName              *string `restricted:"true"`
	Directory               *string `restricted:"true"`
	EnablePublicLink        *bool
	PublicLinkSalt          *string
	InitialFont             *string
	AmazonS3AccessKeyId     *string `restricted:"true"`
	AmazonS3SecretAccessKey *string `restricted:"true"`
	AmazonS3Bucket          *string `restricted:"true"`
	AmazonS3Region          *string `restricted:"true"`
	AmazonS3Endpoint        *string `restricted:"true"`
	AmazonS3SSL             *bool   `restricted:"true"`
	AmazonS3SignV2          *bool   `restricted:"true"`
	AmazonS3SSE             *bool   `restricted:"true"`
	AmazonS3Trace           *bool   `restricted:"true"`
}

func (s *StorageSettings) SetDefaults(isUpdate bool) {

}

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

func NewStorage(settings *StorageSettings) (Storage, error) {
	switch *settings.DriverName {
	case STORAGE_DRIVER_S3:
		return &S3Storage{
			endpoint:  *settings.AmazonS3Endpoint,
			accessKey: *settings.AmazonS3AccessKeyId,
			secretKey: *settings.AmazonS3SecretAccessKey,
			secure:    settings.AmazonS3SSL == nil || *settings.AmazonS3SSL,
			signV2:    settings.AmazonS3SignV2 != nil && *settings.AmazonS3SignV2,
			region:    *settings.AmazonS3Region,
			bucket:    *settings.AmazonS3Bucket,
			encrypt:   settings.AmazonS3SSE != nil && *settings.AmazonS3SSE,
			trace:     settings.AmazonS3Trace != nil && *settings.AmazonS3Trace,
		}, nil
	case STORAGE_DRIVER_LOCAL:
		return &LocalStorage{
			directory: *settings.Directory,
		}, nil
	}
	return nil, errors.New("storage no driver!")
}

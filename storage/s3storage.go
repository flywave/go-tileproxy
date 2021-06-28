package storage

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flywave/go-tileproxy/tile"
	s3 "github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/minio/minio-go/v6/pkg/encrypt"
)

type S3Storage struct {
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	signV2    bool
	region    string
	bucket    string
	encrypt   bool
	trace     bool
}

func (b *S3Storage) s3New() (*s3.Client, error) {
	var creds *credentials.Credentials

	if b.accessKey == "" && b.secretKey == "" {
		creds = credentials.NewIAM("")
	} else if b.signV2 {
		creds = credentials.NewStatic(b.accessKey, b.secretKey, "", credentials.SignatureV2)
	} else {
		creds = credentials.NewStatic(b.accessKey, b.secretKey, "", credentials.SignatureV4)
	}

	s3Clnt, err := s3.NewWithCredentials(b.endpoint, creds, b.secure, b.region)
	if err != nil {
		return nil, err
	}

	if b.trace {
		s3Clnt.TraceOn(os.Stdout)
	}

	return s3Clnt, nil
}

func (b *S3Storage) TestConnection() error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("TestFileConnection s3 connection error")
	}

	exists, err := s3Clnt.BucketExists(b.bucket)
	if err != nil {
		return errors.New("TestFileConnection s3 bucket  exists")
	}

	if !exists {
		err := s3Clnt.MakeBucket(b.bucket, b.region)
		if err != nil {
			return errors.New("TestFileConnection bucked create error")
		}
	}
	return nil
}

func (b *S3Storage) Reader(path string) (ReadCloseSeeker, error) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, errors.New("Reader error")
	}
	minioObject, err := s3Clnt.GetObject(b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, errors.New("Reader error")
	}
	return minioObject, nil
}

func (b *S3Storage) ReadFile(path string) ([]byte, error) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, errors.New("ReadFile error")
	}
	minioObject, err := s3Clnt.GetObject(b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, errors.New("ReadFile error")
	}
	defer minioObject.Close()
	if f, err := ioutil.ReadAll(minioObject); err != nil {
		return nil, errors.New("ReadFile error")
	} else {
		return f, nil
	}
}

func (b *S3Storage) FileExists(path string) (bool, error) {
	s3Clnt, err := b.s3New()

	if err != nil {
		return false, errors.New("FileExists error")
	}
	_, err = s3Clnt.StatObject(b.bucket, path, s3.StatObjectOptions{})

	if err == nil {
		return true, nil
	}

	if err.(s3.ErrorResponse).Code == "NoSuchKey" {
		return false, nil
	}

	return false, errors.New("FileExists error")
}

func (b *S3Storage) CopyFile(oldPath, newPath string) error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("copyFile error")
	}

	source := s3.NewSourceInfo(b.bucket, oldPath, nil)
	destination, err := s3.NewDestinationInfo(b.bucket, newPath, encrypt.NewSSE(), nil)
	if err != nil {
		return errors.New("copyFile error")
	}
	if err = s3Clnt.CopyObject(destination, source); err != nil {
		return errors.New("copyFile error")
	}
	return nil
}

func (b *S3Storage) MoveFile(oldPath, newPath string) error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("moveFile error")
	}

	source := s3.NewSourceInfo(b.bucket, oldPath, nil)
	destination, err := s3.NewDestinationInfo(b.bucket, newPath, encrypt.NewSSE(), nil)
	if err != nil {
		return errors.New("moveFile error")
	}
	if err = s3Clnt.CopyObject(destination, source); err != nil {
		return errors.New("moveFile error")
	}
	if err = s3Clnt.RemoveObject(b.bucket, oldPath); err != nil {
		return errors.New("moveFile error")
	}
	return nil
}

func (b *S3Storage) WriteFile(fr io.Reader, path string) (int64, error) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return 0, errors.New("WriteFile error")
	}

	var contentType string
	if ext := filepath.Ext(path); tile.IsFileExtTile(ext) {
		contentType, _ = tile.GetMimetype(ext)
	} else {
		contentType = "binary/octet-stream"
	}

	options := s3PutOptions(b.encrypt, contentType)
	var buf bytes.Buffer
	_, err = buf.ReadFrom(fr)
	if err != nil {
		return 0, errors.New("WriteFile error")
	}
	written, err := s3Clnt.PutObject(b.bucket, path, &buf, int64(buf.Len()), options)
	if err != nil {
		return written, errors.New("WriteFile error")
	}

	return written, nil
}

func (b *S3Storage) RemoveFile(path string) error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("RemoveFile error")
	}

	if err := s3Clnt.RemoveObject(b.bucket, path); err != nil {
		return errors.New("RemoveFile error")
	}

	return nil
}

func getPathsFromObjectInfos(in <-chan s3.ObjectInfo) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer close(out)

		for {
			info, done := <-in

			if !done {
				break
			}

			out <- info.Key
		}
	}()

	return out
}

func (b *S3Storage) ListDirectory(path string) (*[]string, error) {
	var paths []string

	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, errors.New("ListDirectory error")
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	if !strings.HasSuffix(path, "/") && len(path) > 0 {
		path = path + "/"
	}
	for object := range s3Clnt.ListObjects(b.bucket, path, false, doneCh) {
		if object.Err != nil {
			return nil, errors.New("ListDirectory error")
		}
		paths = append(paths, strings.Trim(object.Key, "/"))
	}

	return &paths, nil
}

func (b *S3Storage) RemoveDirectory(path string) error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("RemoveDirectory error")
	}

	doneCh := make(chan struct{})

	for err := range s3Clnt.RemoveObjects(b.bucket, getPathsFromObjectInfos(s3Clnt.ListObjects(b.bucket, path, true, doneCh))) {
		if err.Err != nil {
			doneCh <- struct{}{}
			return errors.New("RemoveDirectory error")
		}
	}

	close(doneCh)
	return nil
}

func s3PutOptions(encrypted bool, contentType string) s3.PutObjectOptions {
	options := s3.PutObjectOptions{}
	if encrypted {
		options.ServerSideEncryption = encrypt.NewSSE()
	}
	options.ContentType = contentType

	return options
}

func CheckMandatoryS3Fields(settings *StorageSettings) error {
	if settings.AmazonS3Bucket == nil || len(*settings.AmazonS3Bucket) == 0 {
		return errors.New("S3File missing s3 bucket error")
	}

	if settings.AmazonS3Endpoint == nil || len(*settings.AmazonS3Endpoint) == 0 {
		settings.SetDefaults(true)
	}

	return nil
}

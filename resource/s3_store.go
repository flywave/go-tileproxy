package resource

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	s3 "github.com/minio/minio-go/v6"

	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/minio/minio-go/v6/pkg/encrypt"

	"github.com/flywave/go-tileproxy/tile"
)

type S3Options struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	SignV2    bool
	Region    string
	Bucket    string
	Encrypt   bool
	Trace     bool
}

type S3Store struct {
	Store
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	signV2    bool
	region    string
	bucket    string
	encrypt   bool
	trace     bool
	CacheDir  string
}

func NewS3Store(cache_dir string, setting S3Options) *S3Store {
	c := &S3Store{CacheDir: cache_dir}
	c.endpoint = setting.Endpoint
	c.accessKey = setting.AccessKey
	c.secretKey = setting.SecretKey
	c.secure = setting.Secure
	c.signV2 = setting.SignV2
	c.region = setting.Region
	c.bucket = setting.Bucket
	c.encrypt = setting.Encrypt
	c.trace = setting.Trace
	return c
}

func (b *S3Store) s3New() (*s3.Client, error) {
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

func (b *S3Store) TestConnection() error {
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("TestFileConnection")
	}

	exists, err := s3Clnt.BucketExists(b.bucket)
	if err != nil {
		return errors.New("TestFileConnection")
	}

	if !exists {
		err := s3Clnt.MakeBucket(b.bucket, b.region)
		if err != nil {
			return errors.New("TestFileConnection")
		}
	}
	return nil
}

type readCloseSeeker interface {
	io.ReadCloser
	io.Seeker
}

func (b *S3Store) reader(path string) (readCloseSeeker, error) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, errors.New("Reader")
	}
	minioObject, err := s3Clnt.GetObject(b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, errors.New("Reader")
	}
	return minioObject, nil
}

func (b *S3Store) writeFile(fr io.Reader, path string) (int64, error) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return 0, errors.New("WriteFile")
	}

	format := tile.TileFormat(filepath.Ext(path))

	contentType := format.MimeType()

	options := s3PutOptions(b.encrypt, contentType)
	var buf bytes.Buffer
	_, err = buf.ReadFrom(fr)
	if err != nil {
		return 0, errors.New("WriteFile")
	}
	written, err := s3Clnt.PutObject(b.bucket, path, &buf, int64(buf.Len()), options)
	if err != nil {
		return written, errors.New("WriteFile")
	}

	return written, nil
}

func s3PutOptions(encrypted bool, contentType string) s3.PutObjectOptions {
	options := s3.PutObjectOptions{}
	if encrypted {
		options.ServerSideEncryption = encrypt.NewSSE()
	}
	options.ContentType = contentType

	return options
}

func (b *S3Store) fileExists(path string) (bool, error) {
	s3Clnt, err := b.s3New()

	if err != nil {
		return false, errors.New("FileExists")
	}
	_, err = s3Clnt.StatObject(b.bucket, path, s3.StatObjectOptions{})

	if err == nil {
		return true, nil
	}

	if err.(s3.ErrorResponse).Code == "NoSuchKey" {
		return false, nil
	}

	return false, errors.New("FileExists")
}

func (c *S3Store) Save(r Resource) error {
	if r.IsStored() {
		return nil
	}

	if r.GetLocation() == "" {
		hash := r.Hash()
		r.SetLocation(path.Join(c.CacheDir, base64.RawURLEncoding.EncodeToString(hash)) + "." + r.GetExtension())
	}

	data := r.GetData()

	reader := bytes.NewBuffer(data)

	if _, err := c.writeFile(reader, r.GetLocation()); err != nil {
		return err
	}

	r.SetStored()

	return nil
}

func (c *S3Store) Load(r Resource) error {
	hash := r.Hash()
	r.SetLocation(path.Join(c.CacheDir, base64.RawURLEncoding.EncodeToString(hash)) + "." + r.GetExtension())

	if ok, _ := c.fileExists(r.GetLocation()); ok {
		if reader, err := c.reader(r.GetLocation()); err == nil {
			bufs, e := ioutil.ReadAll(reader)
			if e != nil {
				return e
			}
			r.SetData(bufs)
		} else {
			return err
		}
		return nil
	}

	return errors.New("res not found!")
}

package cache

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	s3 "github.com/minio/minio-go/v6"

	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/minio/minio-go/v6/pkg/encrypt"

	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type S3Cache struct {
	Cache
	endpoint      string
	accessKey     string
	secretKey     string
	secure        bool
	signV2        bool
	region        string
	bucket        string
	encrypt       bool
	trace         bool
	cacheDir      string
	fileExt       string
	tileLocation  func(*Tile, string, string, bool) string
	levelLocation func(int, string) string
	creater       SourceCreater
}

func (b *S3Cache) s3New() (*s3.Client, error) {
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

func (b *S3Cache) TestConnection() error {
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

func (b *S3Cache) reader(path string) (readCloseSeeker, error) {
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

func (b *S3Cache) writeFile(fr io.Reader, path string) (int64, error) {
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

func (b *S3Cache) TileLocation(tile *Tile, create_dir bool) string {
	return b.tileLocation(tile, b.cacheDir, b.fileExt, create_dir)
}

func (b *S3Cache) LevelLocation(level int) string {
	return b.levelLocation(level, b.cacheDir)
}

func (b *S3Cache) fileExists(path string) (bool, error) {
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

func (b *S3Cache) LoadTile(tile *Tile, withMetadata bool) error {
	if !tile.IsMissing() {
		return nil
	}

	location := b.TileLocation(tile, false)

	if ok, _ := b.fileExists(location); ok {
		if withMetadata {
			b.LoadTileMetadata(tile)
		}
		reader, _ := b.reader(location)
		data, _ := ioutil.ReadAll(reader)
		tile.Source = b.creater(data, location)
		return nil
	}
	return errors.New("not found")
}

func (b *S3Cache) LoadTiles(tiles *TileCollection, withMetadata bool) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := b.LoadTile(tile, withMetadata); err != nil {
			errs = err
		}
	}
	return errs
}

func (b *S3Cache) StoreTile(tile *Tile) error {
	if tile.Stored {
		return nil
	}
	tile_loc := b.TileLocation(tile, true)
	return b.store(tile, tile_loc)
}

func (b *S3Cache) store(tile *Tile, location string) error {
	if ok, _ := utils.IsSymlink(location); ok {
		os.Remove(location)
	}
	data := tile.Source.GetBuffer(nil, nil)
	reader := bytes.NewBuffer(data)
	_, err := b.writeFile(reader, location)
	return err
}

func (b *S3Cache) StoreTiles(tiles *TileCollection) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := b.StoreTile(tile); err != nil {
			errs = err
		}
	}
	return errs
}

func (b *S3Cache) RemoveTile(tile *Tile) error {
	location := b.TileLocation(tile, false)
	s3Clnt, err := b.s3New()
	if err != nil {
		return errors.New("RemoveFile")
	}

	if err := s3Clnt.RemoveObject(b.bucket, location); err != nil {
		return errors.New("RemoveFile")
	}

	return nil
}

func (b *S3Cache) RemoveTiles(tiles *TileCollection) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := b.RemoveTile(tile); err != nil {
			errs = err
		}
	}
	return errs
}

func (b *S3Cache) IsCached(tile *Tile) bool {
	if tile.IsMissing() {
		location := b.TileLocation(tile, false)
		if ok, _ := b.fileExists(location); ok {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func (b *S3Cache) LoadTileMetadata(tile *Tile) error {
	location := b.TileLocation(tile, false)

	s3Clnt, err := b.s3New()

	if err != nil {
		return err
	}

	stats, err := s3Clnt.StatObject(b.bucket, location, s3.StatObjectOptions{})

	if err != nil {
		return err
	}

	tile.Timestamp = stats.LastModified
	tile.Size = stats.Size
	return nil
}

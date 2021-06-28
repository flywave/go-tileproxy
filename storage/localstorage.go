package storage

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	if err = os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return
	}
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	stat, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, stat.Mode())
	if err != nil {
		return
	}

	return
}

const (
	TEST_FILE_PATH = "/testfile"
)

type LocalStorage struct {
	directory string
}

func (b *LocalStorage) TestConnection() error {
	f := bytes.NewReader([]byte("testingwrite"))
	if _, err := writeFileLocally(f, filepath.Join(b.directory, TEST_FILE_PATH)); err != nil {
		return errors.New("TestFileConnection test connection error!")
	}
	os.Remove(filepath.Join(b.directory, TEST_FILE_PATH))
	return nil
}

func (b *LocalStorage) Reader(path string) (ReadCloseSeeker, error) {
	f, err := os.Open(filepath.Join(b.directory, path))
	if err != nil {
		return nil, errors.New("Reader reading local error")
	}
	return f, nil
}

func (b *LocalStorage) ReadFile(path string) ([]byte, error) {
	f, err := ioutil.ReadFile(filepath.Join(b.directory, path))
	if err != nil {
		return nil, errors.New("ReadFile reading local error")
	}
	return f, nil
}

func (b *LocalStorage) FileExists(path string) (bool, error) {
	_, err := os.Stat(filepath.Join(b.directory, path))

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, errors.New("ReadFile file exists")
	}
	return true, nil
}

func (b *LocalStorage) CopyFile(oldPath, newPath string) error {
	if err := copyFile(filepath.Join(b.directory, oldPath), filepath.Join(b.directory, newPath)); err != nil {
		return errors.New("copyFile rename error")
	}
	return nil
}

func (b *LocalStorage) MoveFile(oldPath, newPath string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(b.directory, newPath)), 0774); err != nil {
		return errors.New("moveFile rename error")
	}

	if err := os.Rename(filepath.Join(b.directory, oldPath), filepath.Join(b.directory, newPath)); err != nil {
		return errors.New("moveFile move file")
	}

	return nil
}

func (b *LocalStorage) WriteFile(fr io.Reader, path string) (int64, error) {
	return writeFileLocally(fr, filepath.Join(b.directory, path))
}

func writeFileLocally(fr io.Reader, path string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0774); err != nil {
		return 0, errors.New("WriteFile create dir error")
	}
	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, errors.New("WriteFile writing error")
	}
	defer fw.Close()
	written, err := io.Copy(fw, fr)
	if err != nil {
		return written, errors.New("WriteFile writing error")
	}
	return written, nil
}

func (b *LocalStorage) RemoveFile(path string) error {
	if err := os.Remove(filepath.Join(b.directory, path)); err != nil {
		return errors.New("RemoveFile remove file error")
	}
	return nil
}

func (b *LocalStorage) ListDirectory(path string) (*[]string, error) {
	var paths []string
	fileInfos, err := ioutil.ReadDir(filepath.Join(b.directory, path))
	if err != nil {
		if os.IsNotExist(err) {
			return &paths, nil
		}
		return nil, errors.New("ListDirectory list directory error")
	}
	for _, fileInfo := range fileInfos {
		paths = append(paths, filepath.Join(path, fileInfo.Name()))
	}
	return &paths, nil
}

func (b *LocalStorage) RemoveDirectory(path string) error {
	if err := os.RemoveAll(filepath.Join(b.directory, path)); err != nil {
		return errors.New("RemoveDirectory error")
	}
	return nil
}

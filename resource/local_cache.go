package resource

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type LocalStore struct {
	Store
	CacheDir string
	FileExt  string
}

func fileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (c *LocalStore) Save(r Resource) error {
	if r.IsStored() {
		return nil
	}

	if r.GetLocation() == "" {
		hash := r.Hash()
		r.SetLocation(path.Join(c.CacheDir, string(hash)) + "." + c.FileExt)
	}

	data := r.GetData()

	if err := ioutil.WriteFile(r.GetLocation(), data, 0644); err != nil {
		return err
	}

	r.SetStored()

	return nil
}

func (c *LocalStore) Load(r Resource) error {
	hash := r.Hash()
	r.SetLocation(path.Join(c.CacheDir, string(hash)) + "." + c.FileExt)

	if ok, _ := fileExists(r.GetLocation()); ok {
		if f, err := os.Open(r.GetLocation()); err == nil {
			bufs, e := ioutil.ReadAll(f)
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

func NewLocalStore(cache_dir string, file_ext string) *LocalStore {
	return &LocalStore{CacheDir: cache_dir, FileExt: file_ext}
}

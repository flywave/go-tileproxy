package resource

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/flywave/go-tileproxy/utils"
)

type LocalStore struct {
	Store
	CacheDir string
}

func (c *LocalStore) Save(r Resource) error {
	if r.IsStored() {
		return nil
	}

	if r.GetLocation() == "" {
		r.SetLocation(path.Join(c.CacheDir, r.GetFileName()+"."+r.GetExtension()))
	}

	data := r.GetData()

	if err := ioutil.WriteFile(r.GetLocation(), data, 0777); err != nil {
		return err
	}

	r.SetStored()

	return nil
}

func (c *LocalStore) Load(r Resource) error {
	r.SetLocation(path.Join(c.CacheDir, r.GetFileName()+"."+r.GetExtension()))

	if ok := utils.FileExists(r.GetLocation()); ok {
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

	return errors.New("res not found")
}

func NewLocalStore(cache_dir string) *LocalStore {
	if !utils.FileExists(cache_dir) {
		os.MkdirAll(cache_dir, 0777)
	}
	return &LocalStore{CacheDir: cache_dir}
}

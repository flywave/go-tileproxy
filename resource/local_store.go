package resource

import (
	"errors"
	"io"
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

	if err := os.WriteFile(r.GetLocation(), data, 0644); err != nil {
		return err
	}

	r.SetStored()

	return nil
}

func (c *LocalStore) Load(r Resource) error {
	r.SetLocation(path.Join(c.CacheDir, r.GetFileName()+"."+r.GetExtension()))

	if ok := utils.FileExists(r.GetLocation()); ok {
		if f, err := os.Open(r.GetLocation()); err == nil {
			defer f.Close()
			bufs, e := io.ReadAll(f)
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
		os.MkdirAll(cache_dir, 0755)
	}
	return &LocalStore{CacheDir: cache_dir}
}

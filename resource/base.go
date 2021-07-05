package resource

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type Cache interface {
	Store(r Resource) error
	Load(r Resource) error
}

type LocalCache struct {
	Cache
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

func (c *LocalCache) Store(r Resource) error {
	if r.stored() {
		return nil
	}

	if r.location() == "" {
		hash := r.Hash()
		r.set_location(path.Join(c.CacheDir, string(hash)) + "." + c.FileExt)
	}

	data := r.GetData()

	if err := ioutil.WriteFile(r.location(), data, 0644); err != nil {
		return err
	}

	r.set_stored()

	return nil
}

func (c *LocalCache) Load(r Resource) error {
	hash := r.Hash()
	r.set_location(path.Join(c.CacheDir, string(hash)) + "." + c.FileExt)

	if ok, _ := fileExists(r.location()); ok {
		if f, err := os.Open(r.location()); err == nil {
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

func NewLocalCache(cache_dir string, file_ext string) *LocalCache {
	return &LocalCache{CacheDir: cache_dir, FileExt: file_ext}
}

type Resource interface {
	Hash() []byte
	GetData() []byte
	SetData([]byte)
	stored() bool
	set_stored()
	location() string
	set_location(l string)
}

type BaseResource struct {
	Resource
	Stored   bool
	Location string
	ID       string
}

func (r *BaseResource) stored() bool {
	return r.Stored
}

func (r *BaseResource) set_stored() {
	r.Stored = true
}

func (r *BaseResource) location() string {
	return r.Location
}

func (r *BaseResource) set_location(l string) {
	r.Location = l
}

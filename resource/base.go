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

func (c *LocalCache) Load(r Resource) error {
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

func NewLocalCache(cache_dir string, file_ext string) *LocalCache {
	return &LocalCache{CacheDir: cache_dir, FileExt: file_ext}
}

type Resource interface {
	Hash() []byte
	GetData() []byte
	SetData([]byte)
	IsStored() bool
	SetStored()
	GetLocation() string
	SetLocation(l string)
	GetID() string
}

type BaseResource struct {
	Resource
	Stored   bool
	Location string
	ID       string
}

func (r *BaseResource) IsStored() bool {
	return r.Stored
}

func (r *BaseResource) SetStored() {
	r.Stored = true
}

func (r *BaseResource) GetLocation() string {
	return r.Location
}

func (r *BaseResource) SetLocation(l string) {
	r.Location = l
}

func (r *BaseResource) GetID() string {
	return r.ID
}

func (r *BaseResource) SetID(id string) {
	r.ID = id
}

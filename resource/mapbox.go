package resource

import "github.com/flywave/go-tileproxy/images"

type StyleCache struct {
	Cache
	CacheDir string
	FileExt  string
}

func (c *StyleCache) Store(r Resource) error {
	return nil
}

func (c *StyleCache) Load(r Resource) error {
	return nil
}

func NewStyleCache(cache_dir string, file_ext string) *StyleCache {
	return &StyleCache{CacheDir: cache_dir, FileExt: file_ext}
}

type Style struct {
	Resource
	Stored   bool
	Location string
	ID       uint64
	Scale    int
	sprites  []Sprite
	glyphs   []Glyphs
}

type Sprite struct {
	Resource
	Source   images.Source
	Stored   bool
	Location string
	ID       uint64
	Scale    int
}

type Glyphs struct {
	Resource
	FontName string
	Stored   bool
	Location string
	Buffer   []byte
	ID       uint64
}

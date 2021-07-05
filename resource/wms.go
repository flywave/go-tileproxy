package resource

import (
	"github.com/flywave/go-tileproxy/images"
)

type LegendCache struct {
	Cache
	CacheDir string
	FileExt  string
}

func (c *LegendCache) Store(r Resource) error {
	return nil
}

func (c *LegendCache) Load(r Resource) error {
	return nil
}

func NewLegendCache(cache_dir string, file_ext string) *LegendCache {
	return &LegendCache{CacheDir: cache_dir, FileExt: file_ext}
}

type Legend struct {
	Resource
	Source   images.Source
	Stored   bool
	Location string
	ID       uint64
	Scale    float64
}

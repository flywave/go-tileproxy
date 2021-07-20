package cache

import (
	"time"

	"github.com/flywave/go-tileproxy/images"
)

type CacheInfo struct {
	Cacheable bool
	Timestamp time.Time
	Size      [2]uint32
}

type Tile struct {
	Coord     [3]int
	Source    images.Source
	Location  string
	Stored    bool
	Cacheable bool
	Size      [2]uint32
	Timestamp time.Time
}

func NewTile(coord [3]int) *Tile {
	return &Tile{Coord: coord}
}

func (t *Tile) GetCacheInfo() CacheInfo {
	return CacheInfo{Cacheable: t.Cacheable, Timestamp: t.Timestamp,
		Size: t.Size}
}

func (t *Tile) SetCacheInfo(cache CacheInfo) {
	t.Cacheable = cache.Cacheable
	t.Timestamp = cache.Timestamp
	t.Size = cache.Size
}

func (t *Tile) GetSource() images.Source {
	return t.Source
}

func (t *Tile) IsMissing() bool {
	return (t.Source == nil)
}

func (t *Tile) EQ(o *Tile) bool {
	if t.Coord[0] != o.Coord[0] || t.Coord[1] != o.Coord[1] || t.Coord[2] != o.Coord[2] {
		return false
	}
	return true
}

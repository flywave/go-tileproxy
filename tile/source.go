package tile

import "time"

type TileType uint8

const (
	TILE_IMAGERY   = 0
	TILE_DEMRASTER = 1
	TILE_VECTOR    = 2
)

type CacheInfo struct {
	Cacheable bool
	Timestamp time.Time
	Size      int64
}

type Source interface {
	GetType() TileType
	GetSource() interface{}
	SetSource(src interface{})
	GetFileName() string
	GetSize() [2]uint32
	GetBuffer(format *TileFormat, in_tile_opts TileOptions) []byte
	GetTile() interface{}
	GetCacheable() *CacheInfo
	SetCacheable(c *CacheInfo)
	SetTileOptions(options TileOptions)
	GetTileOptions() TileOptions
}

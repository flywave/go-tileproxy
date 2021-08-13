package tile

import (
	"time"

	"github.com/flywave/go-tileproxy/geo"
)

type TileType uint8

const (
	TILE_IMAGERY = 0
	TILE_DEM     = 1
	TILE_VECTOR  = 2
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
	GetGeoReference() *geo.GeoReference
}

type DummyTileSource struct {
	Source
	Data      string
	Cacheable *CacheInfo
	Opts      TileOptions
	Georef    *geo.GeoReference
}

func (s *DummyTileSource) GetType() TileType {
	return TILE_IMAGERY
}

func (s *DummyTileSource) GetSource() interface{} {
	return s.Data
}

func (s *DummyTileSource) SetSource(src interface{}) {
	switch v := src.(type) {
	case string:
		s.Data = v
	case []byte:
		s.Data = string(v)
	}
}

func (s *DummyTileSource) GetFileName() string {
	return "dummy"
}

func (s *DummyTileSource) GetSize() [2]uint32 {
	return [2]uint32{256, 256}
}

func (s *DummyTileSource) GetBuffer(format *TileFormat, in_tile_opts TileOptions) []byte {
	return []byte(s.Data)
}

func (s *DummyTileSource) GetTile() interface{} {
	return s.Data
}

func (s *DummyTileSource) GetCacheable() *CacheInfo {
	return s.Cacheable
}

func (s *DummyTileSource) SetCacheable(c *CacheInfo) {
	s.Cacheable = c
}

func (s *DummyTileSource) SetTileOptions(options TileOptions) {
	s.Opts = options
}

func (s *DummyTileSource) GetTileOptions() TileOptions {
	return s.Opts
}

func (s *DummyTileSource) GetGeoReference() *geo.GeoReference {
	return s.Georef
}

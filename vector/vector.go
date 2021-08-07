package vector

import (
	"io"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type VectorSource struct {
	tile.Source
	data       interface{}
	buf        []byte
	fname      string
	Options    tile.TileOptions
	tile       [3]int
	cacheable  *tile.CacheInfo
	georef     *geo.GeoReference
	decodeFunc func(r io.Reader) (interface{}, error)
	encodeFunc func(data interface{}) ([]byte, error)
}

func (s *VectorSource) GetType() tile.TileType {
	return tile.TILE_VECTOR
}

func (s *VectorSource) GetCacheable() *tile.CacheInfo {
	if s.cacheable == nil {
		s.cacheable = &tile.CacheInfo{Cacheable: false}
	}
	return s.cacheable
}

func (s *VectorSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *VectorSource) GetFileName() string {
	return s.fname
}

func (s *VectorSource) GetSize() [2]uint32 {
	return [2]uint32{4096, 4096}
}

func (s *VectorSource) GetSource() interface{} {
	if s.data != nil {
		return s.data
	} else if len(s.fname) > 0 {
		return s.fname
	}
	return nil
}

func (s *VectorSource) SetSource(src interface{}) {
	s.data = nil
	s.buf = nil
	switch ss := src.(type) {
	case io.Reader:
		s.data, _ = s.decode(ss)
	case string:
		s.fname = ss
	default:
		s.data = ss
	}
}

func (s *VectorSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	return nil
}

func (s *VectorSource) GetTile() interface{} {
	return s.data
}

func (s *VectorSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *VectorSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func (s *VectorSource) decode(r io.Reader) (interface{}, error) {
	return s.decodeFunc(r)
}

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
	size       []uint32
	bounds     []float64
	cacheable  *tile.CacheInfo
	georef     *geo.GeoReference
	decodeFunc func(r io.Reader) (interface{}, error)
}

func (s *VectorSource) GetType() tile.TileType {
	return tile.TILE_VECTOR
}

func (s *VectorSource) GetCacheable() *tile.CacheInfo {
	return s.cacheable
}

func (s *VectorSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *VectorSource) GetFileName() string {
	return s.fname
}

func (s *VectorSource) GetSize() [2]uint32 {
	if s.size == nil {
		s.size = make([]uint32, 2)
	}
	return [2]uint32{s.size[0], s.size[1]}
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
	return nil
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

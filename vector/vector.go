package vector

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type VectorOptions struct {
	tile.TileOptions
	Format      tile.TileFormat
	Tolerance   float64
	Extent      uint16
	Buffer      uint16
	LineMetrics bool
	MaxZoom     uint8
}

func (s *VectorOptions) GetFormat() tile.TileFormat {
	return s.Format
}

type VectorSource struct {
	tile.Source
	data      interface{}
	buf       []byte
	fname     string
	Options   tile.TileOptions
	tile      [3]int
	cacheable *tile.CacheInfo
	georef    *geo.GeoReference
	io        VectorIO
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
	opt := s.Options.(*VectorOptions)
	return [2]uint32{uint32(opt.Extent), uint32(opt.Extent)}
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
	if s.buf == nil {
		var err error
		s.buf, err = s.io.Encode(s.data)
		if err != nil {
			return nil
		}
	}
	return s.buf
}

func (s *VectorSource) GetTile() interface{} {
	if s.data == nil {
		if s.buf == nil {
			f, err := os.Open(s.fname)
			if err != nil {
				return nil
			}
			s.buf, err = ioutil.ReadAll(f)
			if err != nil {
				return nil
			}
		}
		r := bytes.NewBuffer(s.buf)
		var err error
		s.data, err = s.io.Decode(r)
		if err != nil {
			return nil
		}
	}
	return s.data
}

func (s *VectorSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *VectorSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func (s *VectorSource) decode(r io.Reader) (interface{}, error) {
	return s.io.Decode(r)
}

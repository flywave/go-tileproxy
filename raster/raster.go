package raster

import (
	"io"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type RasterType uint32

const (
	RT_CHAR   RasterType = 0
	RT_UCHAR  RasterType = 1
	RT_SHORT  RasterType = 2
	RT_USHORT RasterType = 3
	RT_INT    RasterType = 4
	RT_UINT   RasterType = 5
	RT_FLOAT  RasterType = 6
	RT_DOUBLE RasterType = 7
)

type RasterSource struct {
	tile.Source
	data       interface{}
	tp         RasterType
	buf        []byte
	fname      string
	size       []uint32
	minimum    float64
	maximum    float64
	bounds     vec2d.Rect
	cacheable  bool
	georef     *geo.GeoReference
	nodata     float64
	Options    tile.TileOptions
	decodeFunc func(r io.Reader) (interface{}, error)
}

func (s *RasterSource) GetType() tile.TileType {
	return tile.TILE_DEMRASTER
}

func (s *RasterSource) GetCacheable() bool {
	return s.cacheable
}

func (s *RasterSource) SetCacheable(c bool) {
	s.cacheable = c
}

func (s *RasterSource) GetFileName() string {
	return s.fname
}

func (s *RasterSource) GetSize() [2]uint32 {
	if s.size == nil {
		s.size = make([]uint32, 2)
	}
	return [2]uint32{s.size[0], s.size[1]}
}

func (s *RasterSource) GetSource() interface{} {
	if s.data != nil {
		return s.data
	} else if len(s.fname) > 0 {
		return s.fname
	}
	return nil
}

func (s *RasterSource) SetSource(src interface{}) {
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

func (s *RasterSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *RasterSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func (s *RasterSource) decode(r io.Reader) (interface{}, error) {
	return s.decodeFunc(r)
}

func (s *RasterSource) MinimumValue() float64 {
	return s.minimum
}

func (s *RasterSource) MaximumValue() float64 {
	return s.maximum
}

func (s *RasterSource) RangeValue() float64 {
	return s.maximum - s.minimum
}

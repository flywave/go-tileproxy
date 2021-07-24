package raster

import (
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

type Raster struct {
	tile.Source
	data      interface{}
	tp        RasterType
	buf       []byte
	fname     string
	size      []uint32
	bounds    []float64
	cacheable bool
	georef    *geo.GeoReference
	nodata    float64
}

func (s *Raster) GetCacheable() bool {
	return s.cacheable
}

func (s *Raster) SetCacheable(c bool) {
	s.cacheable = c
}

func (s *Raster) GetFileName() string {
	return s.fname
}

func (s *Raster) GetSize() [2]uint32 {
	if s.size == nil {
		s.size = make([]uint32, 2)
	}
	return [2]uint32{s.size[0], s.size[1]}
}

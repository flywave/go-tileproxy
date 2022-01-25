package vector

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geom"
	"github.com/flywave/go-mapbox/mvt"
	"github.com/flywave/go-tileproxy/tile"
)

type Vector map[string][]*geom.Feature

type VectorOptions struct {
	tile.TileOptions
	Format      tile.TileFormat
	Tolerance   float64
	Extent      uint16
	Buffer      uint16
	LineMetrics bool
	MaxZoom     uint8
	Proto       int
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

func (s *VectorSource) GetGeoReference() *geo.GeoReference {
	return s.georef
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
	var vec_opts *VectorOptions
	if in_tile_opts != nil {
		vec_opts = in_tile_opts.(*VectorOptions)
	} else {
		vec_opts = s.Options.(*VectorOptions)
	}
	if format != nil {
		vec_opts = s.Options.(*VectorOptions)
		vec_opts.Format = *format
	}
	if s.buf == nil {
		var err error
		s.buf, err = EncodeVector(vec_opts, s.tile, s.GetTile().(Vector))
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

func NewBlankVectorSource(size [2]uint32, opts tile.TileOptions, cacheable *tile.CacheInfo) tile.Source {
	format := opts.GetFormat()
	if format.Extension() == "mvt" || format.Extension() == "pbf" {
		_opts := opts.(*VectorOptions)
		return NewEmptyMVTSource(mvt.ProtoType(_opts.Proto), opts)
	} else if format.Extension() == "json" || format.Extension() == "geojson" {
		return NewEmptyGeoJSONVTSource(opts)
	}
	return nil
}

func CreateVectorSourceFromBufer(buf []byte, tile [3]int, opts *VectorOptions) tile.Source {
	format := opts.GetFormat()
	if format.Extension() == "mvt" || format.Extension() == "pbf" {
		mvt := NewMVTSource(tile, mvt.ProtoType(opts.Proto), opts)
		reader := bytes.NewBuffer(buf)
		mvt.SetSource(reader)
		return mvt
	} else if format.Extension() == "json" || format.Extension() == "geojson" {
		geojson := NewGeoJSONVTSource(tile, opts)
		reader := bytes.NewBuffer(buf)
		geojson.SetSource(reader)
		return geojson
	}
	return nil
}

func CreateVectorSourceFromVector(vt Vector, tile [3]int, opts *VectorOptions, cacheable *tile.CacheInfo) tile.Source {
	format := opts.GetFormat()
	if format.Extension() == "mvt" || format.Extension() == "pbf" {
		mvt := NewMVTSource(tile, mvt.ProtoType(opts.Proto), opts)
		mvt.SetSource(vt)
		mvt.SetCacheable(cacheable)
		return mvt
	} else if format.Extension() == "json" || format.Extension() == "geojson" {
		geojson := NewGeoJSONVTSource(tile, opts)
		geojson.SetSource(vt)
		geojson.SetCacheable(cacheable)
		return geojson
	}
	return nil
}

type VectorSourceCreater struct {
	Opt *VectorOptions
}

func (c *VectorSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return NewBlankVectorSource(size, opts, nil)
}

func (c *VectorSourceCreater) Create(data []byte, tile [3]int) tile.Source {
	return CreateVectorSourceFromBufer(data, tile, c.Opt)
}

func (c *VectorSourceCreater) GetExtension() string {
	return c.Opt.Format.Extension()
}

func EncodeVector(opts *VectorOptions, tile [3]int, data Vector) ([]byte, error) {
	if opts.Format.Extension() == "mvt" || opts.Format.Extension() == "pbf" {
		io := &PBFIO{tile: tile, proto: mvt.ProtoType(opts.Proto)}
		return io.Encode(data)
	} else if opts.Format.Extension() == "json" || opts.Format.Extension() == "geojson" {
		io := &GeoJSONVTIO{tile: tile, options: opts}
		return io.Encode(data)
	}
	return nil, errors.New("the format not support")
}

func DecodeVector(opts *VectorOptions, tile [3]int, reader io.Reader) (Vector, error) {
	if opts.Format.Extension() == "mvt" || opts.Format.Extension() == "pbf" {
		io := &PBFIO{tile: tile, proto: mvt.ProtoType(opts.Proto)}
		return io.Decode(reader)
	} else if opts.Format.Extension() == "json" || opts.Format.Extension() == "geojson" {
		io := &GeoJSONVTIO{tile: tile, options: opts}
		return io.Decode(reader)
	}
	return nil, errors.New("the format not support")
}

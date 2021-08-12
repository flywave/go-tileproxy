package raster

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"os"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

const (
	BILINEAR   = "bilinear"
	HYPERBOLIC = "hyperbolic"
)

type RasterOptions struct {
	tile.TileOptions
	Format       tile.TileFormat
	Mode         BorderMode
	DataType     RasterType
	Nodata       float64
	Interpolator string
}

func (s *RasterOptions) GetFormat() tile.TileFormat {
	return s.Format
}

type RasterSource struct {
	tile.Source
	data      *TileData
	buf       []byte
	fname     string
	size      []uint32
	pixelSize []float64
	cacheable *tile.CacheInfo
	georef    *geo.GeoReference
	Options   tile.TileOptions
	io        RasterIO
}

func (s *RasterSource) GetType() tile.TileType {
	return tile.TILE_DEM
}

func (s *RasterSource) GetCacheable() *tile.CacheInfo {
	if s.cacheable == nil {
		s.cacheable = &tile.CacheInfo{Cacheable: false}
	}
	return s.cacheable
}

func (s *RasterSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *RasterSource) GetFileName() string {
	return s.fname
}

func (s *RasterSource) GetTile() interface{} {
	return s.GetTileData()
}

func (s *RasterSource) GetTileData() *TileData {
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
		s.size = s.data.Size[:]
		if s.data.Boxsrs != nil {
			s.georef = geo.NewGeoReference(s.data.Box, s.data.Boxsrs)
		}
	}
	return s.data
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
		s.data, _ = s.io.Decode(ss)
	case string:
		s.fname = ss
	default:
		s.data = ss.(*TileData)
	}
}

func (s *RasterSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	if s.buf == nil {
		var err error
		s.buf, err = s.io.Encode(s.data)
		if err != nil {
			return nil
		}
	}
	return s.buf
}

func (s *RasterSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *RasterSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func (s *RasterSource) caclulatePixelSize(georef *geo.GeoReference) {
	if s.pixelSize == nil {
		s.pixelSize = []float64{0, 0}
		s.pixelSize[0] = (georef.GetBBox().Max[0] - georef.GetBBox().Min[0]) / float64(s.size[0])
		s.pixelSize[1] = (georef.GetBBox().Max[1] - georef.GetBBox().Min[1]) / float64(s.size[1])
	}
}

func getAverageExceptForNoDataValue(noData, valueIfAllBad float64, values ...float64) float64 {
	withValues := []float64{}

	epsilon := math.Nextafter(1, 2) - 1

	for _, v := range values {
		if math.Abs(v-noData) > epsilon {
			withValues = append(withValues, v)
		}
	}
	if len(withValues) > 0 {
		sum := 0.0
		for _, v := range withValues {
			sum += v
		}

		return sum / float64(len(withValues))
	} else {
		return valueIfAllBad
	}
}

const (
	NO_DATA_OUT = 0
)

func (s *RasterSource) GetElevation(lat, lon float64, georef *geo.GeoReference, interpolator Interpolator) float64 {
	heightValue := 0.0
	if georef == nil && s.georef != nil {
		georef = s.georef
	}

	if s.pixelSize == nil {
		s.caclulatePixelSize(georef)
	}

	opt := s.Options.(*RasterOptions)

	epsilon := math.Nextafter(1, 2) - 1

	noData := opt.Nodata

	var yPixel, xPixel, xInterpolationAmount, yInterpolationAmount float64

	dataEndLat := georef.GetOrigin()[1] + float64(s.pixelSize[1])*float64(s.size[1])

	if float64(s.pixelSize[1]) > 0 {
		yPixel = (dataEndLat - lat) / float64(s.pixelSize[1])
	} else {
		yPixel = (lat - dataEndLat) / float64(s.pixelSize[1])
	}
	xPixel = (lon - georef.GetOrigin()[0]) / float64(s.pixelSize[0])

	_, xInterpolationAmount = math.Modf(float64(xPixel))
	_, yInterpolationAmount = math.Modf(float64(yPixel))

	xOnDataPoint := math.Abs(xInterpolationAmount) < epsilon
	yOnDataPoint := math.Abs(yInterpolationAmount) < epsilon

	if xOnDataPoint && yOnDataPoint {
		x := int(math.Round(xPixel))
		y := int(math.Round(yPixel))
		heightValue = s.getElevation(x, y)
	} else {
		xCeiling := int(math.Ceil(xPixel))
		xFloor := int(math.Floor(xPixel))
		yCeiling := int(math.Ceil(yPixel))
		yFloor := int(math.Floor(yPixel))

		northWest := s.getElevation(xFloor, yFloor)
		northEast := s.getElevation(xCeiling, yFloor)
		southWest := s.getElevation(xFloor, yCeiling)
		southEast := s.getElevation(xCeiling, yCeiling)

		avgHeight := getAverageExceptForNoDataValue(noData, NO_DATA_OUT, southWest, southEast, northWest, northEast)

		if northWest == noData {
			northWest = avgHeight
		}
		if northEast == noData {
			northEast = avgHeight
		}
		if southWest == noData {
			southWest = avgHeight
		}
		if southEast == noData {
			southEast = avgHeight
		}

		heightValue = interpolator.Interpolate(southWest, southEast, northWest, northEast, xInterpolationAmount, yInterpolationAmount)
	}

	return heightValue
}

func (s *RasterSource) getElevation(x, y int) float64 {
	data := s.GetTileData()
	if data != nil {
		return data.Get(x, y)
	}
	return 0
}

func (s *RasterSource) Resample(georef *geo.GeoReference, grid *Grid) error {
	if georef == nil && s.georef != nil {
		georef = s.georef
	}
	if georef == nil {
		return errors.New("source georef is nil")
	}
	bbox := grid.GetRect()
	if !grid.srs.Eq(georef.GetSrs()) {
		bbox = grid.srs.TransformRectTo(georef.GetSrs(), bbox, 16)
	}
	if !geo.BBoxContains(georef.GetBBox(), bbox) {
		return errors.New("not Contains target grid")
	}
	opt := s.Options.(*RasterOptions)

	var interpolator Interpolator

	if opt.Interpolator == HYPERBOLIC {
		interpolator = &HyperbolicInterpolator{}
	} else {
		interpolator = &BilinearInterpolator{}
	}

	for i := range grid.Coordinates {
		coord := grid.Coordinates[i]
		lat, lon := coord[0], coord[1]
		if !grid.srs.Eq(georef.GetSrs()) {
			d := grid.srs.TransformTo(georef.GetSrs(), []vec2d.T{{lat, lon}})
			lat, lon = d[0][0], d[0][1]
		}
		grid.Coordinates[i][2] = s.GetElevation(lat, lon, georef, interpolator)
	}
	return nil
}

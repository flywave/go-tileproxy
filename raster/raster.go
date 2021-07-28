package raster

import (
	"io"
	"math"

	vec3d "github.com/flywave/go3d/float64/vec3"

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
	data             interface{}
	tp               RasterType
	buf              []byte
	fname            string
	size             []uint32
	pixelSize        []float64
	cacheable        bool
	georef           *geo.GeoReference
	nodata           float64
	Options          tile.TileOptions
	decodeFunc       func(r io.Reader) (interface{}, error)
	getElevationFunc func(d interface{}, x, y int) float64
	interpolator     Interpolator
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

func (s *RasterSource) caclulatePixelSize() {
	if s.pixelSize == nil {
		s.pixelSize = []float64{0, 0}
		s.pixelSize[0] = (s.georef.GetBBox().Max[0] - s.georef.GetBBox().Min[0]) / float64(s.size[0])
		s.pixelSize[1] = (s.georef.GetBBox().Max[1] - s.georef.GetBBox().Min[1]) / float64(s.size[1])
	}
}

func GetAverageExceptForNoDataValue(noData, valueIfAllBad float64, values ...float64) float64 {
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

func (s *RasterSource) GetElevation(lat, lon float64) float64 {
	heightValue := 0.0

	if s.pixelSize == nil {
		s.caclulatePixelSize()
	}

	epsilon := math.Nextafter(1, 2) - 1

	noData := s.nodata

	var yPixel, xPixel, xInterpolationAmount, yInterpolationAmount float64

	dataEndLat := s.georef.GetOrigin()[1] + float64(s.pixelSize[1])*float64(s.size[1])

	if float64(s.pixelSize[1]) > 0 {
		yPixel = (dataEndLat - lat) / float64(s.pixelSize[1])
	} else {
		yPixel = (lat - dataEndLat) / float64(s.pixelSize[1])
	}
	xPixel = (lon - s.georef.GetOrigin()[0]) / float64(s.pixelSize[0])

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

		avgHeight := GetAverageExceptForNoDataValue(noData, NO_DATA_OUT, southWest, southEast, northWest, northEast)

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

		heightValue = s.interpolator.Interpolate(southWest, southEast, northWest, northEast, xInterpolationAmount, yInterpolationAmount)
	}

	return heightValue
}

func (s *RasterSource) getElevation(x, y int) float64 {
	return s.getElevationFunc(s.data, x, y)
}

func (s *RasterSource) GetHeightMap() *HeightMap {
	opts := s.Options.(*RasterOptions)
	heightMap := NewHeightMap(int(s.size[0]), int(s.size[1]))
	heightMap.Count = heightMap.Width * heightMap.Height
	heightMap.srs = s.georef.GetSrs()
	coords := make(Coordinates, heightMap.Count)

	if s.pixelSize == nil {
		s.caclulatePixelSize()
	}

	for y := 0; y < int(s.size[1]); y++ {
		latitude := s.georef.GetOrigin()[1] + (float64(s.pixelSize[1]) * float64(y))
		for x := 0; x < int(s.size[0]); x++ {
			longitude := s.georef.GetOrigin()[0] + (float64(s.pixelSize[0]) * float64(x))

			heightValue := s.getElevation(x, y)

			if heightValue < float64(32768) {
				heightMap.Minimum = math.Min(opts.MinimumAltitude, heightValue)
				heightMap.Maximum = math.Max(opts.MaximumAltitude, heightValue)
			} else {
				heightValue = 0
			}
			coords = append(coords, vec3d.T{latitude, longitude, heightValue})
		}
	}

	heightMap.Coordinates = coords
	return heightMap
}

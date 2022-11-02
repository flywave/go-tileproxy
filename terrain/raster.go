package terrain

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"os"

	"github.com/flywave/go-geoid"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
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
	MaxError     float64
	Nodata       float64
	Interpolator string
	HeightModel  geoid.VerticalDatum
	HeightOffset float64
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

func (s *RasterSource) SetGeoReference(georef *geo.GeoReference) {
	s.georef = georef
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

		opt := s.Options.(*RasterOptions)
		noData := opt.Nodata

		r := bytes.NewBuffer(s.buf)
		var err error
		s.data, err = s.io.Decode(r)
		if err != nil {
			return nil
		}
		s.data.NoData = noData
		s.size = s.data.Size[:]
		if s.data.Boxsrs != nil {
			s.georef = geo.NewGeoReference(s.data.Box, s.data.Boxsrs)
		}
	}
	return s.data
}

func (s *RasterSource) GetGeoReference() *geo.GeoReference {
	return s.georef
}

func (s *RasterSource) SetGeoref(ref *geo.GeoReference) {
	s.georef = ref
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
		opt := s.Options.(*RasterOptions)
		noData := opt.Nodata
		s.data, _ = s.io.Decode(ss)
		s.size = s.data.Size[:]
		s.data.NoData = noData
		if s.data.Boxsrs != nil {
			s.georef = geo.NewGeoReference(s.data.Box, s.data.Boxsrs)
		}
	case string:
		s.fname = ss
	default:
		s.data = ss.(*TileData)
		s.size = s.data.Size[:]
		if s.data.Boxsrs != nil {
			s.georef = geo.NewGeoReference(s.data.Box, s.data.Boxsrs)
		}
	}
}

func (s *RasterSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	var dem_opts *RasterOptions
	if in_tile_opts != nil {
		dem_opts = in_tile_opts.(*RasterOptions)
	} else {
		dem_opts = s.Options.(*RasterOptions)
	}
	if format != nil {
		dem_opts = s.Options.(*RasterOptions)
		dem_opts.Format = *format
	}
	if s.buf == nil {
		var err error
		s.buf, err = EncodeRaster(dem_opts, s.data)
		if err != nil {
			return nil
		}
	}
	return s.buf
}

func (s *RasterSource) GetRasterOptions() *RasterOptions {
	return s.Options.(*RasterOptions)
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

func (s *RasterSource) GetElevation(lon, lat float64, georef *geo.GeoReference, interpolator Interpolator) float64 {
	heightValue := 0.0
	if georef == nil && s.georef != nil {
		georef = s.georef
	}

	if s.pixelSize == nil {
		s.caclulatePixelSize(georef)
	}

	opt := s.Options.(*RasterOptions)

	noData := opt.Nodata

	var yPixel, xPixel, xInterpolationAmount, yInterpolationAmount float64

	dataEndLat := georef.GetOrigin()[1] + float64(s.pixelSize[1])*float64(s.size[1])

	if float64(s.pixelSize[1]) > 0 {
		yPixel = ((dataEndLat - lat) / float64(s.pixelSize[1]))
	} else {
		yPixel = (lat - dataEndLat) / float64(s.pixelSize[1])
	}
	xPixel = (lon - georef.GetOrigin()[0]) / float64(s.pixelSize[0])

	_, xInterpolationAmount = math.Modf(float64(xPixel))
	_, yInterpolationAmount = math.Modf(float64(yPixel))

	xOnDataPoint := math.Abs(xInterpolationAmount) < 0.1 || math.Abs(xInterpolationAmount) > 0.9
	yOnDataPoint := math.Abs(yInterpolationAmount) < 0.1 || math.Abs(yInterpolationAmount) > 0.9

	if xOnDataPoint && yOnDataPoint {
		x := int(math.Floor(xPixel))
		y := int(math.Floor(yPixel))
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
	if x >= int(data.Size[0]) {
		x = int(data.Size[0] - 1)
	}
	if x < 0 {
		x = 0
	}

	if y >= int(data.Size[1]) {
		y = int(data.Size[1] - 1)
	}
	if y < 0 {
		y = 0
	}
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
		lon, lat := coord[0], coord[1]
		if !grid.srs.Eq(georef.GetSrs()) {
			d := grid.srs.TransformTo(georef.GetSrs(), []vec2d.T{{lon, lat}})
			lon, lat = d[0][0], d[0][1]
		}
		grid.Coordinates[i][2] = s.GetElevation(lon, lat, georef, interpolator)
	}
	return nil
}

func NewBlankRasterSource(size [2]uint32, opts tile.TileOptions, cacheable *tile.CacheInfo) tile.Source {
	switch opt := opts.(type) {
	case *RasterOptions:
		td := NewTileData(size, opt.Mode)
		if opt.Format.Extension() == "webp" || opt.Format.Extension() == "png" {
			src := NewDemRasterSource(ModeMapbox, opt)
			src.SetSource(td)
			return src
		} else if opt.Format.Extension() == "atm" {
			src := NewLercRasterSource(opt.Mode, opt.MaxError, opt)
			src.SetSource(td)
			return src
		} else if opt.Format.Extension() == "tif" || opt.Format.Extension() == "tiff" {
			src := NewGeoTIFFRasterSource(opt.Mode, opt)
			src.SetSource(td)
			return src
		} else if opt.Format.Extension() == "terrain" {
			tiledata := NewTileData(size, opt.Mode)
			tiledata.NoData = -9999
			source, err := GenTerrainSource(tiledata, opt)

			if err != nil {
				return nil
			}
			return source
		}
	}
	return nil
}

func CreateRasterSourceFromBufer(buf []byte, opts *RasterOptions) tile.Source {
	if opts.Format.Extension() == "webp" || opts.Format.Extension() == "png" {
		src := NewDemRasterSource(ModeMapbox, opts)
		reader := bytes.NewBuffer(buf)
		src.SetSource(reader)
		return src
	} else if opts.Format.Extension() == "atm" {
		src := NewLercRasterSource(opts.Mode, opts.MaxError, opts)
		reader := bytes.NewBuffer(buf)
		src.SetSource(reader)
		return src
	} else if opts.Format.Extension() == "tif" || opts.Format.Extension() == "tiff" {
		src := NewGeoTIFFRasterSource(opts.Mode, opts)
		reader := bytes.NewBuffer(buf)
		src.SetSource(reader)
		return src
	} else if opts.Format.Extension() == "terrain" {
		src := NewTerrainSource(opts)
		reader := utils.NewMemFile(buf)
		src.SetSource(reader)
		return src
	}
	return nil
}

func CreateRasterSourceFromTileData(data *TileData, opts *RasterOptions, cacheable *tile.CacheInfo) tile.Source {
	if opts.Format.Extension() == "webp" || opts.Format.Extension() == "png" {
		src := NewDemRasterSource(ModeMapbox, opts)
		src.SetSource(data)
		src.SetCacheable(cacheable)
		return src
	} else if opts.Format.Extension() == "atm" {
		src := NewLercRasterSource(opts.Mode, opts.MaxError, opts)
		src.SetSource(data)
		src.SetCacheable(cacheable)
		return src
	} else if opts.Format.Extension() == "tif" || opts.Format.Extension() == "tiff" {
		src := NewGeoTIFFRasterSource(opts.Mode, opts)
		src.SetSource(data)
		src.SetCacheable(cacheable)
		return src
	} else if opts.Format.Extension() == "terrain" {
		src, _ := GenTerrainSource(data, opts)
		return src
	}
	return nil
}

type RasterSourceCreater struct {
	Opt *RasterOptions
}

func (c *RasterSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return NewBlankRasterSource(size, opts, nil)
}

func (c *RasterSourceCreater) Create(data []byte, tile [3]int) tile.Source {
	return CreateRasterSourceFromBufer(data, c.Opt)
}

func (c *RasterSourceCreater) GetExtension() string {
	return c.Opt.Format.Extension()
}

func EncodeRaster(opts *RasterOptions, data *TileData) ([]byte, error) {
	if opts.Format.Extension() == "webp" || opts.Format.Extension() == "png" {
		io := &DemIO{Mode: ModeMapbox, Format: opts.Format}
		return io.Encode(data)
	} else if opts.Format.Extension() == "atm" {
		io := &LercIO{Mode: opts.Mode, MaxZError: opts.MaxError}
		return io.Encode(data)
	} else if opts.Format.Extension() == "tif" || opts.Format.Extension() == "tiff" {
		io := &GeoTIFFIO{Mode: opts.Mode}
		return io.Encode(data)
	} else if opts.Format.Extension() == "terrain" {
		mesh, err := EncodeQuatMesh(data, opts)
		if err != nil {
			return nil, err
		}
		buf, err := EncodeMesh(mesh)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	return nil, errors.New("the format not support")
}

func DecodeRaster(opts *RasterOptions, reader io.Reader) (*TileData, error) {
	if opts.Format.Extension() == "webp" || opts.Format.Extension() == "png" {
		io := &DemIO{Mode: ModeMapbox, Format: opts.Format}
		return io.Decode(reader)
	} else if opts.Format.Extension() == "atm" {
		io := &LercIO{Mode: opts.Mode, MaxZError: opts.MaxError}
		return io.Decode(reader)
	} else if opts.Format.Extension() == "tif" || opts.Format.Extension() == "tiff" {
		io := &GeoTIFFIO{Mode: opts.Mode}
		return io.Decode(reader)
	}
	return nil, errors.New("the format not support")
}

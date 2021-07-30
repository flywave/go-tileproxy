package raster

import (
	"io"

	geotiff "github.com/flywave/go-geotiff"
)

var (
	GEOTIFF_DT_MAP = map[int]RasterType{
		geotiff.DT_INT8:    RT_CHAR,
		geotiff.DT_UINT8:   RT_UCHAR,
		geotiff.DT_INT16:   RT_SHORT,
		geotiff.DT_UINT16:  RT_USHORT,
		geotiff.DT_INT32:   RT_INT,
		geotiff.DT_UINT32:  RT_UINT,
		geotiff.DT_FLOAT32: RT_FLOAT,
		geotiff.DT_FLOAT64: RT_DOUBLE,
	}
)

type GeoTIFFRasterSource struct {
	RasterSource
}

func LoadTiff(r io.Reader) (error, *geotiff.Raster) {
	rat := r.(io.ReaderAt)
	raster, err := geotiff.CreateRasterFromStream(rat)
	if err != nil {
		return err, nil
	}
	return nil, raster
}

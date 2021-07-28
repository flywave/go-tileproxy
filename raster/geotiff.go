package raster

import (
	"io"

	geotiff "github.com/flywave/go-geotiff"
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

package raster

import "io"

type RasterIO interface {
	Decode(r io.Reader) (interface{}, error)
	Encode(tile interface{}) ([]byte, error)
	GetElevation(tile interface{}, x, y int) float64
}

package raster

import (
	"io"

	mraster "github.com/flywave/go-mapbox/raster"
)

type RasterDemMode uint32

const (
	ModeMapbox    RasterDemMode = mraster.DEM_ENCODING_MAPBOX
	ModeTerrarium RasterDemMode = mraster.DEM_ENCODING_TERRARIUM
)

type DemRasterSource struct {
	RasterSource
	Mode RasterDemMode
}

func LoadDEM(r io.Reader, mode RasterDemMode) (*mraster.DEMData, error) {
	return mraster.LoadDEMDataWithStream(r, int(mode))
}

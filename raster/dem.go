package raster

import (
	"bytes"
	"image"
	"image/color"
	"io"

	mraster "github.com/flywave/go-mapbox/raster"

	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/tile"
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

func EncodeDEM(d *mraster.DEMData, format tile.TileFormat) ([]byte, error) {
	idx := func(x int, y int) int {
		return (y+1)*d.Stride + (x + 1)
	}

	img := image.NewNRGBA(image.Rect(0, 0, d.Dim, d.Dim))

	for y := 0; y < d.Dim; y++ {
		for x := 0; x < d.Dim; x++ {
			value := d.Data[idx(x, y)]
			img.SetNRGBA(x, y, color.NRGBA{
				R: value[0],
				G: value[1],
				B: value[2],
				A: value[3],
			})
		}
	}
	rets := &bytes.Buffer{}
	images.EncodeImage(format.Extension(), rets, img)

	return rets.Bytes(), nil
}

type DemIO struct {
	RasterIO
}

func (d *DemIO) Decode(r io.Reader) (interface{}, error) {
	return nil, nil
}

func (d *DemIO) Encode(tile interface{}) ([]byte, error) {
	return nil, nil
}

func (d *DemIO) GetElevation(tile interface{}, x, y int) float64 {
	return 0
}

package terrain

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"io"

	mraster "github.com/flywave/go-mapbox/raster"

	"github.com/flywave/go-tileproxy/imagery"
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

func NewDemRasterSource(mode RasterDemMode, options tile.TileOptions) *DemRasterSource {
	src := &DemRasterSource{RasterSource: RasterSource{Options: options}}
	src.io = &DemIO{Mode: mode, Format: options.GetFormat()}
	return src
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
	imagery.EncodeImage(format.Extension(), rets, img)

	return rets.Bytes(), nil
}

type DemIO struct {
	RasterIO
	Mode   RasterDemMode
	Format tile.TileFormat
}

func (d *DemIO) Decode(r io.Reader) (*TileData, error) {
	data, err := mraster.LoadDEMDataWithStream(r, int(d.Mode))
	if err != nil {
		return nil, err
	}
	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_BILATERAL)
	for x := 0; x < data.Dim; x++ {
		for y := 0; y < data.Dim; y++ {
			if x > 0 && y > 0 && x < data.Dim-1 && y < data.Dim-1 {
				tiledata.Set(x-1, y-1, data.Get(x, y))
			}

			if x == 0 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_LEFT, y-1, data.Get(x, y))
			}

			if x == data.Dim-1 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_RIGHT, y-1, data.Get(x, y))
			}

			if y == 0 {
				tiledata.FillBorder(BORDER_TOP, x, data.Get(x, y))
			}

			if y == data.Dim-1 {
				tiledata.FillBorder(BORDER_BOTTOM, x, data.Get(x, y))
			}
		}
	}
	return tiledata, nil
}

var (
	_terrariumPacker mraster.DemPacker
	_mapboxPacker    mraster.DemPacker
)

func init() {
	_terrariumPacker = &mraster.TerrariumPacker{}
	_mapboxPacker = &mraster.MapboxPacker{}
}

func (d *DemIO) Encode(tile *TileData) ([]byte, error) {
	if !tile.IsBilateral() {
		return nil, errors.New("dem must sample bilateral border!")
	}
	data, si := tile.GetExtend(nil)
	if si[0] != si[1] {
		return nil, errors.New("row === col!")
	}

	var pack func(h float64) [4]byte

	if d.Mode == ModeMapbox {
		pack = func(h float64) [4]byte {
			return _mapboxPacker.Pack(h)
		}
	} else {
		pack = func(h float64) [4]byte {
			return _terrariumPacker.Pack(h)
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, int(si[0]), int(si[1])))

	for y := 0; y < int(si[1]); y++ {
		for x := 0; x < int(si[0]); x++ {
			rgba := pack(data[y*int(si[0])+x])
			img.SetNRGBA(x, y, color.NRGBA{
				R: rgba[0],
				G: rgba[1],
				B: rgba[2],
				A: rgba[3],
			})
		}
	}
	writer := &bytes.Buffer{}
	imagery.EncodeImage(d.Format.Extension(), writer, img)

	return writer.Bytes(), nil
}

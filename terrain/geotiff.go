package terrain

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	geotiff "github.com/flywave/go-geotiff"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
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

func NewGeoTIFFRasterSource(mode BorderMode, options tile.TileOptions) *GeoTIFFRasterSource {
	src := &GeoTIFFRasterSource{RasterSource: RasterSource{Options: options}}
	src.io = &GeoTIFFIO{Mode: mode}
	return src
}

func LoadTiff(r io.Reader) (*geotiff.Raster, error) {
	rat := r.(io.ReaderAt)
	raster, err := geotiff.CreateRasterFromStream(rat)
	if err != nil {
		return nil, err
	}
	return raster, nil
}

func EncodeTiff(r *geotiff.Raster) ([]byte, error) {
	wr := &bytes.Buffer{}
	r.SetWriter(wr)
	err := r.Save()
	if err != nil {
		return nil, err
	}
	return wr.Bytes(), nil
}

type GeoTIFFIO struct {
	RasterIO
	Mode BorderMode
}

func (d *GeoTIFFIO) Decode(r io.Reader) (*TileData, error) {
	raster, err := LoadTiff(r)
	if err != nil {
		return nil, err
	}
	off := 0
	if d.Mode == BORDER_UNILATERAL {
		off = 1
	} else if d.Mode == BORDER_BILATERAL {
		off = 2
	}

	row, col := raster.Rows(), raster.Columns()
	tiledata := NewTileData([2]uint32{uint32(col - off), uint32(row - off)}, d.Mode)

	tiledata.Box.Min[0], tiledata.Box.Min[1], tiledata.Box.Max[0], tiledata.Box.Max[1] = raster.West(), raster.South(), raster.East(), raster.North()
	tiledata.Boxsrs = geo.NewSRSProj4(fmt.Sprintf("EPSG:%d", raster.GetRasterConfig().EPSGCode))

	if d.Mode == BORDER_UNILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 {
					tiledata.Set(x-1, y-1, raster.Value(y, x))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, raster.Value(y, x))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, raster.Value(y, x))
				}
			}
		}
	} else if d.Mode == BORDER_BILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 && x < col-1 && y < row-1 {
					tiledata.Set(x-1, y-1, raster.Value(y, x))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, raster.Value(y, x))
				}

				if x == col-1 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_RIGHT, y-1, raster.Value(y, x))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, raster.Value(y, x))
				}

				if y == row-1 {
					tiledata.FillBorder(BORDER_BOTTOM, x, raster.Value(y, x))
				}
			}
		}
	} else {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				tiledata.Set(x, y, raster.Value(y, x))
			}
		}
	}
	return tiledata, nil
}

func (d *GeoTIFFIO) Encode(tile *TileData) ([]byte, error) {
	if d.Mode != tile.Border {
		return nil, errors.New("Border mode error")
	}
	data, si, tran := tile.GetExtend()

	var north, south, east, west float64
	if tran[5] < 0 {
		north = tran[3]
		south = tran[3] + tran[5]*float64(si[1])
	} else {
		south = tran[3]
		north = tran[3] + tran[5]*float64(si[1])
	}
	if tran[1] < 0 {
		east = tran[0]
		west = tran[0] + tran[1]*float64(si[0])
	} else {
		west = tran[0]
		east = tran[0] + tran[1]*float64(si[0])
	}

	conf := geotiff.NewDefaultRasterConfig()
	conf.EPSGCode = geo.GetEpsgNum(tile.Boxsrs.GetSrsCode())

	raster, err := geotiff.CreateNewRaster("", int(si[1]), int(si[0]), north, south, east, west, conf)
	if err != nil {
		return nil, err
	}

	raster.SetData(data)
	writer := &bytes.Buffer{}
	raster.SetWriter(writer)
	raster.Save()

	return writer.Bytes(), nil
}

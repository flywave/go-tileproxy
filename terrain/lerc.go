package terrain

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/flywave/go-lerc"
	"github.com/flywave/go-tileproxy/tile"
)

var (
	LERC_DT_MAP = map[int]RasterType{
		lerc.DT_CHAR:   RT_CHAR,
		lerc.DT_UCHAR:  RT_UCHAR,
		lerc.DT_SHORT:  RT_SHORT,
		lerc.DT_USHORT: RT_USHORT,
		lerc.DT_INT:    RT_INT,
		lerc.DT_UINT:   RT_UINT,
		lerc.DT_FLOAT:  RT_FLOAT,
		lerc.DT_DOUBLE: RT_DOUBLE,
	}
)

type LercRasterSource struct {
	RasterSource
}

func NewLercRasterSource(mode BorderMode, maxZError float64, options tile.TileOptions) *LercRasterSource {
	src := &LercRasterSource{RasterSource: RasterSource{Options: options}}
	src.io = &LercIO{Mode: mode, MaxZError: maxZError}
	return src
}

func LoadLerc(r io.Reader) (interface{}, lerc.BlobInfo, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	binfo, err := lerc.GetBlobInfo(src)
	if err != nil {
		return nil, nil, err
	}

	newImg, _, err := lerc.Decode(src)
	if err != nil {
		return nil, nil, err
	}

	switch v := newImg.(type) {
	case []int8:
		return v, binfo, nil
	case []uint8:
		return v, binfo, nil
	case []int16:
		return v, binfo, nil
	case []uint16:
		return v, binfo, nil
	case []int32:
		return v, binfo, nil
	case []uint32:
		return v, binfo, nil
	case []float32:
		return v, binfo, nil
	case []float64:
		return v, binfo, nil
	}
	return nil, binfo, errors.New("format error")
}

func EncodeLerc(data interface{}, dim int, cols int, rows int, bands int, maxZErr float64) ([]byte, error) {
	mask := make([]byte, cols*rows)
	return lerc.Encode(data, dim, cols, rows, bands, mask, maxZErr)
}

type LercIO struct {
	RasterIO
	Mode      BorderMode
	MaxZError float64
	LercDT    int
}

func getValue(data interface{}, row, col int, cols int) float64 {
	switch v := data.(type) {
	case []int8:
		return float64(v[row*cols+col])
	case []uint8:
		return float64(v[row*cols+col])
	case []int16:
		return float64(v[row*cols+col])
	case []uint16:
		return float64(v[row*cols+col])
	case []int32:
		return float64(v[row*cols+col])
	case []uint32:
		return float64(v[row*cols+col])
	case []float32:
		return float64(v[row*cols+col])
	case []float64:
		return v[row*cols+col]
	}
	return 99999
}

func (d *LercIO) Decode(r io.Reader) (*TileData, error) {
	vec, blobInfo, err := LoadLerc(r)
	if err != nil {
		return nil, err
	}
	off := 0
	if d.Mode == BORDER_UNILATERAL {
		off = 1
	} else if d.Mode == BORDER_BILATERAL {
		off = 2
	}

	row, col := int(blobInfo.Rows()), int(blobInfo.Cols())

	tiledata := NewTileData([2]uint32{uint32(col - off), uint32(row - off)}, d.Mode)
	if d.Mode == BORDER_UNILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 {
					tiledata.Set(x-1, y-1, getValue(vec, y, x, col))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, getValue(vec, y, x, col))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, getValue(vec, y, x, col))
				}
			}
		}
	} else if d.Mode == BORDER_BILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 && x < col-1 && y < row-1 {
					tiledata.Set(x-1, y-1, getValue(vec, y, x, col))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, getValue(vec, y, x, col))
				}

				if x == col-1 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_RIGHT, y-1, getValue(vec, y, x, col))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, getValue(vec, y, x, col))
				}

				if y == row-1 {
					tiledata.FillBorder(BORDER_BOTTOM, x, getValue(vec, y, x, col))
				}
			}
		}
	} else {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				tiledata.Set(x, y, getValue(vec, y, x, col))
			}
		}
	}
	return tiledata, nil
}

func (d *LercIO) Encode(tile *TileData) ([]byte, error) {
	if d.Mode != tile.Border {
		return nil, errors.New("border mode error")
	}
	data, si, _ := tile.GetExtend32()

	maxZError := d.MaxZError
	if maxZError == 0 {
		maxZErrorWanted := 0.1
		eps := 0.0001
		maxZError = maxZErrorWanted - eps
	}

	return EncodeLerc(data, 1, int(si[0]), int(si[1]), 1, maxZError)
}

package raster

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/flywave/go-lerc"
)

type LercDataType uint32

const (
	LCHAR   LercDataType = lerc.DT_CHAR
	LUCHAR  LercDataType = lerc.DT_UCHAR
	LSHORT  LercDataType = lerc.DT_SHORT
	LUSHORT LercDataType = lerc.DT_USHORT
	LINT    LercDataType = lerc.DT_INT
	LUINT   LercDataType = lerc.DT_UINT
	LFLOAT  LercDataType = lerc.DT_FLOAT
	LDOUBLE LercDataType = lerc.DT_DOUBLE
)

type LercRasterSource struct {
	RasterSource
	DataType LercDataType
}

func LoadLerc(r io.Reader) (error, interface{}, lerc.BlobInfo) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err, nil, nil
	}

	binfo, err := lerc.GetBlobInfo(src)
	if err != nil {
		return err, nil, nil
	}

	newImg, _, err := lerc.Decode(src)
	if err == nil {
		return err, nil, nil
	}

	switch v := newImg.(type) {
	case []int8:
		return nil, v, binfo
	case []uint8:
		return nil, v, binfo
	case []int16:
		return nil, v, binfo
	case []uint16:
		return nil, v, binfo
	case []int32:
		return nil, v, binfo
	case []uint32:
		return nil, v, binfo
	case []float32:
		return nil, v, binfo
	case []float64:
		return nil, v, binfo
	}
	return errors.New("format error"), nil, binfo
}

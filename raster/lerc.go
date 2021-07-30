package raster

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/flywave/go-lerc"
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

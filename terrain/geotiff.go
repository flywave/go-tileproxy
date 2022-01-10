package terrain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/google/tiff"
)

type GeoTIFFRasterSource struct {
	RasterSource
}

func NewGeoTIFFRasterSource(mode BorderMode, options tile.TileOptions) *GeoTIFFRasterSource {
	src := &GeoTIFFRasterSource{RasterSource: RasterSource{Options: options}}
	src.io = &GeoTIFFIO{Mode: mode}
	return src
}

func LoadTiff(r io.Reader) (*cog.Reader, error) {
	rat := r.(tiff.ReadAtReadSeeker)
	raster := cog.ReadFrom(rat)
	return raster, nil
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

	si := raster.GetSize(0)

	row, col := int(si[1]), int(si[0])
	tiledata := NewTileData([2]uint32{uint32(col - off), uint32(row - off)}, d.Mode)

	tiledata.Box = raster.GetBounds(0)

	epsg, err := raster.GetEPSGCode(0)

	if err != nil {
		return nil, err
	}

	tiledata.Boxsrs = geo.NewProj(fmt.Sprintf("EPSG:%d", epsg))

	imageData := raster.Data[0].([]float64)

	if d.Mode == BORDER_UNILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 {
					tiledata.Set(x-1, y-1, imageData[y*col+x])
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, imageData[y*col+x])
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, imageData[y*col+x])
				}
			}
		}
	} else if d.Mode == BORDER_BILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 && x < col-1 && y < row-1 {
					tiledata.Set(x-1, y-1, imageData[y*col+x])
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, imageData[y*col+x])
				}

				if x == col-1 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_RIGHT, y-1, imageData[y*col+x])
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, imageData[y*col+x])
				}

				if y == row-1 {
					tiledata.FillBorder(BORDER_BOTTOM, x, imageData[y*col+x])
				}
			}
		}
	} else {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				tiledata.Set(x, y, imageData[y*col+x])
			}
		}
	}
	return tiledata, nil
}

func (d *GeoTIFFIO) Encode(tile *TileData) ([]byte, error) {
	if d.Mode != tile.Border {
		return nil, errors.New("border mode error")
	}
	data, si, tran := tile.GetExtend()

	var miny, maxy, minx, maxx float64
	if tran[5] < 0 {
		maxy = tran[3]
		miny = tran[3] + tran[5]*float64(si[1])
	} else {
		miny = tran[3]
		maxy = tran[3] + tran[5]*float64(si[1])
	}
	if tran[1] < 0 {
		maxx = tran[0]
		minx = tran[0] + tran[1]*float64(si[0])
	} else {
		minx = tran[0]
		maxx = tran[0] + tran[1]*float64(si[0])
	}

	bbox := vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
	rect := image.Rect(0, 0, int(si[0]), int(si[1]))
	src := cog.NewSource(data, &rect, cog.CTLZW)

	w := cog.NewTileWriter(src, binary.LittleEndian, false, bbox, tile.Boxsrs, si, nil)
	writer := &bytes.Buffer{}

	err := w.WriteData(writer)

	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

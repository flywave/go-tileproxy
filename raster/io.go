package raster

import (
	"io"

	"github.com/flywave/go-tileproxy/geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type RasterType uint32

const (
	RT_CHAR   RasterType = 0
	RT_UCHAR  RasterType = 1
	RT_SHORT  RasterType = 2
	RT_USHORT RasterType = 3
	RT_INT    RasterType = 4
	RT_UINT   RasterType = 5
	RT_FLOAT  RasterType = 6
	RT_DOUBLE RasterType = 7
)

type BorderMode uint32

const (
	BORDER_NONE       BorderMode = 0
	BORDER_UNILATERAL BorderMode = 1
	BORDER_BILATERAL  BorderMode = 2
)

type BorderType uint32

const (
	BORDER_LEFT   BorderType = 0
	BORDER_TOP    BorderType = 1
	BORDER_RIGHT  BorderType = 2
	BORDER_BOTTOM BorderType = 3
)

type TileData struct {
	Size         [2]uint32
	Datas        []float64
	Border       BorderMode
	LeftBorder   []float64
	TopBorder    []float64
	RightBorder  []float64
	BottomBorder []float64
	Box          vec2d.Rect
	Boxsrs       geo.Proj
}

func NewTileData(size [2]uint32, border BorderMode) *TileData {
	d := &TileData{Size: size, Datas: make([]float64, size[0]*size[1]), Border: border}
	if border == BORDER_UNILATERAL {
		d.LeftBorder = make([]float64, size[1])
		d.TopBorder = make([]float64, size[0]+1)
	} else if border == BORDER_BILATERAL {
		d.LeftBorder = make([]float64, size[1])
		d.TopBorder = make([]float64, size[0]+2)
		d.RightBorder = make([]float64, size[1])
		d.BottomBorder = make([]float64, size[0]+2)
	}
	return d
}

func (d *TileData) HasBorder() bool {
	return d.Border != BORDER_NONE
}

func (d *TileData) IsBilateral() bool {
	return d.Border == BORDER_BILATERAL
}

func (d *TileData) IsUnilateral() bool {
	return d.Border == BORDER_UNILATERAL
}

func (d *TileData) FillBorder(tp BorderType, i int, h float64) {
	switch tp {
	case BORDER_LEFT:
		if i < len(d.LeftBorder) {
			d.LeftBorder[i] = h
		}
	case BORDER_RIGHT:
		if i < len(d.RightBorder) {
			d.RightBorder[i] = h
		}
	case BORDER_TOP:
		if i < len(d.TopBorder) {
			d.TopBorder[i] = h
		}
	case BORDER_BOTTOM:
		if i < len(d.BottomBorder) {
			d.BottomBorder[i] = h
		}
	}
}

func (d *TileData) Set(x, y int, h float64) {
	d.Datas[y*int(d.Size[0])+x] = h
}

func (d *TileData) Get(x, y int) float64 {
	return d.Datas[y*int(d.Size[0])+x]
}

func (d *TileData) GetExtend() ([]float64, [2]uint32) {
	if !d.HasBorder() {
		return d.Datas[:], d.Size
	}
	if d.IsBilateral() {
		w, h := (d.Size[0] + 2), (d.Size[1] + 2)
		ret := make([]float64, w*h)
		copy(ret[:w], d.TopBorder)

		for y := 0; y < int(d.Size[1]); y++ {
			off := (y+1)*int(w) + 1
			ret[off-1] = d.LeftBorder[y]
			ret[off+int(d.Size[0])] = d.RightBorder[y]

			copy(ret[off:off+int(d.Size[0])], d.Datas[y*int(d.Size[0]):(y+1)*int(d.Size[0])])
		}

		off := (d.Size[1] + 1) * w
		copy(ret[off:int(off+w)], d.BottomBorder)

		return ret, [2]uint32{(d.Size[0] + 2), (d.Size[1] + 2)}
	}
	if d.IsUnilateral() {
		w, h := (d.Size[0] + 1), (d.Size[1] + 1)
		ret := make([]float64, w*h)
		copy(ret[:w], d.TopBorder)

		for y := 0; y < int(d.Size[1]); y++ {
			off := (y+1)*int(w) + 1
			ret[off-1] = d.LeftBorder[y]
			copy(ret[off:off+int(d.Size[0])], d.Datas[y*int(d.Size[0]):(y+1)*int(d.Size[0])])
		}

		return ret, [2]uint32{(d.Size[0] + 1), (d.Size[1] + 1)}
	}
	return nil, [2]uint32{}
}

func (d *TileData) GetExtend32() ([]float32, [2]uint32) {
	rawdata, si := d.GetExtend()
	ret := make([]float32, len(rawdata))
	for i := range rawdata {
		ret[i] = float32(rawdata[i])
	}
	return ret, si
}

type RasterIO interface {
	Decode(r io.Reader) (*TileData, error)
	Encode(tile *TileData) ([]byte, error)
}

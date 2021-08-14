package terrain

import (
	"io"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
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
	NoData       float64
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

func (d *TileData) NoDataValue() float64 {
	return d.NoData
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

func (d *TileData) GetExtend() ([]float64, [2]uint32, [6]float64) {
	pixelsize := caclulatePixelSize(int(d.Size[0]), int(d.Size[1]), d.Box)

	if !d.HasBorder() {
		return d.Datas[:], d.Size, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Min[1], 0, -pixelsize[1]}
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

		return ret, [2]uint32{(d.Size[0] + 2), (d.Size[1] + 2)}, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Min[1], 0, -pixelsize[1]}
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

		return ret, [2]uint32{(d.Size[0] + 1), (d.Size[1] + 1)}, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Min[1], 0, -pixelsize[1]}
	}
	return nil, [2]uint32{}, [6]float64{}
}

func (d *TileData) GetExtend32() ([]float32, [2]uint32, [6]float64) {
	rawdata, si, tran := d.GetExtend()
	ret := make([]float32, len(rawdata))
	for i := range rawdata {
		ret[i] = float32(rawdata[i])
	}
	return ret, si, tran
}

func (d *TileData) CopyFrom(src *TileData, pos [2]int) {
	x_, y_ := pos[0], pos[1]
	for y := y_; y < int(y_+int(src.Size[1])); y++ {
		copy(d.Datas[y*int(d.Size[0])+x_:y*int(d.Size[0])+x_+int(src.Size[0])], src.Datas[(y-y_)*int(src.Size[0]):(y-y_+1)*int(src.Size[0])])
	}

	isCopyLeftBorder := (x_ == 0) && d.HasBorder()
	if isCopyLeftBorder {
		copy(d.LeftBorder[y_:y_+int(src.Size[1])], src.LeftBorder)
	}

	isCopyRightBorder := (src.Size[0]+uint32(x_) == d.Size[0]) && d.IsBilateral()
	if isCopyRightBorder {
		copy(d.RightBorder[y_:y_+int(src.Size[1])], src.RightBorder)
	}

	isCopyTopBorder := (y_ == 0) && d.HasBorder()
	if isCopyTopBorder {
		offx := 1
		if x_ == 0 {
			offx = 0
		}
		copy(d.TopBorder[offx+x_:offx+x_+int(src.Size[0])], src.TopBorder[offx:])
	}

	isCopyBottomBorder := (src.Size[1]+uint32(y_) == d.Size[1]) && d.IsBilateral()
	if isCopyBottomBorder {
		offx := 1
		if x_ == 0 {
			offx = 0
		}
		copy(d.BottomBorder[offx+x_:offx+x_+int(src.Size[0])], src.BottomBorder[offx:])
	}
}

type RasterIO interface {
	Decode(r io.Reader) (*TileData, error)
	Encode(tile *TileData) ([]byte, error)
}

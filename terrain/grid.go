package terrain

import (
	"sort"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/go-tileproxy/geo"
)

type Coordinates []vec3d.T

func (s Coordinates) Len() int {
	return len(s)
}

func (s Coordinates) Less(i, j int) bool {
	if s[i][1] == s[j][1] {
		return s[i][0] < s[j][0]
	} else {
		return s[i][1] < s[j][1]
	}
}

func (s Coordinates) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Grid struct {
	Width       int
	Height      int
	Coordinates Coordinates
	Count       int
	Minimum     float64
	Maximum     float64
	Border      BorderMode
	box         *vec3d.Box
	srs         geo.Proj
}

func NewGrid(width, height int, mode BorderMode) *Grid {
	off := 0
	if mode == BORDER_UNILATERAL {
		off = 1
	} else if mode == BORDER_BILATERAL {
		off = 2
	}
	return &Grid{Width: width + off, Height: height + off, Border: mode, Count: width * height, Minimum: 15000, Maximum: -15000}
}

func caclulatePixelSize(width, height int, bbox vec2d.Rect) []float64 {
	pixelSize := []float64{0, 0}
	pixelSize[0] = (bbox.Max[0] - bbox.Min[0]) / float64(width)
	pixelSize[1] = (bbox.Max[1] - bbox.Min[1]) / float64(height)
	return pixelSize
}

func CaclulateGrid(width, height int, mode BorderMode, georef *geo.GeoReference) *Grid {
	grid := NewGrid(width, height, mode)

	grid.Count = grid.Width * grid.Height
	grid.srs = georef.GetSrs()

	coords := make(Coordinates, 0, grid.Count)

	pixelSize := caclulatePixelSize(grid.Width, grid.Height, georef.GetBBox())

	if mode == BORDER_UNILATERAL || mode == BORDER_BILATERAL {
		for y := grid.Height - 1; y >= 0; y-- {
			latitude := georef.GetOrigin()[1] + (float64(pixelSize[1]) * float64(y-1))
			for x := 0; x < grid.Width; x++ {
				longitude := georef.GetOrigin()[0] + (float64(pixelSize[0]) * float64(x-1))
				coords = append(coords, vec3d.T{longitude, latitude, 0})
			}
		}
	} else {
		for y := grid.Height - 1; y >= 0; y-- {
			latitude := georef.GetOrigin()[1] + (float64(pixelSize[1]) * float64(y))
			for x := 0; x < grid.Width; x++ {
				longitude := georef.GetOrigin()[0] + (float64(pixelSize[0]) * float64(x))
				coords = append(coords, vec3d.T{longitude, latitude, 0})
			}
		}
	}
	grid.Coordinates = coords
	return grid
}

func (h *Grid) GetRect() vec2d.Rect {
	bbox := h.GetBBox()
	return vec2d.Rect{Min: vec2d.T{bbox.Min[0], bbox.Min[1]}, Max: vec2d.T{bbox.Max[0], bbox.Max[1]}}
}

func (h *Grid) GetBBox() vec3d.Box {
	if h.box == nil {
		r := vec3d.Box{Min: vec3d.MaxVal, Max: vec3d.MinVal}
		for i := range h.Coordinates {
			r.Extend(&h.Coordinates[i])
		}
		return r
	}
	return *h.box
}

func (h *Grid) GetRange() float64 {
	return h.Maximum - h.Minimum
}

func (h *Grid) Sort() {
	sort.Sort(h.Coordinates)
}

func (h *Grid) Value(row, column int) float64 {
	return h.Coordinates[row*h.Width+column][2]
}

func (h *Grid) getBBox(mode BorderMode) vec3d.Box {
	row, col := h.Height, h.Width
	if mode == BORDER_UNILATERAL {
		r := vec3d.Box{Min: vec3d.MaxVal, Max: vec3d.MinVal}
		for x := 1; x < col; x++ {
			for y := 1; y < row; y++ {
				r.Extend(&h.Coordinates[y*col+x])
			}
		}
		return r
	} else if mode == BORDER_BILATERAL {
		r := vec3d.Box{Min: vec3d.MaxVal, Max: vec3d.MinVal}
		for x := 1; x < col-1; x++ {
			for y := 1; y < row-1; y++ {
				r.Extend(&h.Coordinates[y*col+x])
			}
		}
		return r
	} else {
		return h.GetBBox()
	}
}

func (h *Grid) GetTileDate(mode BorderMode) *TileData {
	off := 0
	if mode == BORDER_UNILATERAL {
		off = 1
	} else if mode == BORDER_BILATERAL {
		off = 2
	}

	tiledata := NewTileData([2]uint32{uint32(h.Width - off), uint32(h.Height - off)}, mode)

	if h.box != nil {
		bbox := h.getBBox(mode)
		tiledata.Box = vec2d.Rect{Min: vec2d.T{bbox.Min[0], bbox.Min[1]}, Max: vec2d.T{bbox.Max[0], bbox.Max[1]}}
		tiledata.Boxsrs = h.srs
	}

	row, col := h.Height, h.Width

	if mode == BORDER_UNILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 {
					tiledata.Set(x-1, y-1, h.Value(y, x))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, h.Value(y, x))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, h.Value(y, x))
				}
			}
		}
	} else if mode == BORDER_BILATERAL {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				if x > 0 && y > 0 && x < col-1 && y < row-1 {
					tiledata.Set(x-1, y-1, h.Value(y, x))
				}

				if x == 0 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_LEFT, y-1, h.Value(y, x))
				}

				if x == col-1 && y != 0 && y != row-1 {
					tiledata.FillBorder(BORDER_RIGHT, y-1, h.Value(y, x))
				}

				if y == 0 {
					tiledata.FillBorder(BORDER_TOP, x, h.Value(y, x))
				}

				if y == row-1 {
					tiledata.FillBorder(BORDER_BOTTOM, x, h.Value(y, x))
				}
			}
		}
	} else {
		for x := 0; x < col; x++ {
			for y := 0; y < row; y++ {
				tiledata.Set(x, y, h.Value(y, x))
			}
		}
	}

	return tiledata
}

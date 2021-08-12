package raster

import (
	"sort"

	"github.com/flywave/go-tileproxy/geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
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

type DemGrid struct {
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

func NewDemGrid(width, height int, mode BorderMode) *DemGrid {
	off := 0
	if mode == BORDER_UNILATERAL {
		off = 1
	} else if mode == BORDER_BILATERAL {
		off = 2
	}
	return &DemGrid{Width: width + off, Height: height + off, Border: mode, Count: width * height, Minimum: 15000, Maximum: -15000}
}

func caclulatePixelSize(width, height int, georef *geo.GeoReference) []float64 {
	pixelSize := []float64{0, 0}
	pixelSize[0] = (georef.GetBBox().Max[0] - georef.GetBBox().Min[0]) / float64(width)
	pixelSize[1] = (georef.GetBBox().Max[1] - georef.GetBBox().Min[1]) / float64(height)
	return pixelSize
}

func CaclulateDemGrid(width, height int, mode BorderMode, georef *geo.GeoReference) *DemGrid {
	grid := NewDemGrid(width, height, mode)

	grid.Count = grid.Width * grid.Height
	grid.srs = georef.GetSrs()

	coords := make(Coordinates, 0, grid.Count)

	pixelSize := caclulatePixelSize(grid.Width, grid.Height, georef)

	if mode == BORDER_UNILATERAL || mode == BORDER_BILATERAL {
		for y := 0; y < grid.Height; y++ {
			latitude := georef.GetOrigin()[1] + (float64(pixelSize[1]) * float64(y-1))
			for x := 0; x < grid.Width; x++ {
				longitude := georef.GetOrigin()[0] + (float64(pixelSize[0]) * float64(x-1))
				coords = append(coords, vec3d.T{latitude, longitude, 0})
			}
		}
	} else {
		for y := 0; y < grid.Height; y++ {
			latitude := georef.GetOrigin()[1] + (float64(pixelSize[1]) * float64(y))
			for x := 0; x < grid.Width; x++ {
				longitude := georef.GetOrigin()[0] + (float64(pixelSize[0]) * float64(x))
				coords = append(coords, vec3d.T{latitude, longitude, 0})
			}
		}
	}
	grid.Coordinates = coords
	return grid
}

func (h *DemGrid) GetRect() vec2d.Rect {
	bbox := h.GetBBox()
	return vec2d.Rect{Min: vec2d.T{bbox.Min[0], bbox.Min[1]}, Max: vec2d.T{bbox.Max[0], bbox.Max[1]}}
}

func (h *DemGrid) GetBBox() vec3d.Box {
	if h.box == nil {
		r := vec3d.Box{}
		for i := range h.Coordinates {
			r.Extend(&h.Coordinates[i])
		}
		return r
	}
	return *h.box
}

func (h *DemGrid) GetRange() float64 {
	return h.Maximum - h.Minimum
}

func (h *DemGrid) Sort() {
	sort.Sort(h.Coordinates)
}

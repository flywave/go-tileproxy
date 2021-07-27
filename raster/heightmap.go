package raster

import (
	"sort"

	mat4d "github.com/flywave/go3d/float64/mat4"
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

type HeightMap struct {
	Width       int
	Height      int
	Coordinates Coordinates
	Count       int
	Minimum     float64
	Maximum     float64
	box         *vec3d.Box
}

func NewHeightMap(width, height int) *HeightMap {
	return &HeightMap{Width: width, Height: height, Count: width * height, Minimum: 15000, Maximum: -15000}
}

func (h *HeightMap) GetBoundingBox() vec3d.Box {
	if h.box == nil {
		r := vec3d.Box{}
		for i := range h.Coordinates {
			r.Extend(&h.Coordinates[i])
		}
		return r
	}
	return *h.box
}

func (h *HeightMap) GetCoordinates() Coordinates {
	return h.Coordinates
}

func (h *HeightMap) GetRange() float64 {
	return h.Maximum - h.Minimum
}

func (h *HeightMap) GetMinimum() float64 {
	return h.Minimum
}

func (h *HeightMap) GetMaximum() float64 {
	return h.Maximum
}

func (h *HeightMap) GetWidth() int {
	return h.Width
}

func (h *HeightMap) GetHeight() int {
	return h.Height
}

func (h *HeightMap) GetCount() int {
	return h.Count
}

func TranslateCoordinates(coords []vec3d.T, offset vec3d.T) []vec3d.T {
	out := make([]vec3d.T, len(coords))
	for i := range coords {
		out[i].Add(&offset)
	}
	return out
}

func (h *HeightMap) CenterOnOriginWithTransform() (mat mat4d.T) {
	bbox := h.GetBoundingBox()

	xOriginOffset := bbox.Max[0] - (bbox.Max[0]-bbox.Min[0])/2.0
	yOriginOffset := bbox.Max[1] - (bbox.Max[1]-bbox.Min[1])/2.0
	zOriginOffset := bbox.Max[2] - (bbox.Max[2]-bbox.Min[2])/2.0

	h.Coordinates = TranslateCoordinates(h.Coordinates, vec3d.T{-xOriginOffset, -yOriginOffset, -zOriginOffset})

	tran := vec3d.T{-xOriginOffset, 0, yOriginOffset}
	mat.Translate(&tran)

	h.box = &vec3d.Box{Min: vec3d.T{bbox.Min[0] - xOriginOffset, bbox.Min[1] - yOriginOffset, bbox.Min[2] - zOriginOffset},
		Max: vec3d.T{bbox.Max[0] - xOriginOffset, bbox.Max[1] - yOriginOffset, bbox.Max[2] - zOriginOffset}}

	return
}

func (h *HeightMap) CenterOnOriginWithOrigin(origin vec3d.T) {
	bbox := h.GetBoundingBox()

	h.Coordinates = TranslateCoordinates(h.Coordinates, vec3d.T{-origin[0], -origin[1], -origin[2]})

	h.box = &vec3d.Box{Min: vec3d.T{bbox.Min[0] - origin[0], bbox.Min[1] - origin[1], bbox.Min[2] - origin[2]},
		Max: vec3d.T{bbox.Max[0] - origin[0], bbox.Max[1] - origin[1], bbox.Max[2] - origin[2]}}
}

func (h *HeightMap) CenterOnOrigin(centerOnZ bool) {
	bbox := h.GetBoundingBox()

	xOriginOffset := bbox.Max[0] - (bbox.Max[0]-bbox.Min[0])/2.0
	yOriginOffset := bbox.Max[1] - (bbox.Max[1]-bbox.Min[1])/2.0
	zOriginOffset := 0.0
	if centerOnZ {
		zOriginOffset = -bbox.Min[2]
	}

	h.Coordinates = TranslateCoordinates(h.Coordinates, vec3d.T{-xOriginOffset, -yOriginOffset, -zOriginOffset})

	h.box = &vec3d.Box{Min: vec3d.T{bbox.Min[0] - xOriginOffset, bbox.Min[1] - yOriginOffset, 0},
		Max: vec3d.T{bbox.Max[0] - xOriginOffset, bbox.Max[1] - yOriginOffset, bbox.Max[2] - bbox.Min[2]}}
}

func (h *HeightMap) CenterOnOriginWithBox(bbox vec3d.Box, centerOnZ bool) {
	xOriginOffset := bbox.Max[0] - (bbox.Max[0]-bbox.Min[0])/2.0
	yOriginOffset := bbox.Max[1] - (bbox.Max[1]-bbox.Min[1])/2.0

	zOriginOffset := 0.0
	if centerOnZ {
		zOriginOffset = -bbox.Min[2]
	}
	h.Coordinates = TranslateCoordinates(h.Coordinates, vec3d.T{-xOriginOffset, -yOriginOffset, -zOriginOffset})

	h.box = &vec3d.Box{Min: vec3d.T{bbox.Min[0] - xOriginOffset, bbox.Min[1] - yOriginOffset, 0},
		Max: vec3d.T{bbox.Max[0] - xOriginOffset, bbox.Max[1] - yOriginOffset, bbox.Max[2] - bbox.Min[2]}}
}

func CoordinatesZScale(coords []vec3d.T, zScale float64) []vec3d.T {
	for i := range coords {
		pos := coords[i]
		pos[2] *= zScale
		coords[i] = pos
	}
	return coords
}

func CoordinatesScale(coords []vec3d.T, x float64, y float64, z float64) []vec3d.T {
	for i := range coords {
		pos := coords[i]
		pos[0] *= x
		pos[1] *= y
		pos[2] *= z
		coords[i] = pos
	}
	return coords
}

func (h *HeightMap) ZScale(zFactor float64) {
	h.Coordinates = CoordinatesZScale(h.Coordinates, zFactor)
	h.Minimum *= zFactor
	h.Maximum *= zFactor
}

func BoxScale(box *vec3d.Box, x float64, y float64, z float64) *vec3d.Box {
	box.Min[0] *= x
	box.Min[1] *= y
	box.Min[2] *= z

	box.Max[0] *= x
	box.Max[1] *= y
	box.Max[2] *= z

	return box
}

func (h *HeightMap) Scale(x float64, y float64, z float64) {
	h.Coordinates = CoordinatesScale(h.Coordinates, x, y, z)
	h.box = BoxScale(h.box, x, y, z)
	h.Minimum *= z
	h.Maximum *= z
}

func (h *HeightMap) FitInto(maxSize float64) float64 {
	scale := 1.0
	boxWidth := h.box.Max[0] - h.box.Min[0]
	boxHeight := h.box.Max[1] - h.box.Min[1]

	if boxWidth > boxHeight {
		scale = (maxSize / boxWidth)
	} else {
		scale = (maxSize / boxHeight)
	}
	h.Coordinates = CoordinatesScale(h.Coordinates, scale, scale, scale)
	h.box = BoxScale(h.box, scale, scale, scale)
	return scale
}

func (h *HeightMap) ZTranslate(distance float64) {
	h.Coordinates = TranslateCoordinates(h.Coordinates, vec3d.T{0, 0, distance})
	h.Minimum += distance
	h.Maximum += distance
}

func (h *HeightMap) Translate(pt vec3d.T) {
	h.Coordinates = TranslateCoordinates(h.Coordinates, pt)
	h.Minimum += pt[2]
	h.Maximum += pt[2]
}

func (h *HeightMap) Sort() {
	sort.Sort(h.Coordinates)
}

func DownsampleCoordinates(input Coordinates, w, h, step int) Coordinates {
	ret := make(Coordinates, 0)
	for lat := 0; lat <= h; lat += step {
		for lon := 0; lon <= w; lon += step {
			ret = append(ret, input[lon+lat*h])
		}
	}
	return ret
}

func (h *HeightMap) Downsample(step int) *HeightMap {
	if step == 0 || step%2 != 0 {
		return nil
	}

	hMap := NewHeightMap(h.Width/step, h.Height/step)

	hMap.Maximum = h.Maximum
	hMap.Minimum = h.Minimum
	hMap.box = h.box
	hMap.Coordinates = DownsampleCoordinates(h.Coordinates, h.Width, h.Height, step)

	return hMap
}

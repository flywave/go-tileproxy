package setting

import (
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/go-tileproxy/tile"
)

func NewFloat64(f float64) *float64 {
	return &f
}

func NewBool(f bool) *bool {
	return &f
}

func NewString(s string) *string {
	return &s
}

func NewInt(s int) *int {
	return &s
}

func NewFormat(s tile.TileFormat) *tile.TileFormat {
	return &s
}

func NewRect(s vec2d.Rect) *vec2d.Rect {
	return &s
}

func NewVec2d(s vec2d.T) *vec2d.T {
	return &s
}

func NewVec3d(s vec3d.T) *vec3d.T {
	return &s
}

func NewBBox(s vec3d.Box) *vec3d.Box {
	return &s
}

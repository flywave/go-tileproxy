package setting

import (
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/go-tileproxy/tile"
)

func NewFloat32(f float32) *float32 {
	return &f
}

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

func NewInt32(s int32) *int32 {
	return &s
}

func NewUInt32(s uint32) *uint32 {
	return &s
}

func NewIn64(s int64) *int64 {
	return &s
}

func NewUInt64(s uint64) *uint64 {
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

func NewUInt322(s [2]uint32) *[2]uint32 {
	return &s
}

func NewInt2(i [2]int) *[2]int {
	return &i
}

func NewUInt82(i [4]uint8) *[4]uint8 {
	return &i
}

func NewFloat644(f [4]float64) *[4]float64 {
	return &f
}

func NewDuration(t time.Duration) *time.Duration {
	return &t
}

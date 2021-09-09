package ray

import (
	"fmt"
	"math"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

const (
	Epsilon float64 = 1e-9
	Inf     float64 = 1e99
)

type Ray struct {
	Start     vec3d.T
	Direction vec3d.T
	Depth     int
	Inverse   [3]float64
}

func New(start vec3d.T, direction vec3d.T, depth int) *Ray {
	return &Ray{Start: start, Direction: direction, Depth: depth}
}

func (r *Ray) String() string {
	return fmt.Sprintf("%s -> %s", &r.Start, &r.Direction)
}

func (r *Ray) Init() {
	for i := 0; i < 3; i++ {
		r.Inverse[i] = 0
	}
	if math.Abs(r.Direction[0]) > Epsilon {
		r.Inverse[0] = 1.0 / r.Direction[0]
	}
	if math.Abs(r.Direction[1]) > Epsilon {
		r.Inverse[1] = 1.0 / r.Direction[1]
	}
	if math.Abs(r.Direction[2]) > Epsilon {
		r.Inverse[2] = 1.0 / r.Direction[2]
	}
}

type Intersection struct {
	Point    *vec3d.T
	Incoming *Ray
	Distance float64
}

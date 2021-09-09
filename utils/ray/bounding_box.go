package ray

import (
	"math"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

const (
	Ox = iota
	Oy
	Oz
	Leaf
)

func Between(min, max, point float64) bool {
	return (min-Epsilon <= point && max+Epsilon >= point)
}

type BoundingBox struct {
	MaxVolume, MinVolume [3]float64
}

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{
		MinVolume: [3]float64{Inf, Inf, Inf},
		MaxVolume: [3]float64{-Inf, -Inf, -Inf},
	}
}

func (b *BoundingBox) AddPoint(point vec3d.T) {
	b.MinVolume[0] = math.Min(b.MinVolume[0], point[0])
	b.MinVolume[1] = math.Min(b.MinVolume[1], point[1])
	b.MinVolume[2] = math.Min(b.MinVolume[2], point[2])

	b.MaxVolume[0] = math.Max(b.MaxVolume[0], point[0])
	b.MaxVolume[1] = math.Max(b.MaxVolume[1], point[1])
	b.MaxVolume[2] = math.Max(b.MaxVolume[2], point[2])
}

func (b *BoundingBox) Inside(point vec3d.T) bool {
	return (Between(b.MinVolume[0], b.MaxVolume[0], point[0]) &&
		Between(b.MinVolume[1], b.MaxVolume[1], point[1]) &&
		Between(b.MinVolume[2], b.MaxVolume[2], point[2]))
}

func otherAxes(axis int) (int, int) {
	var otherAxis1, otherAxis2 int
	if axis == 0 {
		otherAxis1 = 1
	} else {
		otherAxis1 = 0
	}
	if axis == 2 {
		otherAxis2 = 1
	} else {
		otherAxis2 = 2
	}
	return otherAxis1, otherAxis2
}

func (b *BoundingBox) IntersectAxis(ray *Ray, axis int) bool {
	directions := [3]float64{ray.Direction[0], ray.Direction[1], ray.Direction[2]}
	start := [3]float64{ray.Start[0], ray.Start[1], ray.Start[2]}

	if (directions[axis] > 0 && start[axis] > b.MaxVolume[axis]) ||
		(directions[axis] < 0 && start[axis] < b.MinVolume[axis]) {
		return false
	}

	if math.Abs(directions[axis]) < Epsilon {
		return false
	}

	otherAxis1, otherAxis2 := otherAxes(axis)

	multiplier := ray.Inverse[axis]
	var intersectionX, intersectionY float64

	distance := (b.MinVolume[axis] - start[axis]) * multiplier
	if distance < 0 {
		return false
	}

	intersectionX = start[otherAxis1] + directions[otherAxis1]*distance
	if Between(b.MinVolume[otherAxis1], b.MaxVolume[otherAxis1], intersectionX) {
		intersectionY = start[otherAxis2] + directions[otherAxis2]*distance
		if Between(b.MinVolume[otherAxis2], b.MaxVolume[otherAxis2], intersectionY) {
			return true
		}
	}

	distance = (b.MaxVolume[axis] - start[axis]) * multiplier
	if distance < 0 {
		return false
	}
	intersectionX = start[otherAxis1] + directions[otherAxis1]*distance
	if Between(b.MinVolume[otherAxis1], b.MaxVolume[otherAxis1], intersectionX) {
		intersectionY = start[otherAxis2] + directions[otherAxis2]*distance
		if Between(b.MinVolume[otherAxis2], b.MaxVolume[otherAxis2], intersectionY) {
			return true
		}
	}
	return false
}

func (b *BoundingBox) Intersect(ray *Ray) bool {
	if b.Inside(ray.Start) {
		return true
	}
	return (b.IntersectAxis(ray, Ox) || b.IntersectAxis(ray, Oy) || b.IntersectAxis(ray, Oz))
}

func (b *BoundingBox) IntersectTriangle(A, B, C vec3d.T) bool {
	if b.Inside(A) || b.Inside(B) || b.Inside(C) {
		return true
	}
	triangle := [3]vec3d.T{A, B, C}
	ray := &Ray{}
	for rayStart := 0; rayStart < 3; rayStart++ {
		for rayEnd := rayStart + 1; rayEnd < 3; rayEnd++ {
			ray.Start = triangle[rayStart]
			ray.Direction = vec3d.Sub(&triangle[rayEnd], &triangle[rayStart])
			ray.Init()
			if b.Intersect(ray) {
				ray.Start = triangle[rayEnd]
				ray.Direction = vec3d.Sub(&triangle[rayStart], &triangle[rayEnd])
				ray.Init()
				if b.Intersect(ray) {
					return true
				}
			}
		}
	}

	AB := vec3d.Sub(&B, &A)
	AC := vec3d.Sub(&C, &A)
	ABxAC := vec3d.Cross(&AB, &AC)
	distance := vec3d.Dot(&A, &ABxAC)
	rayEnd := &vec3d.T{}

	for edgeMask := 0; edgeMask < 7; edgeMask++ {
		for axis := uint(0); axis < 3; axis++ {
			if edgeMask&(1<<axis) != 0 {
				continue
			}

			if edgeMask&1 != 0 {
				ray.Start[0] = b.MaxVolume[0]
			} else {
				ray.Start[0] = b.MinVolume[0]
			}

			if edgeMask&2 != 0 {
				ray.Start[1] = b.MaxVolume[1]
			} else {
				ray.Start[1] = b.MinVolume[1]
			}

			if edgeMask&4 != 0 {
				ray.Start[2] = b.MaxVolume[2]
			} else {
				ray.Start[2] = b.MinVolume[2]
			}

			*rayEnd = ray.Start
			SetDimension(rayEnd, int(axis), b.MaxVolume[axis])

			if (vec3d.Dot(&ray.Start, &ABxAC)-distance)*(vec3d.Dot(rayEnd, &ABxAC)-distance) <= 0 {
				ray.Direction = vec3d.Sub(rayEnd, &ray.Start)
				ray.Init()
				intersected, distance := IntersectTriangle(ray, &A, &B, &C)
				if intersected && distance < 1.0000001 {
					return true
				}
			}
		}
	}

	return false
}

func SetDimension(v *vec3d.T, axis int, value float64) {
	switch axis {
	case Ox:
		v[0] = value
	case Oy:
		v[1] = value
	case Oz:
		v[2] = value
	}
}

func GetDimension(v *vec3d.T, axis int) float64 {
	switch axis {
	case Ox:
		return v[0]
	case Oy:
		return v[1]
	case Oz:
		return v[2]
	}
	return Inf
}

func (b *BoundingBox) Split(axis int, median float64) (*BoundingBox, *BoundingBox) {
	left := &BoundingBox{}
	*left = *b
	right := &BoundingBox{}
	*right = *b
	left.MaxVolume[axis] = median
	right.MinVolume[axis] = median
	return left, right
}

func (b *BoundingBox) IntersectWall(axis int, median float64, ray *Ray) bool {
	directions := [3]float64{ray.Direction[0], ray.Direction[1], ray.Direction[2]}
	start := [3]float64{ray.Start[0], ray.Start[1], ray.Start[2]}
	if math.Abs(directions[axis]) < Epsilon {
		return (math.Abs(start[axis]-median) < Epsilon)
	}

	otherAxis1, otherAxis2 := otherAxes(axis)
	distance := median - start[axis]
	directionInAxis := ray.Inverse[axis]

	if (distance * directionInAxis) < 0 {
		return false
	}

	fac := distance * directionInAxis
	distanceOnAxis1 := start[otherAxis1] +
		directions[otherAxis1]*fac
	if b.MinVolume[otherAxis1] <= distanceOnAxis1 &&
		distanceOnAxis1 <= b.MaxVolume[otherAxis1] {

		distanceOnAxis2 := start[otherAxis2] +
			directions[otherAxis2]*fac
		return b.MinVolume[otherAxis2] <= distanceOnAxis2 &&
			distanceOnAxis2 <= b.MaxVolume[otherAxis2]
	}
	return false
}

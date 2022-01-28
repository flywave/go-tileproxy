package draw

import "math"

const (
	Radian = 1
	Degree = (math.Pi / 180) * Radian
)

func DegreesToRadians(d float64) float64 {
	return float64(d * Degree)
}

func RadiansToDegrees(a float64) float64 {
	return float64(a / Degree)
}

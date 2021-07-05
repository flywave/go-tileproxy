package geo

import (
	"fmt"
	"math"

	"errors"

	"github.com/flywave/go-tileproxy/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

const (
	WebMercator = 3857
	WGS84       = tile.WGS84
	Deg2Rad     = math.Pi / 180
	Rad2Deg     = 180 / math.Pi
	PiDiv2      = math.Pi / 2.0
	PiDiv4      = math.Pi / 4.0
)

type Pt struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (pt Pt) XCoord() float64   { return pt.X }
func (pt Pt) YCoord() float64   { return pt.X }
func (pt Pt) Coords() []float64 { return []float64{pt.X, pt.Y} }

func (pt Pt) IsEqual(pt2 Pt) bool {
	return pt.X == pt2.X && pt.Y == pt2.Y
}

func (pt Pt) Truncate() Pt {
	return Pt{
		X: math.Trunc(pt.X),
		Y: math.Trunc(pt.Y),
	}
}

func (pt Pt) Round() Pt {
	return Pt{
		math.Round(pt.X),
		math.Round(pt.Y),
	}
}

func (pt Pt) Delta(pt2 Pt) (d Pt) {
	return Pt{
		X: pt.X - pt2.X,
		Y: pt.Y - pt2.Y,
	}
}

func (pt Pt) String() string {
	return fmt.Sprintf("{%v,%v}", pt.X, pt.Y)
}

func (pt *Pt) GoString() string {
	if pt == nil {
		return "(nil)"
	}
	return fmt.Sprintf("[%v,%v]", pt.X, pt.Y)
}

type Pointer interface {
	Point() Pt
}

func NewPoints(f []float64) (pts []Pt, err error) {
	if len(f)%2 != 0 {
		return pts, errors.New("Expected even number of points.")
	}
	for x, y := 0, 1; y < len(f); x, y = x+2, y+2 {
		pts = append(pts, Pt{f[x], f[y]})
	}
	return pts, nil
}

func AreaOfRing(points ...Pt) (area float64) {
	n := len(points)
	for i := range points {
		j := (i + 1) % n
		area += points[i].X * points[j].Y
		area -= points[j].X * points[i].Y
	}
	return math.Abs(area) / 2.0
}

func RadToDeg(rad float64) float64 {
	return rad * Rad2Deg
}

func DegToRad(deg float64) float64 {
	return deg * Deg2Rad
}

func XYOrder(pt1, pt2 Pt) int {
	switch {
	case pt1.X > pt2.X:
		return 1
	case pt1.X < pt2.X:
		return -1
	case pt1.Y > pt2.Y:
		return 1
	case pt1.Y < pt2.Y:
		return -1
	}
	return 0
}

func YXorder(pt1, pt2 Pt) int {
	switch {
	case pt1.Y > pt2.Y:
		return 1
	case pt1.Y < pt2.Y:
		return -1
	case pt1.X > pt2.X:
		return 1
	case pt1.X < pt2.X:
		return -1
	}
	return 0
}

func Exp2(p uint64) uint64 {
	if p > 63 {
		p = 63
	}
	return uint64(1) << p
}

func Min(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinUInt32(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

func MaxUInt32(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func NewBool(v bool) *bool {
	return &v
}

func NewFloat64(v float64) *float64 {
	return &v
}

func NewInt(v int) *int {
	return &v
}

func NewRect(r vec2d.Rect) *vec2d.Rect {
	return &r
}

type roundMode byte

const (
	closest roundMode = iota
	down
	up
)

func round(number, digits float64, mode roundMode) float64 {
	var significance float64
	if digits > 0 {
		significance = math.Pow(1/10.0, digits)
	} else {
		significance = math.Pow(10.0, -digits)
	}
	val, res := math.Modf(number / significance)
	switch mode {
	case closest:
		const eps = 0.499999999
		if res >= eps {
			val++
		} else if res <= -eps {
			val--
		}
	case down:
	case up:
		if res > 0 {
			val++
		} else if res < 0 {
			val--
		}
	}
	return val * significance
}

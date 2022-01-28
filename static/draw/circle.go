package draw

import (
	"image/color"
	"log"
	"math"
	"strings"

	"strconv"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flopp/go-coordsparser"
	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/utils"
)

type Circle struct {
	MapObject
	Position vec2d.T
	Srs      geo.Proj
	Color    color.Color
	Fill     color.Color
	Weight   float64
	Radius   float64
}

func NewCircle(pos vec2d.T, srs geo.Proj, col, fill color.Color, radius, weight float64) *Circle {
	return &Circle{
		Position: pos,
		Srs:      srs,
		Color:    col,
		Fill:     fill,
		Weight:   weight,
		Radius:   radius,
	}
}

func ParseCircleString(s string) (circles []*Circle, err error) {
	circles = make([]*Circle, 0)

	var col color.Color = color.RGBA{0xff, 0, 0, 0xff}
	var fill color.Color = color.Transparent
	radius := 100.0
	weight := 5.0

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			col, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "fill:"); ok {
			fill, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "radius:"); ok {
			if radius, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "weight:"); ok {
			if weight, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else {
			lat, lng, err := coordsparser.Parse(ss)
			if err != nil {
				return nil, err
			}
			c := NewCircle(vec2d.T{lat, lng}, geo.NewProj("EPSG:4326"), col, fill, radius, weight)
			circles = append(circles, c)
		}
	}
	return circles, nil
}

func (m *Circle) getLatLng(plus bool) *vec2d.T {
	const (
		R = 6371000.0
	)
	th := m.Radius / R
	br := 0 / float64(Degree)
	if !plus {
		th *= -1
	}
	lat := DegreesToRadians(m.Position[0])
	lat1 := math.Asin(math.Sin(lat)*math.Cos(th) + math.Cos(lat)*math.Sin(th)*math.Cos(br))
	lng1 := DegreesToRadians(m.Position[1]) +
		math.Atan2(math.Sin(br)*math.Sin(th)*math.Cos(lat),
			math.Cos(th)-math.Sin(lat)*math.Sin(lat1))
	return &vec2d.T{
		RadiansToDegrees(lat1),
		RadiansToDegrees(lng1),
	}
}

func (m *Circle) ExtraMarginPixels() (float64, float64, float64, float64) {
	return m.Weight, m.Weight, m.Weight, m.Weight
}

func (m *Circle) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	r.Extend(m.getLatLng(false))
	r.Extend(m.getLatLng(true))
	return r
}

func (m *Circle) SrsProj() geo.Proj {
	return m.Srs
}

func (m *Circle) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("Circle coordinates not displayable: %f/%f", m.Position[0], m.Position[1])
		return
	}

	ll := m.getLatLng(true)
	x, y := trans.LatLngToXY(m.Position, m.Srs)
	x1, y1 := trans.LatLngToXY(*ll, m.Srs)
	radius := math.Sqrt(math.Pow(x1-x, 2) + math.Pow(y1-y, 2))
	gc.ClearPath()
	gc.SetLineWidth(m.Weight)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	gc.DrawCircle(x, y, radius)
	gc.SetColor(m.Fill)
	gc.FillPreserve()
	gc.SetColor(m.Color)
	gc.Stroke()
}

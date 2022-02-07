package draw

import (
	"image/color"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/utils"
)

type Area struct {
	MapObject
	Positions []vec2d.T
	Srs       geo.Proj
	Color     color.Color
	Fill      color.Color
	Weight    float64
}

func NewArea(positions []vec2d.T, srs geo.Proj, col color.Color, fill color.Color, weight float64) *Area {
	a := new(Area)
	a.Positions = positions
	a.Srs = srs
	a.Color = col
	a.Fill = fill
	a.Weight = weight
	return a
}

func ParseAreaString(s string) (*Area, error) {
	area := new(Area)
	area.Color = color.RGBA{0xff, 0, 0, 0xff}
	area.Fill = color.Transparent
	area.Weight = 5.0

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			area.Color, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "fill:"); ok {
			var err error
			area.Fill, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "weight:"); ok {
			var err error
			area.Weight, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			epsg, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
			area.Srs = geo.NewProj(int(epsg))
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			area.Positions = append(area.Positions, vec2d.T{lat, lng})
		}
	}

	if area.Srs == nil {
		area.Srs = geo.NewProj("EPSG:4326")
	}
	return area, nil
}

func (p *Area) ExtraMarginPixels() (float64, float64, float64, float64) {
	return p.Weight, p.Weight, p.Weight, p.Weight
}

func (p *Area) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	for _, ll := range p.Positions {
		r.Extend(&ll)
	}
	return r
}

func (p *Area) SrsProj() geo.Proj {
	return p.Srs
}

func (p *Area) Draw(gc *gg.Context, trans *Transformer) {
	if len(p.Positions) <= 1 {
		return
	}

	gc.ClearPath()
	gc.SetLineWidth(p.Weight)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	for _, ll := range p.Positions {
		gc.LineTo(trans.LatLngToXY(ll, p.Srs))
	}
	gc.ClosePath()
	gc.SetColor(p.Fill)
	gc.FillPreserve()
	gc.SetColor(p.Color)
	gc.Stroke()
}

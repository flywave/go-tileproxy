package draw

import (
	"bytes"
	"image/color"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flopp/go-coordsparser"
	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-gpx"
	"github.com/flywave/go-tileproxy/utils"
)

type Path struct {
	MapObject
	Positions []vec2d.T
	Srs       geo.Proj
	Color     color.Color
	Weight    float64
}

func NewPath(positions []vec2d.T, srs geo.Proj, col color.Color, weight float64) *Path {
	p := new(Path)
	p.Positions = positions
	p.Color = col
	p.Weight = weight
	p.Srs = srs

	return p
}

func ParsePathString(s string) ([]*Path, error) {
	paths := make([]*Path, 0)
	currentPath := new(Path)
	currentPath.Color = color.RGBA{0xff, 0, 0, 0xff}
	currentPath.Weight = 5.0

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			if currentPath.Color, err = ParseColorString(suffix); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "weight:"); ok {
			var err error
			if currentPath.Weight, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "gpx:"); ok {
			gpxData, err := gpx.Read(bytes.NewBufferString(suffix))
			if err != nil {
				return nil, err
			}
			for _, trk := range gpxData.Trk {
				for _, seg := range trk.TrkSeg {
					p := new(Path)
					p.Color = currentPath.Color
					p.Weight = currentPath.Weight
					for _, pt := range seg.TrkPt {
						p.Positions = append(p.Positions, vec2d.T{pt.Lat, pt.Lon})
					}
					if len(p.Positions) > 0 {
						paths = append(paths, p)
					}
				}
			}
		} else {
			lat, lng, err := coordsparser.Parse(ss)
			if err != nil {
				return nil, err
			}
			currentPath.Positions = append(currentPath.Positions, vec2d.T{lat, lng})
		}
	}
	if len(currentPath.Positions) > 0 {
		paths = append(paths, currentPath)
	}
	return paths, nil
}

func (p *Path) ExtraMarginPixels() (float64, float64, float64, float64) {
	return p.Weight, p.Weight, p.Weight, p.Weight
}

func (p *Path) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	for _, ll := range p.Positions {
		r.Extend(&ll)
	}
	return r
}

func (m *Path) SrsProj() geo.Proj {
	return m.Srs
}

func (p *Path) Draw(gc *gg.Context, trans *Transformer) {
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
	gc.SetColor(p.Color)
	gc.Stroke()
}

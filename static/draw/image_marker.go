package draw

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/utils"
)

type ImageMarker struct {
	MapObject
	Position vec2d.T
	Srs      geo.Proj
	Img      image.Image
	OffsetX  float64
	OffsetY  float64
}

func NewImageMarker(pos vec2d.T, srs geo.Proj, img image.Image, offsetX, offsetY float64) *ImageMarker {
	m := new(ImageMarker)
	m.Position = pos
	m.Srs = srs
	m.Img = img
	m.OffsetX = offsetX
	m.OffsetY = offsetY
	return m
}

func ParseImageMarkerString(s string) ([]*ImageMarker, error) {
	markers := make([]*ImageMarker, 0)

	var img image.Image = nil
	offsetX := 0.0
	offsetY := 0.0
	epsg := int64(4326)

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "image:"); ok {
			file, err := os.Open(suffix)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			img, _, err = image.Decode(file)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "offsetx:"); ok {
			var err error
			offsetX, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "offsety:"); ok {
			var err error
			offsetY, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			var err error
			epsg, err = strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			if img == nil {
				return nil, fmt.Errorf("cannot create an ImageMarker without an image: %s", s)
			}
			m := NewImageMarker(vec2d.T{lat, lng}, geo.NewProj(int(epsg)), img, offsetX, offsetY)
			markers = append(markers, m)
		}
	}
	return markers, nil
}

func (m *ImageMarker) SetImage(img image.Image) {
	m.Img = img
}

func (m *ImageMarker) SetOffsetX(offset float64) {
	m.OffsetX = offset
}

func (m *ImageMarker) SetOffsetY(offset float64) {
	m.OffsetY = offset
}

func (m *ImageMarker) ExtraMarginPixels() (float64, float64, float64, float64) {
	size := m.Img.Bounds().Size()
	return m.OffsetX, m.OffsetY, float64(size.X) - m.OffsetX, float64(size.Y) - m.OffsetY
}

func (m *ImageMarker) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	r.Extend(&m.Position)
	return r
}

func (m *ImageMarker) SrsProj() geo.Proj {
	return m.Srs
}

func (m *ImageMarker) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("ImageMarker coordinates not displayable: %f/%f", m.Position[0], m.Position[1])
		return
	}

	x, y := trans.LatLngToXY(m.Position, m.Srs)
	gc.DrawImage(m.Img, int(x-m.OffsetX), int(y-m.OffsetY))
}

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

	"github.com/flopp/go-coordsparser"
	"github.com/flywave/gg"
	"github.com/flywave/go-tileproxy/utils"
	"github.com/golang/geo/s2"
)

type ImageMarker struct {
	MapObject
	Position s2.LatLng
	Img      image.Image
	OffsetX  float64
	OffsetY  float64
}

func NewImageMarker(pos s2.LatLng, img image.Image, offsetX, offsetY float64) *ImageMarker {
	m := new(ImageMarker)
	m.Position = pos
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
		} else {
			lat, lng, err := coordsparser.Parse(ss)
			if err != nil {
				return nil, err
			}
			if img == nil {
				return nil, fmt.Errorf("cannot create an ImageMarker without an image: %s", s)
			}
			m := NewImageMarker(s2.LatLngFromDegrees(lat, lng), img, offsetX, offsetY)
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

func (m *ImageMarker) Bounds() s2.Rect {
	r := s2.EmptyRect()
	r = r.AddPoint(m.Position)
	return r
}

func (m *ImageMarker) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("ImageMarker coordinates not displayable: %f/%f", m.Position.Lat.Degrees(), m.Position.Lng.Degrees())
		return
	}

	x, y := trans.LatLngToXY(m.Position)
	gc.DrawImage(m.Img, int(x-m.OffsetX), int(y-m.OffsetY))
}

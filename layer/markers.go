package layer

import (
	"fmt"
)

type Marker struct {
	Name      string
	Label     string
	Color     string
	Latitude  float64
	Longitude float64
}

func NewMarker() *Marker {
	return &Marker{}
}

func (m *Marker) String() string {
	return fmt.Sprintf("%s-%s+%s(%f,%f)", m.Name, m.Label, m.Color, m.Longitude, m.Latitude)
}

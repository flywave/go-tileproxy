package geo

import (
	"testing"

	"github.com/flywave/go-geoid"
)

func TestGeoid(t *testing.T) {
	lat := float64(10)
	lon := float64(20)
	alt := float64(30)

	h84 := WGS84ToMSL(lon, lat, alt, geoid.EGM84)
	h08 := WGS84ToMSL(lon, lat, alt, geoid.EGM2008)
	h96 := WGS84ToMSL(lon, lat, alt, geoid.EGM96)

	if h84 == 0 || h08 == 0 || h96 == 0 {
		t.FailNow()
	}
}

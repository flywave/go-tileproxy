package maths

import (
	"fmt"
	"testing"

	"github.com/flywave/go-tileproxy/maths/gcj02"
	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	wgsLat, wgsLng float64
	gcjLat, gcjLng float64
}{
	{31.1774276, 121.5272106, 31.17530398364597, 121.531541859215}, // shanghai
	{22.543847, 113.912316, 22.540796131694766, 113.9171764808363}, // shenzhen
	{39.911954, 116.377817, 39.91334545536069, 116.38404722455657}, // beijing
}

func toString(lat, lng float64) string {
	return fmt.Sprintf("%.5f,%.5f", lat, lng)
}

func TestGCJ02(t *testing.T) {
	for i, test := range tests {
		gcjLat, gcjLng := gcj02.WGS84toGCJ02(test.wgsLat, test.wgsLng)
		got := toString(gcjLat, gcjLng)
		target := toString(test.gcjLat, test.gcjLng)
		assert.Equal(t, got, target, "test %d", i)
	}
	for i, test := range tests {
		wgsLat, wgsLng := gcj02.GCJ02toWGS84(test.gcjLat, test.gcjLng)
		d := gcj02.Distance(wgsLat, wgsLng, test.wgsLat, test.wgsLng)
		assert.Equal(t, d < 5, true, "test %d, distance: %f", i, d)
	}
	for i, test := range tests {
		wgsLat, wgsLng := gcj02.GCJ02toWGS84Exact(test.gcjLat, test.gcjLng)
		d := gcj02.Distance(wgsLat, wgsLng, test.wgsLat, test.wgsLng)
		assert.Equal(t, d < 0.5, true, "test %d, distance: %f", i, d)
	}
}

func TestGCJ02Proj(t *testing.T) {
	srs4326 := NewSRSProj4("EPSG:4326")
	pgcj02 := NewGCJ02Proj(true)
	for i, test := range tests {
		pts := srs4326.TransformTo(pgcj02, []vec2d.T{{test.wgsLng, test.wgsLat}})

		got := toString(pts[0][1], pts[0][0])
		target := toString(test.gcjLat, test.gcjLng)
		assert.Equal(t, got, target, "test %d", i)
	}
}

func TestGCJ02To900913Proj(t *testing.T) {
	srs900913 := NewSRSProj4("EPSG:900913")
	srs4326 := NewSRSProj4("EPSG:4326")
	pgcj02 := NewGCJ02Proj(true)
	for i, test := range tests {
		p900913 := srs4326.TransformTo(srs900913, []vec2d.T{{test.wgsLng, test.wgsLat}})
		pts := srs900913.TransformTo(pgcj02, p900913)

		got := toString(pts[0][1], pts[0][0])
		target := toString(test.gcjLat, test.gcjLng)
		assert.Equal(t, got, target, "test %d", i)
	}

	for i, test := range tests {
		pts := pgcj02.TransformTo(srs4326, []vec2d.T{{test.gcjLng, test.gcjLat}})
		d := gcj02.Distance(pts[0][1], pts[0][0], test.wgsLat, test.wgsLng)
		assert.Equal(t, d < 0.5, true, "test %d, distance: %f", i, d)
	}
}

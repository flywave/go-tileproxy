package gcj02

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkBD09toGCJ02(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BD09toGCJ02(116.404, 39.915)
	}
}

func BenchmarkGCJ02toBD09(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GCJ02toBD09(116.404, 39.915)
	}
}

func BenchmarkWGS84toBD09(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WGS84toBD09(116.404, 39.915)
	}
}

func BenchmarkGCJ02toWGS84(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GCJ02toWGS84(116.404, 39.915)
	}
}

func BenchmarkBD09toWGS84(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BD09toWGS84(116.404, 39.915)
	}
}

func BenchmarkWGS84toGCJ02(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WGS84toGCJ02(116.404, 39.915)
	}
}

type cityCoord struct {
	Attrs  map[string]string     `json:"attrs,omitempty"`
	Coords map[string][2]float64 `json:"coords,omitempty"`
}

type cityCoords []cityCoord

func cityCoordsFromJson(data io.Reader) *cityCoords {
	var ts *cityCoords
	json.NewDecoder(data).Decode(&ts)
	return ts
}

func assertWGS84toGCJ02(city string, wgsLat, wgsLng, tgt_gcjLat, tgt_gcjLng float64, t *testing.T) {
	gcjLat, gcjLng := WGS84toGCJ02(wgsLat, wgsLng)
	d := Distance(gcjLat, gcjLng, tgt_gcjLat, tgt_gcjLng)
	assert.Equal(t, d < 1, true, "test %s, distance: %f", city, d)
}

func assertWGS84toBD09(city string, wgsLat, wgsLng, tgt_gcjLat, tgt_gcjLng float64, t *testing.T) {
	gcjLat, gcjLng := WGS84toBD09(wgsLat, wgsLng)
	d := Distance(gcjLat, gcjLng, tgt_gcjLat, tgt_gcjLng)
	assert.Equal(t, d < 1, true, "test %s, distance: %f", city, d)
}

func assertGCJ02toWGS84(city string, gcjLat, gcjLng, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := GCJ02toWGS84(gcjLat, gcjLng)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertGCJ02toWGS84Exact(city string, gcjLat, gcjLng, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := GCJ02toWGS84Exact(gcjLat, gcjLng)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 1, true, "test %s, distance: %f", city, d)
}

func assertBD09toWGS84Exact(city string, bdLat, bdLng, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := BD09toWGS84Exact(bdLat, bdLng)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 1, true, "test %s, distance: %f", city, d)
}

func assertWGS84toEPSG3857(city string, gcjLat, gcjLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := WGS84toEPSG3857(gcjLat, gcjLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 0.05 && diffX < 0.05
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertGCJ02toGCJ02MC(city string, gcjLat, gcjLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := GCJ02toGCJ02MC(gcjLat, gcjLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 0.05 && diffX < 0.05
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertBD09toBD09MC(city string, gcjLat, gcjLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := BD09toBDMC(gcjLat, gcjLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 0.05 && diffX < 0.05
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertWGS884toBD09MC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := WGS84toBDMC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertBMMCtoBD09(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := BDMCtoBD09(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertBMMCtoGCJ02(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := BDMCtoGCJ02(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertBMMCtoWGS84(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := BDMCtoWGS84(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertGCJ02MCtoWGS84(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := GCJ02MCtoWGS84(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertWGS884toGCJ02MC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := WGS84toGCJ02MC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertBD09toGCJ02MC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := BD09toGCJ02MC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertGCJ02MCtoBD09(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := GCJ02MCtoBD09(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func assertGCJ02MCtoBDMC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := GCJ02MCtoBDMC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertGCJ02toBDMC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := GCJ02toBDMC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertBDMCtoGCJ02MC(city string, wgsLat, wgsLng, tgt_mercartorY, tgt_mercartorX float64, t *testing.T) {
	mercartorY, mercartorX := BDMCtoGCJ02MC(wgsLat, wgsLng)
	diffY, diffX := tgt_mercartorY-mercartorY, tgt_mercartorX-mercartorX
	flag := diffY < 1 && diffX < 1
	assert.Equal(t, flag, true, "test %s, distance: %f %f", city, diffY, diffX)
}

func assertGCJ02MCtoGCJ02(city string, mercartorY, mercartorX, tgt_wgsLat, tgt_wgsLng float64, t *testing.T) {
	wgsLat, wgsLng := GCJ02MCtoGCJ02(mercartorY, mercartorX)
	d := Distance(wgsLat, wgsLng, tgt_wgsLat, tgt_wgsLng)
	assert.Equal(t, d < 4, true, "test %s, distance: %f", city, d)
}

func TestCoordinateConvert(t *testing.T) {
	r, _ := os.Open("./china-cities.json")

	citys := *cityCoordsFromJson(r)

	if citys == nil {
		t.FailNow()
	}

	for i := range citys {
		wgs := citys[i].Coords["WGS84"]
		gcj := citys[i].Coords["GCJ02"]
		bd := citys[i].Coords["BD09"]
		mc := citys[i].Coords["EPSG3857"]
		bdmc := citys[i].Coords["BD09MC"]
		gcjmc := citys[i].Coords["GCJ02MC"]

		assertWGS84toGCJ02(citys[i].Attrs["city"], wgs[1], wgs[0], gcj[1], gcj[0], t)
		assertGCJ02toWGS84(citys[i].Attrs["city"], gcj[1], gcj[0], wgs[1], wgs[0], t)
		assertGCJ02toWGS84Exact(citys[i].Attrs["city"], gcj[1], gcj[0], wgs[1], wgs[0], t)
		assertBD09toWGS84Exact(citys[i].Attrs["city"], bd[1], bd[0], wgs[1], wgs[0], t)
		assertWGS84toEPSG3857(citys[i].Attrs["city"], wgs[1], wgs[0], mc[1], mc[0], t)
		assertBD09toBD09MC(citys[i].Attrs["city"], bd[1], bd[0], bdmc[1], bdmc[0], t)
		assertGCJ02toGCJ02MC(citys[i].Attrs["city"], gcj[1], gcj[0], gcjmc[1], gcjmc[0], t)
		assertWGS84toBD09(citys[i].Attrs["city"], wgs[1], wgs[0], bd[1], bd[0], t)
		assertWGS884toBD09MC(citys[i].Attrs["city"], wgs[1], wgs[0], bdmc[1], bdmc[0], t)
		assertBMMCtoBD09(citys[i].Attrs["city"], bdmc[1], bdmc[0], bd[1], bd[0], t)
		assertBMMCtoGCJ02(citys[i].Attrs["city"], bdmc[1], bdmc[0], gcj[1], gcj[0], t)
		assertBMMCtoWGS84(citys[i].Attrs["city"], bdmc[1], bdmc[0], wgs[1], wgs[0], t)
		assertGCJ02MCtoWGS84(citys[i].Attrs["city"], gcjmc[1], gcjmc[0], wgs[1], wgs[0], t)
		assertWGS884toGCJ02MC(citys[i].Attrs["city"], wgs[1], wgs[0], gcjmc[1], gcjmc[0], t)
		assertBD09toGCJ02MC(citys[i].Attrs["city"], bd[1], bd[0], gcjmc[1], gcjmc[0], t)
		assertGCJ02MCtoBD09(citys[i].Attrs["city"], gcjmc[1], gcjmc[0], bd[1], bd[0], t)
		assertGCJ02MCtoBDMC(citys[i].Attrs["city"], gcjmc[1], gcjmc[0], bdmc[1], bdmc[0], t)
		assertBDMCtoGCJ02MC(citys[i].Attrs["city"], bdmc[1], bdmc[0], gcjmc[1], gcjmc[0], t)
		assertGCJ02toBDMC(citys[i].Attrs["city"], gcj[1], gcj[0], bdmc[1], bdmc[0], t)
		assertGCJ02MCtoGCJ02(citys[i].Attrs["city"], gcjmc[1], gcjmc[0], gcj[1], gcj[0], t)
	}
}

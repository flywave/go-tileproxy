package geo

import (
	"path/filepath"

	"github.com/flywave/go-geoid"
)

func init() {
	dir := getCurrentDir()
	geoid.SetGeoidPath(filepath.Join(dir, "../geoid_data"))
}

func geoid84_30() *geoid.Geoid {
	return geoid.NewGeoid(geoid.EGM84, false)
}

func geoid2008_5() *geoid.Geoid {
	return geoid.NewGeoid(geoid.EGM2008, false)
}

func geoid96_15() *geoid.Geoid {
	return geoid.NewGeoid(geoid.EGM96, false)
}

func getGeoid(g geoid.VerticalDatum) *geoid.Geoid {
	switch g {
	case geoid.EGM84:
		return geoid84_30()
	case geoid.EGM2008:
		return geoid2008_5()
	case geoid.EGM96:
		return geoid96_15()
	default:
		return geoid84_30()
	}
}

func MSLToWGS84(h, lon, lat float64, g geoid.VerticalDatum) float64 {
	return getGeoid(g).ConvertHeight(lat, lon, h, geoid.GEOIDTOELLIPSOID)
}

func WGS84ToMSL(lon, lat, altitude float64, g geoid.VerticalDatum) float64 {
	return getGeoid(g).ConvertHeight(lat, lon, altitude, geoid.ELLIPSOIDTOGEOID)
}

func HAEToMSL(lon, lat, altitude, ellipsoidOffset float64, g geoid.VerticalDatum) float64 {
	return getGeoid(g).ConvertHeight(lat, lon, altitude+ellipsoidOffset, geoid.ELLIPSOIDTOGEOID)
}

func MSLToHAE(h, lon, lat, ellipsoidOffset float64, g geoid.VerticalDatum) float64 {
	return getGeoid(g).ConvertHeight(lat, lon, h, geoid.GEOIDTOELLIPSOID) + ellipsoidOffset
}

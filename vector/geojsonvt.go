package vector

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/flywave/go-mbgeom/geojsonvt"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoJSONVT map[string]*geojsonvt.FeatureCollection

type GeoJSONVTSource struct {
	VectorSource
}

type GeoJSONVTOptions struct {
	tile.TileOptions
	Options geojsonvt.TileOptions
}

func (s *GeoJSONVTOptions) GetFormat() tile.TileFormat {
	return tile.TileFormat("application/json")
}

func NewGeoJSONVTSource(tile [3]int, options tile.TileOptions) *GeoJSONVTSource {
	src := &GeoJSONVTSource{VectorSource: VectorSource{tile: tile, Options: options}}
	src.decodeFunc = func(r io.Reader) (interface{}, error) {
		geojsonOpt := options.(*GeoJSONVTOptions)
		vt := LoadGeoJSONVT(r, tile, geojsonOpt.Options)
		return vt, nil
	}
	src.encodeFunc = func(data interface{}) ([]byte, error) {
		buf := &bytes.Buffer{}
		err := SaveGeoJSONVT(buf, data.(GeoJSONVT))
		return buf.Bytes(), err
	}

	return src
}

func LoadGeoJSONVT(r io.Reader, tile [3]int, opts geojsonvt.TileOptions) GeoJSONVT {
	jsondata, _ := ioutil.ReadAll(r)
	geojson := geojsonvt.ParseFeatureCollections(string(jsondata))
	return geojson
}

func SaveGeoJSONVT(w io.Writer, fc GeoJSONVT) error {
	json := geojsonvt.StringifyFeatureCollections(fc)
	_, err := w.Write([]byte(json))
	return err
}

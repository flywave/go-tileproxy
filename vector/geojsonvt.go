package vector

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-mbgeom/geojson"
	"github.com/flywave/go-mbgeom/geojsonvt"
	"github.com/flywave/go-tileproxy/tile"
)

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
		err := SaveGeoJSONVT(buf, data.(*geojsonvt.Tile))
		return buf.Bytes(), err
	}

	return src
}

func LoadGeoJSONVT(r io.Reader, tile [3]int, opts geojsonvt.TileOptions) *geojsonvt.Tile {
	jsondata, _ := ioutil.ReadAll(r)
	geomdata, _ := general.UnmarshalFeatureCollection(jsondata)
	geojson := geojson.NewGeoJSONFromGeomFeatureCollection(geomdata)
	return (*geojsonvt.Tile)(geojsonvt.NewGeoJSONVT(geojson, opts).GetTile(uint32(tile[2]), uint32(tile[1]), uint32(tile[0])))
}

func SaveGeoJSONVT(w io.Writer, vts *geojsonvt.Tile) error {
	fc := vts.GetFeatureCollection()
	json := fc.Stringify()
	_, err := w.Write([]byte(json))
	return err
}

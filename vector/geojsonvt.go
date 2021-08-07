package vector

import (
	"io"
	"io/ioutil"

	"github.com/flywave/go-geom/general"
	"github.com/flywave/go-mbgeom/geojson"
	"github.com/flywave/go-mbgeom/geojsonvt"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoJSONVT geojsonvt.Tile

type GeoJSONVTSource struct {
	VectorSource
}

type GeoJSONVTOptions struct {
	tile.TileOptions
	VTOptions geojsonvt.TileOptions
}

func NewGeoJSONVTSource(tile [3]int, options tile.TileOptions) *GeoJSONVTSource {
	src := &GeoJSONVTSource{VectorSource: VectorSource{tile: tile, Options: options}}
	src.decodeFunc = func(r io.Reader) (interface{}, error) {
		geojsonOpt := options.(*GeoJSONVTOptions)
		vt := LoadGeoJSONVT(r, tile, geojsonOpt.VTOptions)
		return vt, nil
	}
	src.encodeFunc = func(data interface{}) ([]byte, error) {
		return nil, nil
	}

	return src
}

func LoadGeoJSONVT(r io.Reader, tile [3]int, opts geojsonvt.TileOptions) *GeoJSONVT {
	jsondata, _ := ioutil.ReadAll(r)
	geomdata, _ := general.UnmarshalFeatureCollection(jsondata)
	geojson := geojson.NewGeoJSONFromGeomFeatureCollection(geomdata)
	return (*GeoJSONVT)(geojsonvt.NewGeoJSONVT(geojson, opts).GetTile(uint32(tile[2]), uint32(tile[1]), uint32(tile[0])))
}

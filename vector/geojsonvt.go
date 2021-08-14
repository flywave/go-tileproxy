package vector

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"

	"github.com/flywave/go-geom"
	"github.com/flywave/go-mbgeom/geojsonvt"
	"github.com/flywave/go-tileproxy/tile"
)

type GeoJSONVTSource struct {
	VectorSource
}

func NewGeoJSONVTSource(tile [3]int, options tile.TileOptions) *GeoJSONVTSource {
	src := &GeoJSONVTSource{VectorSource: VectorSource{tile: tile, Options: options}}
	geojsonOpt := options.(*VectorOptions)
	src.io = &GeoJSONVTIO{tile: tile, options: geojsonOpt}
	return src
}

func NewEmptyGeoJSONVTSource(options tile.TileOptions) *GeoJSONVTSource {
	src := &GeoJSONVTSource{VectorSource: VectorSource{tile: [3]int{0, 0, 0}, Options: options}}
	geojsonOpt := options.(*VectorOptions)
	src.io = &GeoJSONVTIO{tile: [3]int{0, 0, 0}, options: geojsonOpt}
	src.SetSource(Vector{})
	return src
}

type GeoJSONVTIO struct {
	VectorIO
	tile    [3]int
	options *VectorOptions
}

func (i *GeoJSONVTIO) Decode(r io.Reader) (interface{}, error) {
	vt := LoadGeoJSONVT(r, i.tile, i.options)
	return vt, nil
}

func (i *GeoJSONVTIO) Encode(data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := SaveGeoJSONVT(buf, i.tile, i.options, data.(Vector))
	return buf.Bytes(), err
}

func LoadGeoJSONVT(r io.Reader, tile [3]int, opts *VectorOptions) Vector {
	jsondata, _ := ioutil.ReadAll(r)
	geojson := geojsonvt.ParseFeatureCollections(string(jsondata))
	ret := make(Vector)

	for k, v := range geojson {
		ret[k] = make([]*geom.Feature, v.Count())
		for i := range ret[k] {
			ret[k][i], _ = ToGeoJSON(int(opts.Extent), tile, geojson[k].Get(i))
		}
	}
	return ret
}

func SaveGeoJSONVT(w io.Writer, tile [3]int, opts *VectorOptions, fc Vector) error {
	vtfc := make(map[string]*geojsonvt.FeatureCollection)
	for k, v := range fc {
		vtfc[k] = geojsonvt.NewFeatureCollection()
		for i := range v {
			fea, _ := ToGeoJSONVT(int(opts.Extent), tile, v[i])
			vtfc[k].Append(fea)
		}
	}

	json := geojsonvt.StringifyFeatureCollections(vtfc)
	_, err := w.Write([]byte(json))
	return err
}

func project(line [][]int16, x0 float64, y0 float64, size float64) [][]float64 {
	retline := make([][]float64, len(line))
	for j := range line {
		p := line[j]
		y2 := 180.0 - (float64(p[1])+y0)*360.0/size
		retline[j] = []float64{
			(float64(p[0])+x0)*360.0/size - 180.0,
			360.0/math.Pi*math.Atan(math.Exp(y2*math.Pi/180.0)) - 90.0}
	}
	return retline
}

const mercatorPole = 20037508.34

func convertPoint(point []float64) []float64 {
	x := mercatorPole * point[0] / 180.0

	y := math.Log(math.Tan((90.0+point[1])*math.Pi/360.0)) / math.Pi * mercatorPole
	y = math.Max(-mercatorPole, math.Min(y, mercatorPole))
	return []float64{x, y}
}

func tile_ul(tileid [3]int) []float64 {
	n := math.Pow(2.0, float64(tileid[2]))
	lon_deg := float64(tileid[0])/n*360.0 - 180.0
	lat_rad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(tileid[1])/n)))
	lat_deg := (180.0 / math.Pi) * lat_rad
	return []float64{lon_deg, lat_deg}
}

func tileBounds(tileid [3]int) []float64 {
	a := tile_ul(tileid)
	b := tile_ul([3]int{tileid[0] + 1, tileid[1] + 1, tileid[2]})
	return []float64{a[0], b[1], b[0], a[1]}
}

func unproject(line [][]float64, x0 float64, y0 float64, deltaX float64, deltaY float64, extent int) [][2]int16 {
	retline := make([][2]int16, len(line))
	for i := range line {
		point := line[i]

		factorx := (point[0] - x0) / deltaX
		factory := (y0 - point[1]) / deltaY

		xval := int16(math.Round(factorx * float64(extent)))
		yval := int16(math.Round(factory * float64(extent)))

		retline[i] = [2]int16{xval, yval}
	}
	return retline
}

func ToGeoJSONVT(extent int, tile [3]int, feat *geom.Feature) (*geojsonvt.Feature, error) {
	var err error
	defer func() {
		if recover() != nil {
			err = errors.New("Error in feature.ToGeoJSON()")
		}
	}()
	bound := tileBounds(tile)
	deltax := bound[2] - bound[0]
	deltay := bound[3] - bound[1]

	x0 := bound[0]
	y0 := bound[3]

	geometry := feat.GeometryData

	if geometry.Type == "" && feat.Geometry != nil {
		geometry = *geom.NewGeometryData(feat.Geometry)
	}

	var newGeometry *geojsonvt.Geometry

	switch geometry.GetType() {
	case "Point":
		point := unproject([][]float64{geometry.Point}, x0, y0, deltax, deltay, extent)[0]
		pt := geojsonvt.NewPoint(float64(point[0]), float64(point[1]))
		newGeometry = pt.Geom()
	case "MultiPoint":
		multiPoint := unproject(geometry.MultiPoint, x0, y0, deltax, deltay, extent)
		pt := geojsonvt.NewMultiPoint(multiPoint)
		newGeometry = pt.Geom()
	case "LineString":
		lineString := unproject(geometry.LineString, x0, y0, deltax, deltay, extent)
		pt := geojsonvt.NewLineString(lineString)
		newGeometry = pt.Geom()
	case "MultiLineString":
		lineStrings := make([]*geojsonvt.LineString, len(geometry.MultiLineString))
		for i, ls := range geometry.MultiLineString {
			lineStrings[i] = geojsonvt.NewLineString(unproject(ls, x0, y0, deltax, deltay, extent))
		}
		pt := geojsonvt.NewMultiLineString(lineStrings)
		newGeometry = pt.Geom()
	case "Polygon":
		linearRings := make([]*geojsonvt.LinearRing, len(geometry.Polygon))
		for i, lr := range geometry.Polygon {
			linearRings[i] = geojsonvt.NewLinearRing(unproject(lr, x0, y0, deltax, deltay, extent))
		}
		pt := geojsonvt.NewPolygon(linearRings)
		newGeometry = pt.Geom()
	case "MultiPolygon":
		polygons := make([]*geojsonvt.Polygon, len(geometry.MultiPolygon))
		for i, polys := range geometry.MultiPolygon {
			linearRings := make([]*geojsonvt.LinearRing, len(polys))
			for j, ls := range polys {
				linearRings[j] = geojsonvt.NewLinearRing(unproject(ls, x0, y0, deltax, deltay, extent))
			}
			polygons[i] = geojsonvt.NewPolygon(linearRings)
		}
		pt := geojsonvt.NewMultiPolygon(polygons)
		newGeometry = pt.Geom()
	}

	newFeature := geojsonvt.NewFeature(newGeometry)

	pmap := geojsonvt.NewPropertyMapRaw(feat.Properties)
	newFeature.SetPropertyMap(pmap)

	identifier := geojsonvt.NewIdentifier(feat.ID)
	if identifier != nil {
		newFeature.SetIdentifier(identifier)
	}
	return newFeature, err
}

func ToGeoJSON(extent int, tile [3]int, feat *geojsonvt.Feature) (*geom.Feature, error) {
	var err error
	defer func() {
		if recover() != nil {
			err = errors.New("Error in feature.ToGeoJSON()")
		}
	}()
	size := float64(extent) * float64(math.Pow(2, float64(tile[2])))
	x0 := float64(extent) * float64(tile[0])
	y0 := float64(extent) * float64(tile[1])
	geometry := feat.GetGeometry()
	if err != nil {
		return &geom.Feature{}, err
	}

	geomd := &geom.GeometryData{Type: geom.GeometryType(geometry.GetType())}
	switch geometry.GetType() {
	case "Point":
		pt := geometry.Cast().(*geojsonvt.Point)
		geomd.Point = project([][]int16{pt.Data()}, x0, y0, size)[0]
	case "MultiPoint":
		pt := geometry.Cast().(*geojsonvt.MultiPoint)
		geomd.MultiPoint = project(pt.Data(), x0, y0, size)
	case "LineString":
		pt := geometry.Cast().(*geojsonvt.LineString)
		geomd.LineString = project(pt.Data(), x0, y0, size)
	case "MultiLineString":
		pt := geometry.Cast().(*geojsonvt.MultiLineString)
		geomd.MultiLineString = make([][][]float64, pt.Count())
		for i, ls := range pt.Data() {
			geomd.MultiLineString[i] = project(ls, x0, y0, size)
		}
	case "Polygon":
		pt := geometry.Cast().(*geojsonvt.Polygon)
		geomd.Polygon = make([][][]float64, len(pt.Data()))
		for i, lr := range pt.Data() {
			geomd.Polygon[i] = project(lr, x0, y0, size)
		}
	case "MultiPolygon":
		pt := geometry.Cast().(*geojsonvt.MultiPolygon)
		geomd.MultiPolygon = make([][][][]float64, pt.Count())
		for i, polys := range pt.Data() {
			geomd.MultiPolygon[i] = make([][][]float64, len(polys))
			for j, ls := range polys {
				geomd.MultiPolygon[i][j] = project(ls, x0, y0, size)
			}
		}
	}

	newFeature := geom.NewFeatureFromGeometryData(geomd)
	newFeature.Properties = feat.GetPropertyMap().RawMap()
	newFeature.ID = feat.GetIdentifier().Get()

	return newFeature, err
}

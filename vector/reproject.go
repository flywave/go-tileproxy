package vector

import (
	"github.com/flywave/go-geom"
	"github.com/flywave/go-tileproxy/geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type VectorTransformer struct {
	SrcSRS geo.Proj
	DstSRS geo.Proj
}

func NewVectorTransformer(src, dst geo.Proj) *VectorTransformer {
	return &VectorTransformer{SrcSRS: src, DstSRS: dst}
}

func (t *VectorTransformer) project(line [][]float64) [][]float64 {
	retline := make([][]float64, len(line))
	for i, p := range line {
		pv := t.SrcSRS.TransformTo(t.DstSRS, []vec2d.T{{p[0], p[1]}})
		retline[i] = pv[0][:]
	}
	return retline
}

func (t *VectorTransformer) apply(feat *geom.Feature) *geom.Feature {
	geometry := feat.GeometryData

	if geometry.Type == "" && feat.Geometry != nil {
		geometry = *geom.NewGeometryData(feat.Geometry)
	}

	switch geometry.GetType() {
	case "Point":
		geometry.Point = t.project([][]float64{geometry.Point})[0]
	case "MultiPoint":
		geometry.MultiPoint = t.project(geometry.MultiPoint)
	case "LineString":
		geometry.LineString = t.project(geometry.LineString)
	case "MultiLineString":
		for i, ls := range geometry.MultiLineString {
			geometry.MultiLineString[i] = t.project(ls)
		}
	case "Polygon":
		for i, lr := range geometry.Polygon {
			geometry.Polygon[i] = t.project(lr)
		}
	case "MultiPolygon":
		for i, polys := range geometry.MultiPolygon {
			for j, ls := range polys {
				geometry.MultiPolygon[i][j] = t.project(ls)
			}
		}
	}

	newFeature := geom.NewFeatureFromGeometryData(&geometry)
	newFeature.Properties = feat.Properties
	newFeature.ID = feat.ID

	return newFeature
}

func (t *VectorTransformer) Apply(feats []*geom.Feature) []*geom.Feature {
	rets := make([]*geom.Feature, len(feats))
	for i, f := range feats {
		rets[i] = t.apply(f)
	}
	return rets
}

func (t *VectorTransformer) ApplyVector(layers Vector) Vector {
	for k, l := range layers {
		layers[k] = t.Apply(l)
	}
	return layers
}

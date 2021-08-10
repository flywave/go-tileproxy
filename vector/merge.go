package vector

import (
	"github.com/flywave/go-geom"
	"github.com/flywave/go-mbgeom/vtile"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type PointHandler struct{}

func (h *PointHandler) pointsBegin(uint32) {}

func (h *PointHandler) pointsPoint(pt []float64) {}

func (h *PointHandler) pointsEnd() {}

type LineStringHandler struct{}

func (ls *LineStringHandler) linestringBegin(count uint32) {}

func (ls *LineStringHandler) linestringPoint(pt []float64) {}

func (ls *LineStringHandler) linestringEnd() {}

type PolygonHandler struct{}

func (h *PolygonHandler) ringBegin(uint32) {}

func (h *PolygonHandler) ringPoint(pt []float64) {}

func (h *PolygonHandler) ringEnd() {}

type FeatureBuilder struct {
	builder *LayerBuilder
}

func (l *FeatureBuilder) applyGeometryPoint(feature *geom.Feature) {

}

func (l *FeatureBuilder) applyGeometryLinestring(feature *geom.Feature) {

}

func (l *FeatureBuilder) applyGeometryPolygon(feature *geom.Feature) {

}

func (l *FeatureBuilder) Apply(feature *geom.Feature) {
	switch feature.GeometryData.Type {
	case "Point":
		l.applyGeometryPoint(feature)
		break
	case "LineString":
		l.applyGeometryLinestring(feature)
		break
	case "Polygon":
		l.applyGeometryPolygon(feature)
		break
	default:
		break
	}
}

type LayerBuilder struct {
	dst      [3]int
	coverage geo.Coverage
	result   []*geom.Feature
	builder  *TileBuilder
}

func (l *LayerBuilder) AddFeatures(feats []*geom.Feature) {
	for _, f := range feats {
		l.addFeature(f)
	}
}

func (l *LayerBuilder) addFeature(feature *geom.Feature) {

}

type TileBuilder struct {
	dst      [3]int
	coverage geo.Coverage
	layers   map[string]*LayerBuilder
}

func NewTileBuilder(dst [3]int, coverage geo.Coverage) *TileBuilder {
	return &TileBuilder{dst: dst, coverage: coverage, layers: make(map[string]*LayerBuilder)}
}

func (b *TileBuilder) AddLayers(layers map[string][]*geom.Feature) {
	for l, f := range layers {
		b.addLayer(l, f)
	}
}

func (b *TileBuilder) addLayer(layer string, feats []*geom.Feature) {
	if _, ok := b.layers[layer]; !ok {
		b.layers[layer] = &LayerBuilder{dst: b.dst, coverage: b.coverage, result: []*geom.Feature{}}
	}
	b.layers[layer].AddFeatures(feats)
}

type VectorMerger struct {
	tile.Merger
	Layers    []tile.Source
	Coverages []geo.Coverage
	Cacheable *tile.CacheInfo
}

func (l *VectorMerger) getTiles() []*vtile.TileObject {
	return nil
}

func (l *VectorMerger) AddSource(src tile.Source, cov geo.Coverage) {

}

func (l *VectorMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	return nil
}

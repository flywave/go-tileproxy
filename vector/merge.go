package vector

import (
	"github.com/flywave/go-geom"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type FeatureBuilder struct {
	feat     *geom.Feature
	coverage geo.Coverage
	srs      geo.Proj
}

func NewFeatureBuilder(cov geo.Coverage, srs geo.Proj) *FeatureBuilder {
	return &FeatureBuilder{feat: nil, coverage: cov, srs: srs}
}

func (l *FeatureBuilder) applyGeometryPoint(feature *geom.Feature) {
	pt := feature.GeometryData.Point
	if l.coverage.ContainsPoint(pt, l.srs) {
		l.feat = geom.NewFeatureFromGeometryData(&feature.GeometryData)
		l.feat.ID = feature.ID
		l.feat.Properties = feature.Properties
	}
}

func (l *FeatureBuilder) applyGeometryLinestring(feature *geom.Feature) {
	if feature.BoundingBox == nil {
		feature.BoundingBox = geom.BoundingBoxFromGeometryData(&feature.GeometryData)
	}
	rect := vec2d.Rect{Min: vec2d.T{feature.BoundingBox[0], feature.BoundingBox[1]}, Max: vec2d.T{feature.BoundingBox[2], feature.BoundingBox[3]}}
	if l.coverage.Intersects(rect, l.srs) {
		l.feat = geom.NewFeatureFromGeometryData(&feature.GeometryData)
		l.feat.ID = feature.ID
		l.feat.Properties = feature.Properties
	}
}

func (l *FeatureBuilder) applyGeometryPolygon(feature *geom.Feature) {
	if feature.BoundingBox == nil {
		feature.BoundingBox = geom.BoundingBoxFromGeometryData(&feature.GeometryData)
	}
	rect := vec2d.Rect{Min: vec2d.T{feature.BoundingBox[0], feature.BoundingBox[1]}, Max: vec2d.T{feature.BoundingBox[2], feature.BoundingBox[3]}}
	if l.coverage.Intersects(rect, l.srs) {
		l.feat = geom.NewFeatureFromGeometryData(&feature.GeometryData)
		l.feat.ID = feature.ID
		l.feat.Properties = feature.Properties
	}
}

func (l *FeatureBuilder) Apply(feature *geom.Feature) {
	if feature.GeometryData.Type == "" && feature.Geometry == nil {
		feature.GeometryData = *geom.NewGeometryData(feature.Geometry)
	}
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

func (l *FeatureBuilder) Finalize() *geom.Feature {
	return l.feat
}

type LayerBuilder struct {
	coverage geo.Coverage
	result   []*geom.Feature
	builder  *TileBuilder
	srs      geo.Proj
}

func (l *LayerBuilder) AddFeatures(feats []*geom.Feature) {
	for _, f := range feats {
		l.addFeature(f)
	}
}

func (l *LayerBuilder) addFeature(feature *geom.Feature) {
	builder := NewFeatureBuilder(l.coverage, l.srs)
	builder.Apply(feature)
	l.result = append(l.result, builder.Finalize())
}

func (l *LayerBuilder) Finalize() []*geom.Feature {
	return l.result
}

type TileBuilder struct {
	coverage geo.Coverage
	layers   map[string]*LayerBuilder
}

func NewTileBuilder(coverage geo.Coverage) *TileBuilder {
	return &TileBuilder{coverage: coverage, layers: make(map[string]*LayerBuilder)}
}

func (b *TileBuilder) AddLayers(layers map[string][]*geom.Feature) {
	for l, f := range layers {
		b.addLayer(l, f)
	}
}

func (b *TileBuilder) addLayer(layer string, feats []*geom.Feature) {
	if _, ok := b.layers[layer]; !ok {
		b.layers[layer] = &LayerBuilder{coverage: b.coverage, result: []*geom.Feature{}}
	}
	b.layers[layer].AddFeatures(feats)
}

func (b *TileBuilder) Finalize() map[string][]*geom.Feature {
	ret := make(map[string][]*geom.Feature)
	for l, f := range b.layers {
		ret[l] = f.Finalize()
	}
	return ret
}

type VectorMerger struct {
	tile.Merger
	Layers []tile.Source
}

func NewVectorMerger(tiles []tile.Source) *VectorMerger {
	return &VectorMerger{Layers: tiles}
}

func (l *VectorMerger) AddSource(src tile.Source, cov geo.Coverage) {
	l.Layers = append(l.Layers, src)
}

func (l *VectorMerger) Merge(opts tile.TileOptions, size []uint32, tile [3]int, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) Vector {
	if len(l.Layers) == 1 {
		t := l.Layers[0].GetTile()
		feats := t.(map[string][]*geom.Feature)
		return feats
	}

	if size == nil {
		ss := l.Layers[0].GetSize()
		size = ss[:]
	}

	if coverage == nil {
		coverage = geo.NewBBoxCoverage(bbox, bbox_srs, false)
	}

	builder := NewTileBuilder(coverage)
	for i := range l.Layers {
		layer_vec := l.Layers[i]

		t := layer_vec.GetTile()
		if t == nil {
			return nil
		}
		feats, ok := t.(map[string][]*geom.Feature)
		if !ok {
			return nil
		}
		builder.AddLayers(feats)
	}
	feats := builder.Finalize()
	return feats
}

type TiledVector struct {
	Tiles    []tile.Source
	TileGrid [2]int
	TileSize [2]uint32
	SrcBBox  vec2d.Rect
	SrcSRS   geo.Proj
}

func NewTiledVector(tiles []tile.Source, tile_grid [2]int, tile_size [2]uint32, src_bbox vec2d.Rect, src_srs geo.Proj) *TiledVector {
	return &TiledVector{Tiles: tiles, TileGrid: tile_grid, TileSize: tile_size, SrcBBox: src_bbox, SrcSRS: src_srs}
}

func (t *TiledVector) GetVector(v_opts *VectorOptions, tile [3]int) Vector {
	tm := NewVectorMerger(t.Tiles)
	return tm.Merge(v_opts, t.TileSize[:], tile, t.SrcBBox, t.SrcSRS, nil)
}

func (t *TiledVector) Transform(req_bbox vec2d.Rect, grid geo.TileGrid, vec_opts *VectorOptions) tile.Source {
	_, _, tiles, _ := grid.GetAffectedTiles(req_bbox, t.TileSize, grid.Srs)

	x, y, z, _ := tiles.Next()

	src_img := t.GetVector(vec_opts, [3]int{x, y, z})

	transformer := NewVectorTransformer(t.SrcSRS, grid.Srs)

	vecs := transformer.ApplyVector(src_img)

	return CreateVectorSourceFromVector(vecs, [3]int{x, y, z}, vec_opts, nil)
}

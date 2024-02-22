package vector

import (
	"github.com/flywave/go-geom"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
)

type FeatureBuilder struct {
	feat     *geom.Feature
	coverage geo.Coverage
	srs      geo.Proj
}

func NewFeatureBuilder(cov geo.Coverage) *FeatureBuilder {
	return &FeatureBuilder{feat: nil, coverage: cov, srs: cov.GetSrs()}
}

func (l *FeatureBuilder) applyGeometryPoint(feature *geom.Feature) {
	pt := feature.GeometryData.Point
	if l.coverage == nil || l.coverage.ContainsPoint(pt, l.srs) {
		l.feat = geom.NewFeatureFromGeometryData(&feature.GeometryData)
		l.feat.ID = feature.ID
		l.feat.Properties = feature.Properties
	}
}

func (l *FeatureBuilder) applyGeometryLinestring(feature *geom.Feature) {
	if l.coverage != nil && feature.BoundingBox == nil {
		feature.BoundingBox = geom.BoundingBoxFromGeometryData(&feature.GeometryData)
	}
	rect := vec2d.Rect{Min: vec2d.T{feature.BoundingBox[0][0], feature.BoundingBox[0][1]}, Max: vec2d.T{feature.BoundingBox[1][0], feature.BoundingBox[1][1]}}
	if l.coverage == nil || l.coverage.Intersects(rect, l.srs) {
		l.feat = geom.NewFeatureFromGeometryData(&feature.GeometryData)
		l.feat.ID = feature.ID
		l.feat.Properties = feature.Properties
	}
}

func (l *FeatureBuilder) applyGeometryPolygon(feature *geom.Feature) {
	if l.coverage != nil && feature.BoundingBox == nil {
		feature.BoundingBox = geom.BoundingBoxFromGeometryData(&feature.GeometryData)
	}
	rect := vec2d.Rect{Min: vec2d.T{feature.BoundingBox[0][0], feature.BoundingBox[0][1]}, Max: vec2d.T{feature.BoundingBox[1][0], feature.BoundingBox[1][1]}}
	if l.coverage == nil || l.coverage.Intersects(rect, l.srs) {
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
	case "LineString":
		l.applyGeometryLinestring(feature)
	case "Polygon":
		l.applyGeometryPolygon(feature)
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
}

func (l *LayerBuilder) AddFeatures(feats []*geom.Feature) {
	for _, f := range feats {
		l.addFeature(f)
	}
}

func (l *LayerBuilder) addFeature(feature *geom.Feature) {
	builder := NewFeatureBuilder(l.coverage)
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

func (l *VectorMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) Vector {
	if len(l.Layers) == 1 {
		t := l.Layers[0].GetTile()
		feats := t.(Vector)
		return feats
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
		feats, ok := t.(Vector)
		if !ok {
			return nil
		}
		builder.AddLayers(feats)
	}
	feats := builder.Finalize()
	return feats
}

type TiledVector struct {
	Tiles        []tile.Source
	TileGrid     [2]int
	TileSize     [2]uint32
	SrcBBox      vec2d.Rect
	SrcSRS       geo.Proj
	BufferFactor float64
}

func NewTiledVector(tiles []tile.Source, tile_grid [2]int, tile_size [2]uint32, src_bbox vec2d.Rect, src_srs geo.Proj) *TiledVector {
	return &TiledVector{Tiles: tiles, TileGrid: tile_grid, TileSize: tile_size, SrcBBox: src_bbox, SrcSRS: src_srs, BufferFactor: 0.2}
}

func (t *TiledVector) GetVector(v_opts *VectorOptions, dest_bbox vec2d.Rect, dest_srs geo.Proj) Vector {
	var tranbbox vec2d.Rect
	if !t.SrcSRS.Eq(dest_srs) {
		tranbbox = dest_srs.TransformRectTo(t.SrcSRS, dest_bbox, 16)
	}

	if t.BufferFactor != 0 {
		xbuf, ybuf := tranbbox.Width()*t.BufferFactor, tranbbox.Height()*t.BufferFactor

		tranbbox.Min[0] = tranbbox.Min[0] - xbuf
		tranbbox.Min[1] = tranbbox.Min[1] - ybuf

		tranbbox.Max[0] = tranbbox.Max[0] + xbuf
		tranbbox.Max[1] = tranbbox.Max[1] + ybuf
	}

	geo_srs := geo.GetLatLongProj(t.SrcSRS)
	if !t.SrcSRS.IsLatLong() {
		tranbbox = t.SrcSRS.TransformRectTo(geo_srs, dest_bbox, 16)
	}

	tm := NewVectorMerger(t.Tiles)
	return tm.Merge(v_opts, t.TileSize[:], tranbbox, geo_srs, nil)
}

func (t *TiledVector) Transform(grid *geo.TileGrid, req_bbox vec2d.Rect, req_srs geo.Proj, out_size [2]uint32, vec_opts *VectorOptions) tile.Source {
	src_img := t.GetVector(vec_opts, req_bbox, req_srs)

	geo_src_srs := geo.GetLatLongProj(t.SrcSRS)
	geo_req_srs := geo.GetLatLongProj(req_srs)

	transformer := NewVectorTransformer(geo_src_srs, geo_req_srs)

	vecs := transformer.ApplyVector(src_img)

	_, _, tiles, _ := grid.GetAffectedTiles(req_bbox, t.TileSize, req_srs)

	x, y, z, _ := tiles.Next()

	return CreateVectorSourceFromVector(vecs, [3]int{x, y, z}, vec_opts, nil)
}

func Resample(tiles []tile.Source, tile_grid [2]int, tile_size [2]uint32, grid *geo.TileGrid, src_bbox vec2d.Rect, src_srs geo.Proj, req_bbox vec2d.Rect, req_srs geo.Proj, out_size [2]uint32, src_opts, dest_opts *VectorOptions) tile.Source {
	rr := NewTiledVector(tiles, tile_grid, tile_size, src_bbox, src_srs)
	return rr.Transform(grid, req_bbox, req_srs, out_size, dest_opts)
}

package layer

import (
	"errors"
	"math"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type InfoLayer interface {
	GetInfo(query *InfoQuery) resource.FeatureInfoDoc
}

type LegendLayer interface {
	GetLegend(query *LegendQuery) tile.Source
}

type CesiumLayerJSONLayer interface {
	GetLayerJSON(id string) *resource.LayerJson
}

type MapboxTileJSONLayer interface {
	GetTileJSON(id string) *resource.TileJSON
}

type Layer interface {
	GetMap(query *MapQuery) (tile.Source, error)
	GetResolutionRange() *geo.ResolutionRange
	IsSupportMetaTiles() bool
	GetExtent() *geo.MapExtent
	CombinedLayer(other Layer, query *MapQuery) Layer
	GetCoverage() geo.Coverage
	GetOptions() tile.TileOptions
}

type MapLayer struct {
	Layer
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
	Options          tile.TileOptions
}

func NewMapLayer(opts tile.TileOptions) *MapLayer {
	return &MapLayer{Options: opts}
}

func (l *MapLayer) GetResolutionRange() *geo.ResolutionRange {
	return l.ResRange
}

func (l *MapLayer) IsSupportMetaTiles() bool {
	return l.SupportMetaTiles
}

func (l *MapLayer) GetCoverage() geo.Coverage {
	return l.Coverage
}

func (l *MapLayer) CheckResRange(query *MapQuery) error {
	if l.ResRange != nil &&
		!l.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return errors.New("res range not set")
	}
	return nil
}

func (l *MapLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	return nil
}

func (l *MapLayer) GetExtent() *geo.MapExtent {
	return l.Extent
}

func (l *MapLayer) GetOptions() tile.TileOptions {
	return l.Options
}

type LimitedLayer struct {
	layer    Layer
	coverage geo.Coverage
}

func (l *LimitedLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	var combined Layer
	if l.coverage.Equals(other.GetCoverage()) {
		combined = l.layer.CombinedLayer(other, query)
	}
	if combined != nil {
		return &LimitedLayer{layer: combined, coverage: l.coverage}
	}
	return nil
}

func (l *LimitedLayer) IsSupportMetaTiles() bool {
	return l.layer.IsSupportMetaTiles()
}

func (l *LimitedLayer) GetResolutionRange() *geo.ResolutionRange {
	return l.layer.GetResolutionRange()
}

func (l *LimitedLayer) GetCoverage() geo.Coverage {
	return l.layer.GetCoverage()
}

func (l *LimitedLayer) GetMap(query *MapQuery) (tile.Source, error) {
	return l.layer.GetMap(query)
}

func (l *LimitedLayer) GetExtent() *geo.MapExtent {
	return l.layer.GetExtent()
}

func (l *LimitedLayer) GetOptions() tile.TileOptions {
	return l.layer.GetOptions()
}

type ResolutionConditional struct {
	MapLayer
	a          Layer
	b          Layer
	resolution float64
	srs        geo.Proj
}

func MergeLayerResRanges(layers []Layer) *geo.ResolutionRange {
	ranges := []*geo.ResolutionRange{}
	for _, l := range layers {
		ranges = append(ranges, l.GetResolutionRange())
	}
	var ret *geo.ResolutionRange
	if len(ranges) > 0 {
		ret = ranges[0]
		for _, r := range ranges[1:] {
			ret = geo.MergeResolutionRange(ret, r)
		}
	}
	return ret
}

func NewResolutionConditional(a, b Layer, resolution float64, srs geo.Proj, ext *geo.MapExtent) *ResolutionConditional {
	res_range := MergeLayerResRanges([]Layer{a, b})
	ret := &ResolutionConditional{MapLayer: MapLayer{SupportMetaTiles: true, ResRange: res_range, Extent: ext}, a: a, b: b, resolution: resolution, srs: srs}
	return ret
}

func (r *ResolutionConditional) GetMap(query *MapQuery) (tile.Source, error) {
	if err := r.CheckResRange(query); err != nil {
		return nil, errors.New("res error")
	}
	bbox := query.BBox
	if !query.Srs.Eq(r.srs) {
		bbox = query.Srs.TransformRectTo(r.srs, bbox, 16)
	}
	xres := (bbox.Max[0] - bbox.Min[0]) / float64(query.Size[0])
	yres := (bbox.Max[1] - bbox.Min[1]) / float64(query.Size[1])
	res := math.Min(xres, yres)

	if res > r.resolution {
		return r.a.GetMap(query)
	} else {
		return r.b.GetMap(query)
	}
}

type SRSConditional struct {
	MapLayer
	supportedSRS *geo.SupportedSRS
	srsMap       map[string]Layer
}

func NewSRSConditional(lmap map[string]Layer, ext *geo.MapExtent, preferred_srs geo.PreferredSrcSRS) *SRSConditional {
	ret := &SRSConditional{MapLayer: MapLayer{SupportMetaTiles: true}}
	layers := []Layer{}
	for _, layer := range lmap {
		layers = append(layers, layer)
	}
	res_range := MergeLayerResRanges(layers)
	ret.srsMap = make(map[string]Layer)

	supported_srs := []geo.Proj{}
	for srs, layer := range lmap {
		supported_srs = append(supported_srs, geo.NewProj(srs))
		ret.srsMap[srs] = layer
	}
	ret.supportedSRS = &geo.SupportedSRS{Srs: supported_srs, Preferred: preferred_srs}
	ret.Extent = ext
	ret.ResRange = res_range
	return ret
}

func (r *SRSConditional) GetMap(query *MapQuery) (tile.Source, error) {
	if err := r.CheckResRange(query); err != nil {
		return nil, errors.New("res error")
	}
	layer := r.selectLayer(query.Srs)
	return layer.GetMap(query)
}

func (r *SRSConditional) selectLayer(query_srs geo.Proj) Layer {
	srs, _ := r.supportedSRS.BestSrs(query_srs)
	return r.srsMap[srs.GetSrsCode()]
}

type DirectMapLayer struct {
	MapLayer
	source Layer
}

func NewDirectMapLayer(src Layer, ext *geo.MapExtent) *DirectMapLayer {
	return &DirectMapLayer{MapLayer: MapLayer{SupportMetaTiles: true, Extent: ext}, source: src}
}

func (r *DirectMapLayer) GetMap(query *MapQuery) (tile.Source, error) {
	if err := r.CheckResRange(query); err != nil {
		return nil, errors.New("res error")
	}
	return r.source.GetMap(query)
}

func MergeLayerExtents(layers []Layer) *geo.MapExtent {
	if len(layers) == 0 {
		return geo.MapExtentFromDefault()
	}
	extent := layers[0].GetExtent()
	layers = layers[1:]
	for _, layer := range layers[1:] {
		extent = extent.Add(layer.GetExtent())
	}
	return extent
}

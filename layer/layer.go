package layer

import (
	"errors"
	"math"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/resource"
)

type Layer interface {
	GetMap(query *MapQuery) images.Source
	GetResolutionRange() *geo.ResolutionRange
	IsSupportMetaTiles() bool
}

type MapLayer struct {
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
	Options          *images.ImageOptions
}

func NewMapLayer(opts *images.ImageOptions) *MapLayer {
	return &MapLayer{Options: opts}
}

func (l *MapLayer) IsSupportMetaTiles() bool {
	return l.SupportMetaTiles
}

func (l *MapLayer) GetOpacity() float64 {
	return *l.Options.Opacity
}

func (l *MapLayer) SetOpacity(value float64) {
	l.Options.Opacity = geo.NewFloat64(value)
}

func (l *MapLayer) IsOpaque(query *MapQuery) bool {
	return false
}

func (l *MapLayer) CheckResRange(query *MapQuery) error {
	if l.ResRange != nil &&
		!l.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return errors.New("BlankImage")
	}
	return nil
}

func (l *MapLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	return nil
}

type LimitedLayer struct {
	layer    Layer
	coverage geo.Coverage
}

func (l *LimitedLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	return nil
}

func (l *LimitedLayer) GetInfo(query *InfoQuery) *resource.FeatureInfo {
	return nil
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

func NewResolutionConditional(a, b Layer, resolution float64, srs geo.Proj, ext *geo.MapExtent, opacity *float64) *ResolutionConditional {
	res_range := MergeLayerResRanges([]Layer{a, b})
	ret := &ResolutionConditional{MapLayer: MapLayer{SupportMetaTiles: true, ResRange: res_range, Extent: ext}, a: a, b: b, resolution: resolution, srs: srs}
	if opacity != nil {
		ret.SetOpacity(*opacity)
	}
	return ret
}

func (r *ResolutionConditional) GetMap(query *MapQuery) images.Source {
	if err := r.CheckResRange(query); err != nil {
		return nil
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
	layers       []Layer
	supportedSRS *geo.SupportedSRS
	srsMap       map[string]Layer
}

func NewSRSConditional(lmap map[string]Layer, ext *geo.MapExtent, opacity *float64, preferred_srs geo.PreferredSrcSRS) *SRSConditional {
	ret := &SRSConditional{MapLayer: MapLayer{SupportMetaTiles: true}}
	layers := []Layer{}
	for _, layer := range lmap {
		layers = append(layers, layer)
	}
	res_range := MergeLayerResRanges(layers)
	ret.srsMap = make(map[string]Layer)

	supported_srs := []geo.Proj{}
	for srs, layer := range lmap {
		supported_srs = append(supported_srs, geo.NewSRSProj4(srs))
		ret.srsMap[srs] = layer
	}
	ret.supportedSRS = &geo.SupportedSRS{Srs: supported_srs, Preferred: preferred_srs}
	ret.Extent = ext
	if opacity != nil {
		ret.SetOpacity(*opacity)
	}
	ret.ResRange = res_range
	return ret
}

func (r *SRSConditional) GetMap(query *MapQuery) images.Source {
	if err := r.CheckResRange(query); err != nil {
		return nil
	}
	layer := r.selectLayer(query.Srs)
	return layer.GetMap(query)
}

func (r *SRSConditional) selectLayer(query_srs geo.Proj) Layer {
	srs, _ := r.supportedSRS.BestSrs(query_srs)
	return r.srsMap[srs.GetDef()]
}

type DirectMapLayer struct {
	MapLayer
	source Layer
}

func NewDirectMapLayer(src Layer, ext *geo.MapExtent) *DirectMapLayer {
	return &DirectMapLayer{MapLayer: MapLayer{SupportMetaTiles: true, Extent: ext}, source: src}
}

func (r *DirectMapLayer) GetMap(query *MapQuery) images.Source {
	if err := r.CheckResRange(query); err != nil {
		return nil
	}
	return r.source.GetMap(query)
}

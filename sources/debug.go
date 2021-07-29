package sources

import (
	"fmt"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type DebugSource struct {
	ImagerySource
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
}

func NewDebugSource() *DebugSource {
	return &DebugSource{Extent: geo.MapExtentFromDefault(), SupportMetaTiles: true}
}

func (s *DebugSource) GetMap(query *layer.MapQuery) tile.Source {
	bbox := query.BBox
	w := bbox.Max[0] - bbox.Min[0]
	h := bbox.Max[1] - bbox.Min[1]
	res_x := w / float64(query.Size[0])
	res_y := h / float64(query.Size[1])
	debug_info := fmt.Sprintf("bbox: {Min: {%.8f, %.8f}, Max: {%.8f, %.8f}} \nres: %.8f(%.8f)", bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1], res_x, res_y)
	return images.GenMessageImage(debug_info, query.Size, &images.ImageOptions{Transparent: geo.NewBool(true)}, nil, true)
}

type DummySource struct {
	ImagerySource
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
}

func NewDummySource(cov geo.Coverage) *DummySource {
	ds := &DummySource{Coverage: cov, SupportMetaTiles: true}
	ds.Extent = &geo.MapExtent{
		BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
		Srs:  geo.NewSRSProj4("EPSG:4326"),
	}
	if cov != nil {
		ds.Extent = &geo.MapExtent{BBox: cov.GetBBox(), Srs: cov.GetSrs()}
	}
	return ds
}

func (s *DummySource) GetMap(query *layer.MapQuery) tile.Source {
	return images.NewBlankImageSource(query.Size, &images.ImageOptions{Transparent: geo.NewBool(true)}, nil)
}

package sources

import (
	"bytes"
	"image"
	"image/color"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type WMSSource struct {
	layer.MapLayer
	Client                    client.MapClient
	SupportedSRS              *geo.SupportedSRS
	SupportedFormats          []string
	ExtReqParams              map[string]string
	TransparentColor          color.Color
	TransparentColorTolerance *float64
	Opacity                   *float64
}

func NewWMSSource(client client.MapClient, image_opts *imagery.ImageOptions, coverage geo.Coverage, res_range *geo.ResolutionRange,
	transparent_color color.Color, transparent_color_tolerance *float64,
	supported_srs *geo.SupportedSRS, supported_formats []string, fwd_req_params map[string]string) *WMSSource {
	src := &WMSSource{Client: client, MapLayer: layer.MapLayer{Options: image_opts, Coverage: coverage, ResRange: res_range, SupportMetaTiles: true}, TransparentColor: transparent_color, TransparentColorTolerance: transparent_color_tolerance, SupportedSRS: supported_srs, SupportedFormats: supported_formats, ExtReqParams: fwd_req_params}
	if transparent_color != nil {
		src.Options.Transparent = geo.NewBool(true)
	}
	if coverage != nil {
		src.Extent = &geo.MapExtent{BBox: coverage.GetBBox(), Srs: coverage.GetSrs()}
	} else {
		src.Extent = geo.MapExtentFromDefault()
	}
	return src
}

func (s *WMSSource) GetClient() client.MapClient {
	return s.Client
}

func (s *WMSSource) IsOpaque(query *layer.MapQuery) bool {
	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}

	if s.Options.Transparent != nil && *s.Options.Transparent {
		return false
	}

	if s.Opacity != nil && ((*s.Opacity) > 0.0 && (*s.Opacity) < 0.99) {
		return false
	}

	if s.Coverage == nil {
		return true
	}

	if s.Coverage.Contains(query.BBox, query.Srs) {
		return true
	}
	return false
}

func (s *WMSSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return imagery.NewBlankImageSource(query.Size, s.Options, nil), nil
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return imagery.NewBlankImageSource(query.Size, s.Options, nil), nil
	}
	resp := s.getMap(query)
	opts := resp.GetTileOptions().(*imagery.ImageOptions)
	opts.Opacity = s.Opacity
	return resp, nil
}

func (s *WMSSource) getMap(query *layer.MapQuery) tile.Source {
	format := s.Options.Format
	if format == "" {
		format = tile.TileFormat(query.Format)
	}
	if s.SupportedFormats != nil && !utils.ContainsString(s.SupportedFormats, format.MimeType()) {
		format = tile.TileFormat(s.SupportedFormats[0])
	}
	if s.SupportedSRS != nil {
		var request_srs geo.Proj
		for _, srs := range s.SupportedSRS.Srs {
			if query.Srs.Eq(srs) {
				request_srs = srs
				break
			}
		}
		if request_srs == nil {
			return s.getTransformed(query, format)
		}
		if !query.Srs.Eq(request_srs) {
			query.Srs = request_srs
		}
	}
	if s.Extent != nil && !s.Extent.Contains(&geo.MapExtent{BBox: query.BBox, Srs: query.Srs}) {
		return s.getSubQuery(query, format)
	}
	resp := s.Client.Retrieve(query, &format)
	src := imagery.CreateImageSource(query.Size, s.Options)
	src.SetSource(bytes.NewBuffer(resp))
	return src
}

func (s *WMSSource) getSubQuery(query *layer.MapQuery, format tile.TileFormat) tile.Source {
	size, offset, bbox := imagery.BBoxPositionInImage(query.BBox, query.Size, s.Extent.BBoxFor(query.Srs))
	if size[0] == 0 || size[1] == 0 {
		return imagery.NewBlankImageSource(size, s.Options, nil)
	}
	src_query := &layer.MapQuery{BBox: bbox, Size: size, Srs: query.Srs, Format: format, Dimensions: query.Dimensions}
	resp := s.Client.Retrieve(src_query, &format)
	src := imagery.CreateImageSource(query.Size, s.Options)
	src.SetSource(bytes.NewBuffer(resp))
	return imagery.SubImageSource(src, query.Size, offset[:], s.Options, nil)
}

func (s *WMSSource) getTransformed(query *layer.MapQuery, format tile.TileFormat) tile.Source {
	dst_srs := query.Srs
	src_srs, _ := s.SupportedSRS.BestSrs(dst_srs)
	dst_bbox := query.BBox
	src_bbox := dst_srs.TransformRectTo(src_srs, dst_bbox, 16)

	src_width, src_height := src_bbox.Max[0]-src_bbox.Min[0], src_bbox.Max[1]-src_bbox.Min[1]
	ratio := src_width / src_height
	dst_size := query.Size
	xres, yres := src_width/float64(dst_size[0]), src_height/float64(dst_size[1])
	var src_size [2]uint32
	if xres < yres {
		src_size = [2]uint32{dst_size[0], uint32(float64(dst_size[0])/ratio + 0.5)}
	} else {
		src_size = [2]uint32{uint32(float64(dst_size[1])*ratio + 0.5), dst_size[1]}
	}

	src_query := &layer.MapQuery{BBox: src_bbox, Size: src_size, Srs: src_srs, Format: format, Dimensions: query.Dimensions}
	var img tile.Source
	if s.Coverage != nil && !s.Coverage.Contains(src_bbox, src_srs) {
		img = s.getSubQuery(src_query, format)
	} else {
		resp := s.Client.Retrieve(src_query, &format)
		img = imagery.CreateImageSource(src_size, s.Options)
		img.SetSource(bytes.NewBuffer(resp))
	}

	img = imagery.NewImageTransformer(src_srs, dst_srs, nil).Transform(img, src_bbox,
		query.Size, dst_bbox, s.Options)

	opts := img.GetTileOptions().(*imagery.ImageOptions)
	opts.Format = format
	return img
}

func (s *WMSSource) isCompatible(other *WMSSource, query *layer.MapQuery) bool {
	if s.Opacity != nil || other.Opacity != nil {
		return false
	}

	if !s.SupportedSRS.Eq(other.SupportedSRS) {
		return false
	}

	if !utils.EqualsStrings(s.SupportedFormats, other.SupportedFormats) {
		return false
	}

	sr, sg, sb, sa := s.TransparentColor.RGBA()
	tr, tg, tb, ta := other.TransparentColor.RGBA()

	if sr != tr || sg != tg || sb != tb || sa != ta {
		return false
	}

	if (s.TransparentColorTolerance != nil && other.TransparentColorTolerance == nil) ||
		(s.TransparentColorTolerance == nil && other.TransparentColorTolerance != nil) ||
		(*s.TransparentColorTolerance != *other.TransparentColorTolerance) {
		return false
	}

	if s.Coverage != nil && !s.Coverage.Equals(other.Coverage) {
		return false
	}

	if !layer.EqualsParams(query.DimensionsForParams(s.ExtReqParams), query.DimensionsForParams(other.ExtReqParams)) {
		return false
	}

	return true
}

func (s *WMSSource) CombinedLayer(other layer.Layer, query *layer.MapQuery) layer.Layer {
	o := other.(*WMSSource)
	if !s.isCompatible(o, query) {
		return nil
	}
	c := s.Client.CombinedClient(s.Client, query)
	var wclient client.MapClient
	if c != nil {
		wclient = c
	} else {
		return nil
	}

	return NewWMSSource(wclient, s.Options, s.Coverage, s.ResRange, s.TransparentColor, s.TransparentColorTolerance, s.SupportedSRS, s.SupportedFormats, s.ExtReqParams)
}

type WMSInfoSource struct {
	Client      *client.WMSInfoClient
	Coverage    geo.Coverage
	Transformer func(feature resource.FeatureInfoDoc) resource.FeatureInfoDoc
}

func (s *WMSInfoSource) GetClient() *client.WMSInfoClient {
	return s.Client
}

func (s *WMSInfoSource) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
	if s.Coverage != nil && !s.Coverage.Contains(query.BBox, query.Srs) {
		return nil
	}
	doc := s.Client.GetInfo(query)
	if s.Transformer != nil {
		doc = s.Transformer(doc)
	}
	return doc
}

type WMSLegendSource struct {
	Clients    []client.WMSLegendClient
	Identifier string
	Cache      *resource.LegendCache
	Size       []uint32
	Static     bool
}

func NewWMSLegendSource(id string, clients []client.WMSLegendClient, cache *resource.LegendCache) *WMSLegendSource {
	return &WMSLegendSource{Identifier: id, Clients: clients, Cache: cache}
}

func (s *WMSLegendSource) GetSize() []uint32 {
	if s.Size == nil {
		legend := s.GetLegend(&layer.LegendQuery{Format: "image/png", Scale: -1})
		rect := legend.GetTile().(image.Image).Bounds()
		s.Size = []uint32{uint32(rect.Dx()), uint32(rect.Dy())}
	}
	return s.Size[:]
}

func (s *WMSLegendSource) GetLegend(query *layer.LegendQuery) tile.Source {
	var legend *resource.Legend
	if s.Static {
		legend = &resource.Legend{BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: -1}
	} else {
		legend = &resource.Legend{BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: query.Scale}
	}

	var error_occured bool
	legends := make([]tile.Source, 0)
	if s.Cache.Load(legend) == nil {
		error_occured = false
		for _, client := range s.Clients {
			legends = append(legends, client.GetLegend(query).Source)
		}
	}

	format := request.SplitMimeType(query.Format)[0]
	legend = &resource.Legend{
		Source:       imagery.ConcatLegends(legends, imagery.RGBA, tile.TileFormat(format), nil, nil, false),
		BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: query.Scale,
	}

	if !error_occured {
		s.Cache.Store(legend)
	}
	return legend.Source
}

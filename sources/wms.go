package sources

import (
	"bytes"
	"image/color"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/utils"
)

type WMSSource struct {
	ImagerySource
	Client                    *client.WMSClient
	ImageOpts                 *images.ImageOptions
	SupportsMetaTiles         bool
	SupportedSRS              *geo.SupportedSRS
	SupportedFormats          []string
	ExtReqParams              map[string]string
	TransparentColor          color.Color
	TransparentColorTolerance *float64
	Coverage                  geo.Coverage
	ResRange                  *geo.ResolutionRange
	Extent                    *geo.MapExtent
	ErrorHandler              HTTPSourceErrorHandler
	Opacity                   *float64
}

func NewWMSSource(client *client.WMSClient, image_opts *images.ImageOptions, coverage geo.Coverage, res_range *geo.ResolutionRange,
	transparent_color color.Color, transparent_color_tolerance *float64,
	supported_srs *geo.SupportedSRS, supported_formats []string, fwd_req_params map[string]string,
	error_handler HTTPSourceErrorHandler) *WMSSource {
	src := &WMSSource{Client: client, ImageOpts: image_opts, Coverage: coverage, ResRange: res_range, TransparentColor: transparent_color, TransparentColorTolerance: transparent_color_tolerance, SupportedSRS: supported_srs, SupportedFormats: supported_formats, SupportsMetaTiles: false, ExtReqParams: fwd_req_params, ErrorHandler: error_handler}
	if transparent_color != nil {
		src.ImageOpts.Transparent = geo.NewBool(true)
	}
	if coverage != nil {
		src.Extent = &geo.MapExtent{BBox: coverage.GetBBox(), Srs: coverage.GetSrs()}
	} else {
		src.Extent = geo.MapExtentFromDefault()
	}
	return src
}

func (s *WMSSource) IsOpaque(query layer.MapQuery) bool {
	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}

	if s.ImageOpts.Transparent != nil && *s.ImageOpts.Transparent {
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

func (s *WMSSource) GetMap(query *layer.MapQuery) images.Source {
	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return images.NewBlankImageSource(query.Size, s.ImageOpts, false)
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return images.NewBlankImageSource(query.Size, s.ImageOpts, false)
	}
	resp := s.getMap(query)
	resp.GetImageOptions().Opacity = s.Opacity
	return resp
}

func (s *WMSSource) getMap(query *layer.MapQuery) images.Source {
	format := s.ImageOpts.Format
	if format == "" {
		format = images.ImageFormat(query.Format)
	}
	if s.SupportedFormats != nil && !utils.ContainsString(s.SupportedFormats, format.MimeType()) {
		format = images.ImageFormat(s.SupportedFormats[0])
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
	resp := s.Client.Retrieve(query, format)
	src := images.CreateImageSource(query.Size, s.ImageOpts)
	src.SetSource(bytes.NewBuffer(resp))
	return src
}

func (s *WMSSource) getSubQuery(query *layer.MapQuery, format images.ImageFormat) images.Source {
	size, offset, bbox := images.BBoxPositionInImage(query.BBox, query.Size, s.Extent.BBoxFor(query.Srs))
	if size[0] == 0 || size[1] == 0 {
		return images.NewBlankImageSource(size, s.ImageOpts, false)
	}
	src_query := &layer.MapQuery{BBox: bbox, Size: size, Srs: query.Srs, Format: format, Dimensions: query.Dimensions}
	resp := s.Client.Retrieve(src_query, format)
	src := images.CreateImageSource(query.Size, s.ImageOpts)
	src.SetSource(bytes.NewBuffer(resp))
	return images.SubImageSource(src, query.Size, offset[:], s.ImageOpts, false)
}

func (s *WMSSource) getTransformed(query *layer.MapQuery, format images.ImageFormat) images.Source {
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
	var img images.Source
	if s.Coverage != nil && !s.Coverage.Contains(src_bbox, src_srs) {
		img = s.getSubQuery(src_query, format)
	} else {
		resp := s.Client.Retrieve(src_query, format)
		img = images.CreateImageSource(src_size, s.ImageOpts)
		img.SetSource(bytes.NewBuffer(resp))
	}

	img = images.NewImageTransformer(src_srs, dst_srs, nil).Transform(img, src_bbox,
		query.Size, dst_bbox, s.ImageOpts)

	img.GetImageOptions().Format = format
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

func (s *WMSSource) CombinedLayer(other *WMSSource, query *layer.MapQuery) *WMSSource {
	if !s.isCompatible(other, query) {
		return nil
	}
	c := s.Client.CombinedClient(s.Client, query)
	var wclient *client.WMSClient
	if c != nil {
		wc, ok := c.(*client.WMSClient)
		if !ok {
			return nil
		}
		wclient = wc
	} else {
		return nil
	}

	return NewWMSSource(wclient, s.ImageOpts, s.Coverage, s.ResRange, s.TransparentColor, s.TransparentColorTolerance, s.SupportedSRS, s.SupportedFormats, s.ExtReqParams, s.ErrorHandler)
}

type WMSInfoSource struct {
	InfoSource
	Client      *client.WMSInfoClient
	Coverage    geo.Coverage
	Transformer func(feature *resource.FeatureInfo) *resource.FeatureInfo
}

func (s *WMSInfoSource) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
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
	LegendSource
	Clients    []client.WMSLegendClient
	Identifier string
	Cache      *resource.LegendCache
	Size       []uint32
	Static     bool
}

func (s *WMSLegendSource) GetSize() []uint32 {
	if s.Size == nil {
		legend := s.GetLegend(&layer.LegendQuery{Format: "image/png", Scale: -1})
		rect := legend.GetImage().Bounds()
		s.Size = []uint32{uint32(rect.Dx()), uint32(rect.Dy())}
	}
	return s.Size[:]
}

func (s *WMSLegendSource) GetLegend(query *layer.LegendQuery) images.Source {
	var legend *resource.Legend
	if s.Static {
		legend = &resource.Legend{BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: -1}
	} else {
		legend = &resource.Legend{BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: query.Scale}
	}
	var error_occured bool
	legends := make([]images.Source, 0)
	if s.Cache.Load(legend) == nil {
		error_occured = false
		for _, client := range s.Clients {
			legends = append(legends, client.GetLegend(query).Source)
		}
	}

	format := request.SplitMimeType(query.Format)[0]
	legend = &resource.Legend{Source: images.ConcatLegends(legends, images.RGBA, images.ImageFormat(format), nil, nil, false),
		BaseResource: resource.BaseResource{ID: s.Identifier}, Scale: query.Scale}

	if !error_occured {
		s.Cache.Store(legend)
	}
	return legend.Source
}

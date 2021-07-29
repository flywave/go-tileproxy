package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	vec2d "github.com/flywave/go3d/float64/vec2"
	_ "github.com/flywave/ogc-specifications/pkg/wms130"
)

type WMSService struct {
	BaseService
	RootLayer       *WMSGroupLayer
	Layers          map[string]wmsLayer
	Strict          bool
	ImageFormats    map[string]*images.ImageOptions
	TileLayers      []*TileLayer
	Metadata        map[string]string
	InfoFormats     map[string]string
	Srs             geo.Proj
	SrsExtents      map[string]*geo.MapExtent
	MaxOutputPixels int
	MaxTileAge      time.Duration
	InspireMetadata map[string]string
}

func (s *WMSService) GetMap(req request.Request) *Response {
	s.checkMapRequest(req)

	params := req.GetParams()
	map_request := req.(*request.WMSRequest)
	mapreq := request.NewWMSMapRequestParams(params)
	query := &layer.MapQuery{BBox: mapreq.GetBBox(), Size: mapreq.GetSize(), Srs: geo.NewSRSProj4(mapreq.GetSrs()), Format: mapreq.GetFormat()}

	if params.GetOne("tiled", "false") == "true" {
		query.TiledOnly = true
	}
	orig_query := query

	var offset [2]uint32
	var sub_size [2]uint32
	var sub_bbox vec2d.Rect

	if _, ok := s.SrsExtents[mapreq.GetSrs()]; s.SrsExtents != nil && ok {
		query_extent := &geo.MapExtent{BBox: mapreq.GetBBox(), Srs: geo.NewSRSProj4(mapreq.GetSrs())}
		if !s.SrsExtents[mapreq.GetSrs()].Contains(query_extent) {
			limited_extent := s.SrsExtents[mapreq.GetSrs()].Intersection(query_extent)
			if limited_extent == nil {
				img_opts := s.ImageFormats[mapreq.GetFormatMimeType()]
				img_opts.BgColor = mapreq.GetBGColor()
				img_opts.Transparent = geo.NewBool(mapreq.GetTransparent())
				img := images.NewBlankImageSource(mapreq.GetSize(), img_opts, nil)

				return NewResponse(img.GetBuffer(nil, nil), 200, "", img_opts.Format.MimeType())
			}
			sub_size, offset, sub_bbox = images.BBoxPositionInImage(mapreq.GetBBox(), mapreq.GetSize(), limited_extent.BBox)
			query = &layer.MapQuery{BBox: sub_bbox, Size: sub_size, Srs: geo.NewSRSProj4(mapreq.GetSrs()), Format: mapreq.GetFormat()}
		}
	}

	actual_layers := make(map[string]wmsLayer)
	for _, layer_name := range mapreq.GetLayers() {
		layer := s.Layers[layer_name]
		if layer.rendersQuery(query) {
			if layer.IsOpaque(query) {
				actual_layers = make(map[string]wmsLayer)
			}
			for layer_name, map_layers := range layer.mapLayersForQuery(query) {
				actual_layers[layer_name] = map_layers
			}
		}
	}

	var keys []string
	for k := range actual_layers {
		keys = append(keys, k)
	}

	authorized_layers, coverage := s.authorizedLayers("map", keys, &geo.MapExtent{BBox: mapreq.GetBBox(), Srs: geo.NewSRSProj4(mapreq.GetSrs())})

	s.filterActualLayers(actual_layers, mapreq.GetLayers(), authorized_layers)

	render_layers := []wmsLayer{}
	for _, v := range actual_layers {
		render_layers = append(render_layers, v)
	}

	s.updateQueryWithFWDParams(query, mapreq, render_layers)

	renderer := &LayerRenderer{layers: render_layers, query: query, params: mapreq}

	merger := &images.LayerMerger{}
	renderer.render(merger)

	img_opts := s.ImageFormats[mapreq.GetFormatMimeType()]
	img_opts.BgColor = mapreq.GetBGColor()
	img_opts.Transparent = geo.NewBool(mapreq.GetTransparent())
	si := mapreq.GetSize()
	result := merger.Merge(img_opts, si[:], mapreq.GetBBox(), geo.NewSRSProj4(mapreq.GetSrs()), coverage)

	if !query.EQ(orig_query) {
		imageSrc := result.(*images.ImageSource)
		result = images.SubImageSource(imageSrc, orig_query.Size, offset[:], img_opts, nil)
	}

	result = s.DecorateImg(result, "wms.map", keys, &geo.MapExtent{Srs: geo.NewSRSProj4(mapreq.GetSrs()), BBox: mapreq.GetBBox()})
	imagesource := result.(*images.ImageSource)

	imagesource.SetGeoReference(geo.NewGeoReference(mapreq.GetBBox(), geo.NewSRSProj4(mapreq.GetSrs())))
	result_buf := result.GetBuffer(nil, img_opts)
	f := img_opts.GetFormat()
	resp := NewResponse(result_buf, 200, "", f.MimeType())

	if query.TiledOnly && result.GetCacheable() != nil {
		cache_info := result.GetCacheable()
		resp.cacheHeaders(&cache_info.Timestamp, []string{cache_info.Timestamp.String(), strconv.Itoa(int(cache_info.Size))},
			int(s.MaxTileAge.Seconds()))
		resp.makeConditional(map_request.Http)
	}

	if result.GetCacheable() == nil {
		resp.noCacheHeaders()
	}

	return resp
}

func (s *WMSService) authorizedLayers(feature string, layers []string, ext *geo.MapExtent) ([]*TileLayer, geo.Coverage) {
	return nil, nil
}

func (s *WMSService) GetCapabilities(req request.Request) *Response {
	params := req.GetParams()
	map_request := req.(*request.WMSRequest)

	var tile_layers []*TileLayer

	if strings.ToLower(params.GetOne("tiled", "false")) == "true" {
		tile_layers = s.TileLayers
	} else {
		tile_layers = nil
	}

	service := s.serviceMetadata(req)
	root_layer := s.authorizedCapabilityLayers()

	info_types := []string{"text", "html", "xml"}
	if s.InfoFormats != nil {
		info_types = s.InfoFormats
	} else if self.fi_transformers {
		info_types = self.fi_transformers.keys()
	}
	info_formats := []string{}
	for i := range info_types {
		info_formats = append(info_formats, mimetype_from_infotype(info_types[i]))
	}

	cap := newCapabilities(service, root_layer, tile_layers,
		s.ImageFormats, info_formats, s.Srs, s.SrsExtents,
		s.InspireMetadata, s.MaxOutputPixels)
	result := cap.render(map_request)

	return NewResponse(result, 200, "", "application/xml")
}

func (s *WMSService) GetFeatureInfo(req request.Request) *Response {
	infos := []*resource.FeatureInfo{}
	s.checkFeatureinfoRequest(req)
	p := req.GetParams()
	freq := request.NewWMSFeatureInfoRequestParams(req.GetParams())

	var feature_count *int

	if req.GetParams() != nil {
		if v, ok := req.GetParams()["feature_count"]; ok {
			fc, _ := strconv.Atoi(v[0])
			feature_count = &fc
		}
	}

	query := &layer.InfoQuery{BBox: freq.GetBBox(), Size: [2]uint32{freq.GetSize()[0], freq.GetSize()[1]}, Srs: geo.NewSRSProj4(freq.GetSrs()), Pos: info_request.Pos,
		InfoFormat: freq.GetFormat(), FeatureCount: feature_count}

	actual_layers := make(map[string]wmsLayer)

	for layer_name := range freq.GetLayers() {
		layer = self.layers[layer_name]
		if !layer.Queryable() {
			//raise RequestError('layer %s is not queryable' % layer_name, request=request)
		}
		for layer_name, map_layers := range layer.infoLayersForQuery(query) {
			actual_layers[layer_name] = map_layers
		}
	}

	var keys []string
	for k := range actual_layers {
		keys = append(keys, k)
	}

	authorized_layers, coverage := s.authorizedLayers("featureinfo", keys, &geo.MapExtent{BBox: freq.GetBBox(), Srs: geo.NewSRSProj4(freq.GetSrs())})
	s.filterActualLayers(actual_layers, freq.GetLayers(), authorized_layers)

	if coverage != nil && !coverage.ContainsPoint(query.GetCoord(), query.Srs) {
		infos = nil
	} else {
		for _, source := range tile_layer.infoSources {
			info := source.GetInfo(query)
			if info == nil {
				continue
			}
			infos = append(infos, info)
		}
	}

	mimetype := info_request.Infoformat

	if infos == nil || len(infos) == 0 {
		return NewResponse([]byte{}, 200, "", mimetype)
	}

	var resp []byte
	var info_type string

	if self.fi_transformers {
		if !mimetype {
			if utils.ContainsString(self.fi_transformers, "xml") {
				info_type = "xml"
			} else if utils.ContainsString(self.fi_transformers, "html") {
				info_type = "html"
			} else {
				info_type = "text"
			}
			mimetype = mimetype_from_infotype(request.version, info_type)
		} else {
			info_type = infotype_from_mimetype(request.version, mimetype)
		}
		resp, actual_info_type = combine_docs(infos, self.fi_transformers[info_type])
		if actual_info_type != nil && info_type != actual_info_type {
			mimetype = mimetype_from_infotype(request.version, actual_info_type)
		}
	} else {
		resp, info_type = combine_docs(infos)
		mimetype = mimetype_from_infotype(request.version, info_type)
	}

	return NewResponse(resp, 200, "", mimetype)
}

func (s *WMSService) checkMapRequest(req request.Request) {

}

func (s *WMSService) checkFeatureinfoRequest(req request.Request) {

}

func (s *WMSService) updateQueryWithFWDParams(query *layer.MapQuery, params request.WMSMapRequestParams, layers []wmsLayer) {

}

func (s *WMSService) validateLayers(req request.Request) {

}

func (s *WMSService) checkLegendRequest(req request.Request) {

}

func (s *WMSService) Legendgraphic(req request.Request) {

}
func (s *WMSService) serviceMetadata(tms_request request.Request) map[string]string {
	req := tms_request.(*request.BaseRequest)
	md := s.Metadata
	md["url"] = req.Http.URL.Host
	if s.RootLayer.hasLegend {
		md["has_legend"] = "true"
	} else {
		md["has_legend"] = "false"
	}
	return md
}

func (s *WMSService) filterActualLayers(actual_layers map[string]wmsLayer, layers []string, authorized_layers []*TileLayer) {

}

func (s *WMSService) authorizedCapabilityLayers() {

}

type WMSCapabilities struct {
}

type LayerRenderer struct {
	layers []wmsLayer
	query  *layer.MapQuery
	params request.WMSMapRequestParams
}

func (l *LayerRenderer) render(layer *images.LayerMerger) {

}

type wmsLayer interface {
	layer.Layer
	rendersQuery(query *layer.MapQuery) bool
	mapLayersForQuery(query *layer.MapQuery) map[string]wmsLayer
	infoLayersForQuery(query *layer.InfoQuery) []wmsLayer
	legend(query *layer.LegendQuery) []tile.Source
	GetLegendSize() int
	GetName() string
	GetLegendUrl() string
	HasLegend() bool
	Queryable() bool
}

type WMSLayerBase struct {
	wmsLayer
	name       string
	title      string
	isActive   bool
	layers     map[string]wmsLayer
	metadata   map[string]string
	queryable  bool
	hasLegend  bool
	legendUrl  string
	legendSize []int
	resRange   *geo.ResolutionRange
	extent     *geo.MapExtent
}

type WMSLayer struct {
	WMSLayerBase
	mapLayers    map[string]wmsLayer
	infoLayers   []wmsLayer
	legendLayers []wmsLayer
}

func NewWMSLayer(name string, title string, map_layers map[string]wmsLayer, infos []wmsLayer, legends []wmsLayer, res_range *geo.ResolutionRange, md map[string]string) *WMSLayer {
	queryable := false
	if len(infos) > 0 {
		queryable = true
	}

	has_legend := false
	if len(legends) > 0 {
		has_legend = true
	}

	ret := &WMSLayer{WMSLayerBase: WMSLayerBase{name: name, title: title, metadata: md, isActive: false, layers: nil, hasLegend: has_legend, queryable: queryable}, mapLayers: map_layers, infoLayers: infos, legendLayers: legends}

	ret.extent = mergeLayerExtents(map_layers)
	if res_range == nil {
		ret.resRange = mergeLayerResRanges(map_layers)
	} else {
		ret.resRange = res_range
	}

	return ret
}

func (l *WMSLayer) IsOpaque(query *layer.MapQuery) bool {
	for i := range l.mapLayers {
		if l.mapLayers[i].IsOpaque(query) {
			return true
		}
	}
	return false
}

func (l *WMSLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSLayer) mapLayersForQuery(query *layer.MapQuery) map[string]wmsLayer {
	if l.mapLayers == nil {
		return nil
	}
	return l.mapLayers
}

func (l *WMSLayer) infoLayersForQuery(query *layer.InfoQuery) []wmsLayer {
	if l.infoLayers == nil {
		return nil
	}
	return l.infoLayers
}

func (l *WMSLayer) legend(query *layer.LegendQuery) []tile.Source {
	legend := []tile.Source{}
	for _, lyr := range l.legendLayers {
		legend = append(legend, lyr.legend(query)...)
	}
	return legend
}

type WMSGroupLayer struct {
	WMSLayerBase
	this wmsLayer
}

func mergeLayerResRanges(layers map[string]wmsLayer) *geo.ResolutionRange {
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

func mergeLayerExtents(layers map[string]wmsLayer) *geo.MapExtent {
	if layers == nil || len(layers) == 0 {
		return geo.MapExtentFromDefault()
	}
	var extent *geo.MapExtent
	for _, v := range layers {
		if extent == nil {
			extent = v.GetExtent()
		} else {
			extent = extent.Add(v.GetExtent())
		}
	}
	return extent
}

func cloneLayers(tags map[string]wmsLayer) map[string]wmsLayer {
	cloneTags := make(map[string]wmsLayer)
	for k, v := range tags {
		cloneTags[k] = v
	}
	return cloneTags
}

func NewWMSGroupLayer(name string, title string, this wmsLayer, layers map[string]wmsLayer, md map[string]string) *WMSGroupLayer {
	is_active := false
	if this != nil {
		is_active = true
	}
	has_legend := false
	if this != nil && this.HasLegend() {
		has_legend = true
	} else {
		for _, l := range layers {
			if l.HasLegend() {
				has_legend = true
			}
		}
	}
	queryable := false
	if this != nil && this.Queryable() {
		queryable = true
	} else {
		for _, l := range layers {
			if l.Queryable() {
				queryable = true
			}
		}
	}

	ret := &WMSGroupLayer{WMSLayerBase: WMSLayerBase{name: name, title: title, metadata: md, isActive: is_active, layers: layers, hasLegend: has_legend, queryable: queryable}, this: this}

	all_layers := cloneLayers(layers)
	all_layers[this.GetName()] = this
	ret.extent = mergeLayerExtents(all_layers)
	ret.resRange = mergeLayerResRanges(all_layers)

	return ret
}

func (l *WMSGroupLayer) IsOpaque(query *layer.MapQuery) bool {
	for i := range l.layers {
		if l.layers[i].IsOpaque(query) {
			return true
		}
	}
	return false
}

func (l *WMSGroupLayer) GetLegendSize() int {
	return l.this.GetLegendSize()
}

func (l *WMSGroupLayer) GetLegendUrl() string {
	return l.this.GetLegendUrl()
}

func (l *WMSGroupLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSGroupLayer) mapLayersForQuery(query *layer.MapQuery) map[string]wmsLayer {
	if l.this != nil {
		return l.this.mapLayersForQuery(query)
	} else {
		layers := make(map[string]wmsLayer)
		for _, layer := range l.layers {
			ll := layer.mapLayersForQuery(query)
			for name, l := range ll {
				layers[name] = l
			}
		}
		return layers
	}
}

func (l *WMSGroupLayer) infoLayersForQuery(query *layer.InfoQuery) []wmsLayer {
	if l.this != nil {
		return l.this.infoLayersForQuery(query)
	} else {
		layers := []wmsLayer{}
		for _, layer := range l.layers {
			layers = append(layers, layer.infoLayersForQuery(query)...)
		}
		return layers
	}
}

func (l *WMSGroupLayer) GetChildLayers() map[string]wmsLayer {
	layers := make(map[string]wmsLayer)
	if l.name != "" {
		layers[l.name] = l
	}
	for _, lyr := range l.layers {
		if iface, ok := lyr.(interface {
			GetChildLayers() map[string]wmsLayer
		}); ok {
			new := iface.GetChildLayers()
			for k, v := range new {
				layers[k] = v
			}
		} else if lyr.GetName() != "" {
			layers[lyr.GetName()] = lyr
		}
	}
	return layers
}

func (l *WMSGroupLayer) legend(query *layer.LegendQuery) []tile.Source {
	panic("not implemented")
}

func CombinedLayers(layers []layer.Layer, query *layer.MapQuery) []layer.Layer {
	if len(layers) <= 1 {
		return layers
	}
	combined_layers := []layer.Layer{layers[0]}
	layers = layers[1:]
	for i := range layers {
		current_layer := layers[i]
		combined := combined_layers[len(combined_layers)-1].CombinedLayer(current_layer, query)
		if combined != nil {
			combined_layers[len(combined_layers)-1] = combined
		} else {
			combined_layers = append(combined_layers, current_layer)
		}
	}
	return combined_layers
}

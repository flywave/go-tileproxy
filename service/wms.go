package service

import (
	"errors"
	"image/color"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type WMSService struct {
	BaseService
	RootLayer           *WMSGroupLayer
	Layers              map[string]wmsLayer
	Strict              bool
	ImageFormats        map[string]*imagery.ImageOptions
	TileLayers          []*TileProvider
	Metadata            map[string]string
	InfoFormats         map[string]string
	Srs                 *geo.SupportedSRS
	SrsExtents          map[string]*geo.MapExtent
	MaxOutputPixels     int
	MaxTileAge          time.Duration
	FeatureTransformers map[string]*resource.XSLTransformer
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
				tile := cache.GetEmptyTile(mapreq.GetSize(), img_opts)

				return NewResponse(tile.GetBuffer(nil, nil), 200, img_opts.Format.MimeType())
			}
			sub_size, offset, sub_bbox = imagery.BBoxPositionInImage(mapreq.GetBBox(), mapreq.GetSize(), limited_extent.BBox)
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

	render_layers := []layer.Layer{}
	for _, v := range actual_layers {
		render_layers = append(render_layers, v)
	}

	renderer := &LayerRenderer{layers: render_layers, query: query, params: mapreq}

	merger := &imagery.LayerMerger{}
	renderer.render(merger)

	img_opts := s.ImageFormats[mapreq.GetFormatMimeType()]
	img_opts.BgColor = mapreq.GetBGColor()
	img_opts.Transparent = geo.NewBool(mapreq.GetTransparent())
	si := mapreq.GetSize()
	result := merger.Merge(img_opts, si[:], mapreq.GetBBox(), geo.NewSRSProj4(mapreq.GetSrs()), coverage)

	if !query.EQ(orig_query) {
		imageSrc := result.(*imagery.ImageSource)
		result = imagery.SubImageSource(imageSrc, orig_query.Size, offset[:], img_opts, nil)
	}

	result = s.DecorateTile(result, "wms.map", keys, &geo.MapExtent{Srs: geo.NewSRSProj4(mapreq.GetSrs()), BBox: mapreq.GetBBox()})
	imagesource := result.(*imagery.ImageSource)

	imagesource.SetGeoReference(geo.NewGeoReference(mapreq.GetBBox(), geo.NewSRSProj4(mapreq.GetSrs())))
	result_buf := result.GetBuffer(nil, img_opts)
	f := img_opts.GetFormat()
	resp := NewResponse(result_buf, 200, f.MimeType())

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

func (s *WMSService) authorizedLayers(feature string, layers []string, ext *geo.MapExtent) ([]*TileProvider, geo.Coverage) {
	return s.TileLayers, nil
}

func (s *WMSService) GetCapabilities(req request.Request) *Response {
	params := req.GetParams()
	map_request := req.(*request.WMSRequest)

	var tile_layers []*TileProvider

	if strings.ToLower(params.GetOne("tiled", "false")) == "true" {
		tile_layers = s.TileLayers
	} else {
		tile_layers = nil
	}

	service := s.serviceMetadata(req)
	root_layer := s.authorizedCapabilityLayers()

	info_types := []string{"text", "html", "xml"}
	if s.InfoFormats != nil {
		for _, v := range s.InfoFormats {
			info_types = append(info_types, v)
		}
	} else if s.FeatureTransformers != nil {
		info_types = []string{}
		for k := range s.FeatureTransformers {
			info_types = append(info_types, k)
		}
	}
	info_formats := []string{}
	for i := range info_types {
		info_formats = append(info_formats, request.MimetypeFromInfotype(info_types[i]))
	}

	image_formats := []string{}
	for k := range s.ImageFormats {
		image_formats = append(info_formats, k)
	}

	cap := newCapabilities(service, root_layer, tile_layers, image_formats, info_formats, s.Srs, s.SrsExtents, s.MaxOutputPixels)
	result := cap.render(map_request)

	return NewResponse(result, 200, "application/xml")
}

func (s *WMSService) GetFeatureInfo(req request.Request) *Response {
	infos := []resource.FeatureInfoDoc{}
	s.checkFeatureinfoRequest(req)

	freq := request.NewWMSFeatureInfoRequestParams(req.GetParams())

	var feature_count *int

	if req.GetParams() != nil {
		if v, ok := req.GetParams()["feature_count"]; ok {
			fc, _ := strconv.Atoi(v[0])
			feature_count = &fc
		}
	}

	query := &layer.InfoQuery{BBox: freq.GetBBox(), Size: [2]uint32{freq.GetSize()[0], freq.GetSize()[1]}, Srs: geo.NewSRSProj4(freq.GetSrs()), Pos: freq.GetPos(),
		InfoFormat: string(freq.GetFormat()), FeatureCount: feature_count}

	actual_layers := make(map[string]wmsLayer)

	for _, layer_name := range freq.GetLayers() {
		layer := s.Layers[layer_name]
		if !layer.Queryable() {
			return NewResponse(nil, 400, DefaultContentType)
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
		info_layers := []*TileProvider{}
		for _, layers := range authorized_layers {
			info_layers = append(info_layers, layers)
		}
		for _, layer := range info_layers {
			for _, source := range layer.infoSources {
				info := source.GetInfo(query)
				if info == nil {
					continue
				}
				infos = append(infos, info)
			}
		}
	}

	mimetype := freq.GetFormatString()

	if infos == nil || len(infos) == 0 {
		return NewResponse([]byte{}, 200, mimetype)
	}

	var resp []byte
	var info_type string
	var actual_info_type string
	if s.FeatureTransformers != nil {
		if mimetype != "" {
			if _, ok := s.FeatureTransformers["xml"]; ok {
				info_type = "xml"
			} else if _, ok := s.FeatureTransformers["html"]; ok {
				info_type = "html"
			} else {
				info_type = "text"
			}
			mimetype = request.MimetypeFromInfotype(info_type)
		} else {
			info_type = request.InfotypeFromMimetype(mimetype)
		}
		resp, actual_info_type = resource.CombineDocs(infos, s.FeatureTransformers[info_type])
		if actual_info_type != "" && info_type != actual_info_type {
			mimetype = request.MimetypeFromInfotype(actual_info_type)
		}
	} else {
		resp, info_type = resource.CombineDocs(infos, nil)
		mimetype = request.MimetypeFromInfotype(info_type)
	}

	return NewResponse(resp, 200, mimetype)
}

func (s *WMSService) checkMapRequest(req request.Request) error {
	mapreq := req.(*request.WMSMapRequest)
	mapparams := request.NewWMSMapRequestParams(req.GetParams())
	si := mapparams.GetSize()
	if s.MaxOutputPixels != -1 && (si[0]*si[1]) > uint32(s.MaxOutputPixels) {
		return errors.New("image size too large")
	}

	s.validateLayers(req)

	formats := []string{}
	for k := range s.ImageFormats {
		formats = append(formats, k)
	}

	mapreq.ValidateFormat(formats)

	srss := []string{}
	for _, s := range s.Srs.Srs {
		srss = append(srss, s.GetDef())
	}

	mapreq.ValidateSrs(srss)
	return nil
}

func (s *WMSService) checkFeatureinfoRequest(req request.Request) {
	mapreq := req.(*request.WMSMapRequest)
	s.validateLayers(req)

	srss := []string{}
	for _, s := range s.Srs.Srs {
		srss = append(srss, s.GetDef())
	}

	mapreq.ValidateSrs(srss)
}

func (s *WMSService) validateLayers(req request.Request) error {
	mapparams := request.NewWMSMapRequestParams(req.GetParams())
	query_layers := mapparams.GetLayers()
	for _, layer := range query_layers {
		if _, ok := s.Layers[layer]; !ok {
			return errors.New("unknown layer: " + layer)
		}
	}
	return nil
}

func (s *WMSService) checkLegendRequest(req request.Request) error {
	mapparams := request.NewWMSLegendGraphicRequestParams(req.GetParams())
	layer := mapparams.GetLayer()
	if _, ok := s.Layers[layer]; !ok {
		return errors.New("unknown layer: " + layer)
	}
	return nil
}

func (s *WMSService) Legendgraphic(req request.Request) *Response {
	mapparams := request.NewWMSLegendGraphicRequestParams(req.GetParams())
	layer := mapparams.GetLayer()
	mapparams.GetFormatMimeType()

	s.checkLegendRequest(req)
	if !s.Layers[layer].HasLegend() {
		return NewResponse(nil, 400, DefaultContentType)
	}
	legends := s.Layers[layer].legend(req.(*request.WMSLegendGraphicRequest))

	result := imagery.ConcatLegends(legends, imagery.RGBA, tile.TileFormat("png"), nil, color.White, true)
	mimetype := mapparams.GetFormatMimeType()
	if mimetype == "" {
		mimetype = "image/png"
	}
	img_opts := s.ImageFormats[mimetype]
	return NewResponse(result.GetBuffer(nil, img_opts), 200, mimetype)
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

func (s *WMSService) filterActualLayers(actual_layers map[string]wmsLayer, layers []string, authorized_layers []*TileProvider) {

}

func (s *WMSService) authorizedCapabilityLayers() *WMSGroupLayer {
	return s.RootLayer
}

type LayerRenderer struct {
	layers []layer.Layer
	query  *layer.MapQuery
	params request.WMSMapRequestParams
}

func (l *LayerRenderer) render(layer_merger *imagery.LayerMerger) {
	render_layers := CombinedLayers(l.layers, l.query)
	if render_layers == nil {
		return
	}
	for i := range render_layers {
		layer, layer_img := l.renderLayer(render_layers[i])
		if layer_img != nil {
			layer_merger.AddSource(layer_img, layer.GetCoverage())
		}
	}
}

func (l *LayerRenderer) renderLayer(layer layer.Layer) (layer.Layer, tile.Source) {
	layer_img, _ := layer.GetMap(l.query)
	return layer, layer_img
}

type wmsLayer interface {
	layer.Layer
	rendersQuery(query *layer.MapQuery) bool
	mapLayersForQuery(query *layer.MapQuery) map[string]wmsLayer
	infoLayersForQuery(query *layer.InfoQuery) map[string]wmsLayer
	legend(query *request.WMSLegendGraphicRequest) []tile.Source
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

func (l *WMSLayer) legend(query *request.WMSLegendGraphicRequest) []tile.Source {
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

func (l *WMSGroupLayer) infoLayersForQuery(query *layer.InfoQuery) map[string]wmsLayer {
	if l.this != nil {
		return l.this.infoLayersForQuery(query)
	} else {
		layers := make(map[string]wmsLayer)
		for _, layer := range l.layers {
			ll := layer.infoLayersForQuery(query)
			for name, l := range ll {
				layers[name] = l
			}
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

func (l *WMSGroupLayer) legend(query *request.WMSLegendGraphicRequest) []tile.Source {
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

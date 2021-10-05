package service

import (
	"fmt"
	"image/color"
	"net/http"
	"strconv"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/ogc-specifications/pkg/wms130"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type WMSMetadata struct {
	Title          string
	Abstract       string
	KeywordList    []string
	URL            string
	OnlineResource struct {
		Xlink *string
		Type  *string
		Href  *string
	}
	Fees              *string
	AccessConstraints *string
	HasLegend         bool
	Extended          *wms130.ExtendedCapabilities
	Contact           *wms130.ContactInformation
}

type WMSService struct {
	BaseService
	RootLayer           *WMSGroupLayer
	Layers              map[string]WMSLayer
	Strict              bool
	ImageFormats        map[string]*imagery.ImageOptions
	Metadata            *WMSMetadata
	InfoFormats         map[string]string
	Srs                 *geo.SupportedSRS
	SrsExtents          map[string]*geo.MapExtent
	MaxOutputPixels     int
	MaxTileAge          *time.Duration
	FeatureTransformers map[string]*resource.XSLTransformer
}

func NewWMSService(rootLayer *WMSGroupLayer, layers map[string]WMSLayer, metadata *WMSMetadata, srs *geo.SupportedSRS, imageFormats map[string]*imagery.ImageOptions, infoFormats map[string]string, srsExtents map[string]*geo.MapExtent, maxOutputPixels int, maxTileAge *time.Duration, strict bool, ftransformers map[string]*resource.XSLTransformer) *WMSService {
	ret := &WMSService{RootLayer: rootLayer, Strict: strict, ImageFormats: imageFormats, Metadata: metadata, InfoFormats: infoFormats, Srs: srs, SrsExtents: srsExtents, MaxOutputPixels: maxOutputPixels, MaxTileAge: maxTileAge, FeatureTransformers: ftransformers}
	if rootLayer == nil {
		ret.Layers = layers
	} else {
		ret.Layers = ret.RootLayer.layers
	}

	ret.router = map[string]func(r request.Request) *Response{
		"map": func(r request.Request) *Response {
			return ret.GetMap(r)
		},
		"featureinfo": func(r request.Request) *Response {
			return ret.GetFeatureInfo(r)
		},
		"capabilities": func(r request.Request) *Response {
			return ret.GetCapabilities(r)
		},
		"legendgraphic": func(r request.Request) *Response {
			return ret.Legendgraphic(r)
		},
	}
	ret.requestParser = func(r *http.Request) request.Request {
		return request.MakeWMSRequest(r, false)
	}
	return ret
}

func (s *WMSService) GetMap(req request.Request) *Response {
	err := s.checkMapRequest(req)
	if err != nil {
		return err.Render()
	}

	params := req.GetParams()
	map_request := req.(*request.WMSMapRequest)
	mapreq := request.NewWMSMapRequestParams(params)
	query := &layer.MapQuery{BBox: mapreq.GetBBox(), Size: mapreq.GetSize(), Srs: geo.NewProj(mapreq.GetCrs()), Format: mapreq.GetFormat()}

	if params.GetOne("tiled", "false") == "true" {
		query.TiledOnly = true
	}
	orig_query := query

	var offset [2]uint32
	var sub_size [2]uint32
	var sub_bbox vec2d.Rect

	if _, ok := s.SrsExtents[mapreq.GetCrs()]; s.SrsExtents != nil && ok {
		query_extent := &geo.MapExtent{BBox: mapreq.GetBBox(), Srs: geo.NewProj(mapreq.GetCrs())}
		if !s.SrsExtents[mapreq.GetCrs()].Contains(query_extent) {
			limited_extent := s.SrsExtents[mapreq.GetCrs()].Intersection(query_extent)
			if limited_extent == nil {
				img_opts := s.ImageFormats[mapreq.GetFormatMimeType()]
				img_opts.BgColor = mapreq.GetBGColor()
				img_opts.Transparent = geo.NewBool(mapreq.GetTransparent())
				tile := cache.GetEmptyTile(mapreq.GetSize(), img_opts)
				return NewResponse(tile.GetBuffer(nil, nil), 200, img_opts.Format.MimeType())
			}
			sub_size, offset, sub_bbox = imagery.BBoxPositionInImage(mapreq.GetBBox(), mapreq.GetSize(), limited_extent.BBox)
			query = &layer.MapQuery{BBox: sub_bbox, Size: sub_size, Srs: geo.NewProj(mapreq.GetCrs()), Format: mapreq.GetFormat()}
		}
	}

	actual_layers := make(map[string]layer.Layer)
	for _, layer_name := range mapreq.GetLayers() {
		l := s.Layers[layer_name]
		if l.rendersQuery(query) {
			for layer_name, map_layers := range l.mapLayersForQuery(query) {
				actual_layers[layer_name] = map_layers
			}
		}
	}

	var keys []string
	for k := range actual_layers {
		keys = append(keys, k)
	}

	authorized_layers, coverage := s.authorizedLayers("map", keys, &geo.MapExtent{BBox: mapreq.GetBBox(), Srs: geo.NewProj(mapreq.GetCrs())})

	s.filterActualLayers(actual_layers, mapreq.GetLayers(), authorized_layers)

	render_layers := []layer.Layer{}
	for _, v := range actual_layers {
		render_layers = append(render_layers, v)
	}

	renderer := &LayerRenderer{layers: render_layers, query: query, params: mapreq}

	merger := &imagery.LayerMerger{}
	err = renderer.render(merger)
	if err != nil {
		return err.Render()
	}

	img_opts := s.ImageFormats[mapreq.GetFormatMimeType()]
	img_opts.BgColor = mapreq.GetBGColor()
	img_opts.Transparent = geo.NewBool(mapreq.GetTransparent())
	si := mapreq.GetSize()
	result := merger.Merge(img_opts, si[:], mapreq.GetBBox(), geo.NewProj(mapreq.GetCrs()), coverage)

	if !query.Eq(orig_query) {
		imageSrc := result.(*imagery.ImageSource)
		result = imagery.SubImageSource(imageSrc, orig_query.Size, offset[:], img_opts, nil)
	}

	result = s.DecorateTile(result, "wms.map", keys, &geo.MapExtent{Srs: geo.NewProj(mapreq.GetCrs()), BBox: mapreq.GetBBox()})
	imagesource := result.(*imagery.ImageSource)

	imagesource.SetGeoReference(geo.NewGeoReference(mapreq.GetBBox(), geo.NewProj(mapreq.GetCrs())))
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

func (s *WMSService) authorizedLayers(feature string, layers []string, ext *geo.MapExtent) ([]string, geo.Coverage) {
	return layers, nil
}

func (s *WMSService) GetCapabilities(req request.Request) *Response {
	map_request := req.(*request.WMSRequest)

	service := s.serviceMetadata(req, s.Metadata)
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

	cap := newCapabilities(&service, root_layer, image_formats, info_formats, s.Srs, s.SrsExtents, s.MaxOutputPixels)
	result := cap.render(map_request)

	return NewResponse(result, 200, "application/xml")
}

func (s *WMSService) GetFeatureInfo(req request.Request) *Response {
	infos := []resource.FeatureInfoDoc{}
	err := s.checkFeatureinfoRequest(req)
	if err != nil {
		return err.Render()
	}
	freq := request.NewWMSFeatureInfoRequestParams(req.GetParams())

	var feature_count *int

	if req.GetParams() != nil {
		if v, ok := req.GetParams()["feature_count"]; ok {
			fc, _ := strconv.Atoi(v[0])
			feature_count = &fc
		}
	}

	query := &layer.InfoQuery{BBox: freq.GetBBox(), Size: [2]uint32{freq.GetSize()[0], freq.GetSize()[1]}, Srs: geo.NewProj(freq.GetCrs()), Pos: freq.GetPos(),
		InfoFormat: string(freq.GetFormat()), FeatureCount: feature_count}

	actual_layers := make(map[string]layer.InfoLayer)

	for _, layer_name := range freq.GetLayers() {
		layer := s.Layers[layer_name]
		if !layer.Queryable() {
			resp := NewRequestError(fmt.Sprintf("layer %s is not queryable", layer_name), "InvalidParameterValue", &WMS130ExceptionHandler{}, req, false, nil)
			return resp.Render()
		}
		for layer_name, map_layers := range layer.infoLayersForQuery(query) {
			actual_layers[layer_name] = map_layers
		}
	}

	var keys []string
	for k := range actual_layers {
		keys = append(keys, k)
	}

	authorized_layers, coverage := s.authorizedLayers("featureinfo", keys, &geo.MapExtent{BBox: freq.GetBBox(), Srs: geo.NewProj(freq.GetCrs())})
	s.filterActualInfoLayers(actual_layers, freq.GetLayers(), authorized_layers)

	if coverage != nil && !coverage.ContainsPoint(query.GetCoord(), query.Srs) {
		infos = nil
	} else {
		info_layers := []layer.InfoLayer{}
		for _, layer := range actual_layers {
			info_layers = append(info_layers, layer)
		}
		for _, layer := range info_layers {
			info := layer.GetInfo(query)
			if info == nil {
				continue
			}
			infos = append(infos, info)
		}
	}

	mimetype := freq.GetFormatString()

	if len(infos) == 0 {
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

func (s *WMSService) checkMapRequest(req request.Request) *RequestError {
	mapreq := req.(*request.WMSMapRequest)
	mapparams := request.NewWMSMapRequestParams(req.GetParams())
	si := mapparams.GetSize()
	if s.MaxOutputPixels != -1 && (si[0]*si[1]) > uint32(s.MaxOutputPixels) {
		return NewRequestError("image size too large", "", &WMS130ExceptionHandler{}, req, false, nil)
	}

	errr := s.validateLayers(req)
	if errr != nil {
		return errr
	}

	formats := []string{}
	for k := range s.ImageFormats {
		formats = append(formats, k)
	}

	err := mapreq.ValidateFormat(formats)
	if err != nil {
		return NewRequestError(err.Error(), "", &WMS130ExceptionHandler{}, req, false, nil)
	}

	srss := []string{}
	for _, s := range s.Srs.Srs {
		srss = append(srss, s.GetSrsCode())
	}

	err = mapreq.ValidateSrs(srss)
	if err != nil {
		return NewRequestError(err.Error(), "", &WMS130ExceptionHandler{}, req, false, nil)
	}

	return nil
}

func (s *WMSService) checkFeatureinfoRequest(req request.Request) *RequestError {
	mapreq := req.(*request.WMSMapRequest)
	errr := s.validateLayers(req)
	if errr != nil {
		return errr
	}

	srss := []string{}
	for _, s := range s.Srs.Srs {
		srss = append(srss, s.GetSrsCode())
	}

	err := mapreq.ValidateSrs(srss)
	if err != nil {
		return NewRequestError(err.Error(), "", &WMS130ExceptionHandler{}, req, false, nil)
	}
	return nil
}

func (s *WMSService) validateLayers(req request.Request) *RequestError {
	mapparams := request.NewWMSMapRequestParams(req.GetParams())
	query_layers := mapparams.GetLayers()
	for _, layer := range query_layers {
		if _, ok := s.Layers[layer]; !ok {
			return NewRequestError("unknown layer: "+layer, "", &WMS130ExceptionHandler{}, req, false, nil)
		}
	}
	return nil
}

func (s *WMSService) checkLegendRequest(req request.Request) *RequestError {
	mapparams := request.NewWMSLegendGraphicRequestParams(req.GetParams())
	layer := mapparams.GetLayer()
	if _, ok := s.Layers[layer]; !ok {
		return NewRequestError("unknown layer: "+layer, "", &WMS130ExceptionHandler{}, req, false, nil)
	}
	return nil
}

func (s *WMSService) Legendgraphic(req request.Request) *Response {
	mapparams := request.NewWMSLegendGraphicRequestParams(req.GetParams())
	l := mapparams.GetLayer()
	format := mapparams.GetFormatMimeType()
	scale := mapparams.GetScale()

	err := s.checkLegendRequest(req)
	if err != nil {
		return err.Render()
	}

	if !s.Layers[l].HasLegend() {
		resp := NewRequestError(fmt.Sprintf("layer %s has no legend graphic", l), "", &WMS130ExceptionHandler{}, req, false, nil)
		return resp.Render()
	}

	lquery := &layer.LegendQuery{Format: format, Scale: scale}

	legends := s.Layers[l].legend(lquery)

	result := imagery.ConcatLegends(legends, imagery.RGBA, tile.TileFormat("png"), nil, color.White, true)
	mimetype := mapparams.GetFormatMimeType()
	if mimetype == "" {
		mimetype = "image/png"
	}
	img_opts := s.ImageFormats[mimetype]
	return NewResponse(result.GetBuffer(nil, img_opts), 200, mimetype)
}

func (s *WMSService) serviceMetadata(tms_request request.Request, metadata *WMSMetadata) WMSMetadata {
	req := tms_request.(*request.BaseRequest)
	md := *metadata
	md.URL = req.Http.URL.Host
	if s.RootLayer.hasLegend {
		md.HasLegend = true
	} else {
		md.HasLegend = false
	}
	return md
}

func (s *WMSService) filterActualLayers(actual_layers map[string]layer.Layer, layers []string, authorized_layers []string) {

}

func (s *WMSService) filterActualInfoLayers(actual_layers map[string]layer.InfoLayer, layers []string, authorized_layers []string) {

}

func (s *WMSService) authorizedCapabilityLayers() *WMSGroupLayer {
	return s.RootLayer
}

type LayerRenderer struct {
	layers  []layer.Layer
	query   *layer.MapQuery
	params  request.WMSMapRequestParams
	request request.Request
}

func (l *LayerRenderer) render(layer_merger *imagery.LayerMerger) *RequestError {
	render_layers := CombinedLayers(l.layers, l.query)
	if render_layers == nil {
		return nil
	}
	for i := range render_layers {
		layer, layer_img, err := l.renderLayer(render_layers[i])
		if err != nil {
			return err
		}
		if layer_img != nil {
			layer_merger.AddSource(layer_img, layer.GetCoverage())
		}
	}
	return nil
}

func (l *LayerRenderer) renderLayer(layer layer.Layer) (layer.Layer, tile.Source, *RequestError) {
	layer_img, err := layer.GetMap(l.query)
	if err != nil {
		resp := NewRequestError("Invalid request.", "", &WMS130ExceptionHandler{}, l.request, false, nil)
		return nil, nil, resp
	}
	return layer, layer_img, nil
}

type WMSLayerMetadata struct {
	Abstract     string
	KeywordList  *wms130.Keywords
	AuthorityURL *wms130.AuthorityURL
	Identifier   *wms130.Identifier
	MetadataURL  []*wms130.MetadataURL
	Style        []*wms130.Style
}

type WMSLayer interface {
	layer.Layer
	rendersQuery(query *layer.MapQuery) bool
	mapLayersForQuery(query *layer.MapQuery) map[string]layer.Layer
	infoLayersForQuery(query *layer.InfoQuery) map[string]layer.InfoLayer
	legend(query *layer.LegendQuery) []tile.Source
	GetLegendSize() []int
	GetName() string
	GetTitle() string
	GetLegendURL() string
	HasLegend() bool
	Queryable() bool
	GetMetadata() *WMSLayerMetadata
	GetExtent() *geo.MapExtent
}

type WMSLayerBase struct {
	WMSLayer
	name       string
	title      string
	isActive   bool
	layers     map[string]WMSLayer
	metadata   *WMSLayerMetadata
	queryable  bool
	hasLegend  bool
	legendUrl  string
	legendSize []int
	resRange   *geo.ResolutionRange
	extent     *geo.MapExtent
}

func (l *WMSLayerBase) GetResolutionRange() *geo.ResolutionRange {
	return l.resRange
}

func (l *WMSLayerBase) Queryable() bool {
	return l.queryable
}

func (l *WMSLayerBase) HasLegend() bool {
	return l.hasLegend
}

func (l *WMSLayerBase) GetName() string {
	return l.name
}

func (l *WMSLayerBase) GetTitle() string {
	return l.title
}

func (l *WMSLayerBase) GetLegendURL() string {
	return l.legendUrl
}

func (l *WMSLayerBase) GetLegendSize() []int {
	return l.legendSize
}

func (l *WMSLayerBase) GetMetadata() *WMSLayerMetadata {
	return l.metadata
}

func (l *WMSLayerBase) GetExtent() *geo.MapExtent {
	return l.extent
}

type WMSNodeLayer struct {
	WMSLayerBase
	mapLayers    map[string]layer.Layer
	infoLayers   map[string]layer.InfoLayer
	legendLayers []layer.LegendLayer
}

func NewWMSNodeLayer(name string, title string, map_layers map[string]layer.Layer, infos map[string]layer.InfoLayer, legends []layer.LegendLayer, res_range *geo.ResolutionRange, md *WMSLayerMetadata) *WMSNodeLayer {
	queryable := false
	if len(infos) > 0 {
		queryable = true
	}

	has_legend := false
	if len(legends) > 0 {
		has_legend = true
	}

	ret := &WMSNodeLayer{WMSLayerBase: WMSLayerBase{name: name, title: title, metadata: md, isActive: false, layers: nil, hasLegend: has_legend, queryable: queryable}, mapLayers: map_layers, infoLayers: infos, legendLayers: legends}

	ret.extent = mergeLayerExtents(map_layers)
	if res_range == nil {
		ret.resRange = mergeLayerResRanges(map_layers)
	} else {
		ret.resRange = res_range
	}

	return ret
}

func (l *WMSNodeLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSNodeLayer) mapLayersForQuery(query *layer.MapQuery) map[string]layer.Layer {
	if l.mapLayers == nil {
		return nil
	}
	return l.mapLayers
}

func (l *WMSNodeLayer) infoLayersForQuery(query *layer.InfoQuery) map[string]layer.InfoLayer {
	if l.infoLayers == nil {
		return nil
	}
	return l.infoLayers
}

func (l *WMSNodeLayer) legend(query *layer.LegendQuery) []tile.Source {
	legend := []tile.Source{}

	for _, lyr := range l.legendLayers {
		legend = append(legend, lyr.GetLegend(query))
	}
	return legend
}

type WMSGroupLayer struct {
	WMSLayerBase
	this WMSLayer
}

func mergeLayerResRanges(layers map[string]layer.Layer) *geo.ResolutionRange {
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

func mergeLayerExtents(layers map[string]layer.Layer) *geo.MapExtent {
	if len(layers) == 0 {
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

func cloneLayers(tags map[string]WMSLayer) map[string]layer.Layer {
	cloneTags := make(map[string]layer.Layer)
	for k, v := range tags {
		cloneTags[k] = v
	}
	return cloneTags
}

func NewWMSGroupLayer(name string, title string, this WMSLayer, layers map[string]WMSLayer, md *WMSLayerMetadata) *WMSGroupLayer {
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

func (l *WMSGroupLayer) GetLegendSize() []int {
	return l.this.GetLegendSize()
}

func (l *WMSGroupLayer) GetLegendURL() string {
	return l.this.GetLegendURL()
}

func (l *WMSGroupLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSGroupLayer) mapLayersForQuery(query *layer.MapQuery) map[string]layer.Layer {
	if l.this != nil {
		return l.this.mapLayersForQuery(query)
	} else {
		layers := make(map[string]layer.Layer)
		for _, layer := range l.layers {
			ll := layer.mapLayersForQuery(query)
			for name, l := range ll {
				layers[name] = l
			}
		}
		return layers
	}
}

func (l *WMSGroupLayer) infoLayersForQuery(query *layer.InfoQuery) map[string]layer.InfoLayer {
	if l.this != nil {
		return l.this.infoLayersForQuery(query)
	} else {
		layers := make(map[string]layer.InfoLayer)
		for _, layer := range l.layers {
			ll := layer.infoLayersForQuery(query)
			for name, l := range ll {
				layers[name] = l
			}
		}
		return layers
	}
}

func (l *WMSGroupLayer) GetChildLayers() map[string]WMSLayer {
	layers := make(map[string]WMSLayer)
	if l.name != "" {
		layers[l.name] = l
	}
	for _, lyr := range l.layers {
		if iface, ok := lyr.(interface {
			GetChildLayers() map[string]WMSLayer
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

var (
	WMS130ExceptionCodes = map[string]string{
		"InvalidFormat":         "Request contains a Format not offered by the server.",
		"InvalidCRS":            "Request contains a CRS not offered by the server for one or more of the Layers in the request.",
		"LayerNotDefined":       "GetMap request is for a Layer not offered by the server, or GetFeatureInfo request is for a Layer not shown on the map.",
		"StyleNotDefined":       "Request is for a Layer in a Style not offered by the server.",
		"LayerNotQueryable":     "GetFeatureInfo request is applied to a Layer which is not declared queryable.",
		"InvalidPoint":          "GetFeatureInfo request contains invalid I or J value.",
		"CurrentUpdateSequence": "Value of (optional) UpdateSequence parameter in GetCapabilities request is equal to current value of service metadata update sequence number.",
		"InvalidUpdateSequence": "Value of (optional) UpdateSequence parameter in GetCapabilities request is greater than current value of service metadata update sequence number.",
		"MissingDimensionValue": "Request does not include a sample dimension value, and the server did not declare a default value for that dimension.",
		"InvalidDimensionValue": "Request contains an invalid sample dimension value. OperationNotSupported Request is for an optional operation that is not supported by the server.",
	}
)

type WMS130ExceptionHandler struct {
	ExceptionHandler
}

func (h *WMS130ExceptionHandler) Render(err *RequestError) *Response {
	exp := wms130.Exceptions(wms130.NewExceptions(err.Message, err.Code))
	report := exp.ToReport()
	return NewResponse(report.ToBytes(), 400, "text/xml")
}

type WMSImageExceptionHandler struct {
	ExceptionHandler
}

func (h *WMSImageExceptionHandler) Render(request_error *RequestError) *Response {
	req := request_error.Request
	params := req.GetParams()
	mapreq := request.NewWMSMapRequestParams(params)
	format := mapreq.GetFormat()
	size := mapreq.GetSize()
	if size == [2]uint32{0, 0} {
		size = [2]uint32{256, 256}
	}
	transparent := mapreq.GetTransparent()
	bgcolor := mapreq.GetBGColor()
	image_opts := &imagery.ImageOptions{Format: format, BgColor: bgcolor, Transparent: &transparent}
	result := imagery.GenMessageImage(request_error.Message, size, image_opts)
	return NewResponse(result.GetBuffer(nil, nil), 200, format.MimeType())
}

type WMSBlankExceptionHandler struct {
	WMSImageExceptionHandler
}

func (h *WMSBlankExceptionHandler) Render(request_error *RequestError) *Response {
	request_error.Message = ""
	return h.WMSImageExceptionHandler.Render(request_error)
}

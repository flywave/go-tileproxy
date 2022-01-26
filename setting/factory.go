package setting

import (
	"image/color"
	"path"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/ogc-osgeo/pkg/wms130"
	"github.com/flywave/ogc-osgeo/pkg/wsc110"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	"github.com/flywave/go-tileproxy/vector"
)

type CacheFactory interface {
	CreateCache(cache *CacheInfo, opts tile.TileOptions) cache.Cache
	CreateStore(cache *StoreInfo) resource.Store
}

func GetPreferredSrcSRS(srs *Srs) geo.PreferredSrcSRS {
	ret := make(geo.PreferredSrcSRS)
	for k, ss := range srs.PreferredSrcProj {
		for i := range ss {
			ret[k] = append(ret[k], geo.NewProj(ss[i]))
		}
	}
	return ret
}

func NewImageOptions(opt *ImageOpts) *imagery.ImageOptions {
	image_opt := &imagery.ImageOptions{}
	image_opt.Transparent = opt.Transparent
	image_opt.Format = tile.TileFormat(opt.Format)
	if opt.ResamplingMethod != "" {
		image_opt.Resampling = opt.ResamplingMethod
	} else {
		image_opt.Resampling = DefaultResamplingMethod
	}
	if opt.Mode != "" {
		image_opt.Mode = imagery.ImageModeFromString(opt.Mode)
	}
	image_opt.EncodingOptions = opt.EncodingOptions
	return image_opt
}

func NewRasterOptions(opt *RasterOpts) *terrain.RasterOptions {
	raster_opt := &terrain.RasterOptions{}
	raster_opt.Format = tile.TileFormat(opt.Format)
	raster_opt.MaxError = opt.MaxError
	if opt.Nodata != nil {
		raster_opt.Nodata = *opt.Nodata
	}
	if opt.Interpolator != nil {
		raster_opt.Interpolator = *opt.Interpolator
	}
	if opt.Mode != nil {
		raster_opt.Mode = terrain.BorderModeFromString(*opt.Mode)
	}
	if opt.DataType != nil {
		raster_opt.DataType = terrain.RasterTypeFromString(*opt.DataType)
	}
	return raster_opt
}

func NewVectorOptions(opt *VectorOpts) *vector.VectorOptions {
	vector_opt := &vector.VectorOptions{}
	if opt.Buffer != nil {
		vector_opt.Buffer = *opt.Buffer
	}
	vector_opt.Format = tile.TileFormat(opt.Format)
	vector_opt.Extent = opt.Extent
	if opt.LineMetrics != nil {
		vector_opt.LineMetrics = *opt.LineMetrics
	}
	if opt.MaxZoom != nil {
		vector_opt.MaxZoom = *opt.MaxZoom
	}
	if opt.Tolerance != nil {
		vector_opt.Tolerance = *opt.Tolerance
	}
	if opt.Proto != nil {
		vector_opt.Proto = *opt.Proto
	}
	return vector_opt
}

func ConvertGridOpts(opt *GridOpts) *geo.TileGrid {
	conf := make(geo.TileGridOptions)
	conf[geo.TILEGRID_NAME] = opt.Name
	conf[geo.TILEGRID_SRS] = opt.Srs
	if opt.BBox != nil {
		rect := &vec2d.Rect{Min: vec2d.T{opt.BBox[0], opt.BBox[1]}, Max: vec2d.T{opt.BBox[2], opt.BBox[3]}}
		conf[geo.TILEGRID_BBOX] = rect
	}
	if opt.NumLevels != nil {
		conf[geo.TILEGRID_NUM_LEVELS] = *opt.NumLevels
	}
	if len(opt.Resolutions) > 0 {
		conf[geo.TILEGRID_RES] = opt.Resolutions
	}
	if opt.ResFactor != nil {
		conf[geo.TILEGRID_RES_FACTOR] = opt.ResFactor
	}
	if opt.MaxRes != nil {
		conf[geo.TILEGRID_MAX_RES] = *opt.MaxRes
	}
	if opt.MinRes != nil {
		conf[geo.TILEGRID_MIN_RES] = *opt.MinRes
	}
	if opt.MaxStretchFactor != nil {
		conf[geo.TILEGRID_MAX_STRETCH_FACTOR] = *opt.MaxStretchFactor
	} else {
		conf[geo.TILEGRID_MAX_STRETCH_FACTOR] = DefaultStretchFactor
	}
	if opt.MaxShrinkFactor != nil {
		conf[geo.TILEGRID_MAX_SHRINK_FACTOR] = *opt.MaxShrinkFactor
	} else {
		conf[geo.TILEGRID_MAX_SHRINK_FACTOR] = DefaultMaxShrinkFactor
	}
	if opt.AlignResolutionsWith != "" {
		conf[geo.TILEGRID_ALIGN_WITH] = opt.AlignResolutionsWith
	}
	if opt.Origin != "" {
		conf[geo.TILEGRID_ORIGIN] = geo.OriginFromString(opt.Origin)
	}
	if opt.TileSize != nil {
		conf[geo.TILEGRID_TILE_SIZE] = (*opt.TileSize)[:]
	} else {
		conf[geo.TILEGRID_TILE_SIZE] = DefaultTileSize[:]
	}
	if opt.InitialResMin != nil {
		conf[geo.TILEGRID_INITIAL_RES_MIN] = *opt.InitialResMin
	}
	return geo.NewTileGrid(conf)
}

func newWaterMarkFilter(w *WaterMark) cache.Filter {
	var c color.Color
	if w.Color != nil {
		c = color.NRGBA{R: w.Color[0], G: w.Color[1], B: w.Color[2], A: w.Color[3]}
	}
	return cache.NewWatermark(w.Text, w.Opacity, w.Spacing, w.FontSize, &c)
}

func LoadFilter(f interface{}) cache.Filter {
	switch filter := f.(type) {
	case *WaterMark:
		return newWaterMarkFilter(filter)
	}
	return nil
}

func PreLoadCacheManager(c *CacheSource, globals *GlobalsSetting, instance ProxyInstance, fac CacheFactory) cache.Manager {
	grid := instance.GetGrid(c.Grid)
	request_format := c.RequestFormat
	var request_format_ext string
	if strings.Contains(request_format, "/") {
		request_format_ext = strings.Split(request_format, "/")[1]
	} else {
		request_format_ext = request_format
	}

	var meta_buffer int
	if c.MetaBuffer != nil {
		meta_buffer = *c.MetaBuffer
	} else {
		meta_buffer = -1
	}

	var meta_size [2]uint32
	if c.MetaSize != nil {
		meta_size = [2]uint32{c.MetaSize[0], c.MetaSize[1]}
	} else {
		meta_size = [2]uint32{1, 1}
	}

	var bulk_meta_tiles bool
	if c.BulkMetaTiles != nil {
		bulk_meta_tiles = *c.BulkMetaTiles
	} else {
		bulk_meta_tiles = globals.Cache.BulkMetaTiles
	}

	var minimize_meta_requests bool
	if c.MinimizeMetaRequests != nil {
		minimize_meta_requests = *c.MinimizeMetaRequests
	} else {
		minimize_meta_requests = globals.Cache.MinimizeMetaRequests
	}

	var cache_rescaled_tiles bool

	if c.CacheRescaledTiles != nil {
		cache_rescaled_tiles = *c.CacheRescaledTiles
	} else {
		cache_rescaled_tiles = false
	}

	upscale_tiles := c.UpscaleTiles
	if upscale_tiles != nil && *upscale_tiles < 0 {
		return nil
	}

	downscale_tiles := c.DownscaleTiles
	if downscale_tiles != nil && *downscale_tiles < 0 {
		return nil
	}

	if upscale_tiles != nil && downscale_tiles != nil {
		return nil
	}

	rescale_tiles := 0
	if upscale_tiles != nil {
		rescale_tiles = -*upscale_tiles
	}
	if downscale_tiles != nil {
		rescale_tiles = *downscale_tiles
	}

	name := c.Name

	var locker cache.TileLocker

	if c.LockDir != "" {
		locker = cache.NewFileTileLocker(c.LockDir, time.Duration(c.LockRetryDelay*int(time.Second)))
	} else {
		locker = &cache.DummyTileLocker{}
	}

	var pre_store_filter []cache.Filter

	if c.Filters != nil {
		for i := range c.Filters {
			pre_store_filter = append(pre_store_filter, LoadFilter(c.Filters[i]))
		}
	}

	tilegrid := grid.(*geo.TileGrid)

	var opts tile.TileOptions
	switch o := c.TileOptions.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}

	var cacheB cache.Cache

	if fac != nil {
		cacheB = fac.CreateCache(c.CacheInfo, opts)
	} else {
		cacheB = ConvertLocalCache(c.CacheInfo, opts)
	}

	var reprojectSrcSrs geo.Proj
	if c.ReprojectSrcSrs != nil {
		reprojectSrcSrs = geo.NewProj(*c.ReprojectSrcSrs)
	}

	var reprojectDstSrs geo.Proj
	if c.ReprojectDstSrs != nil {
		reprojectDstSrs = geo.NewProj(*c.ReprojectDstSrs)
	}

	topts := &cache.TileManagerOptions{
		Sources:              nil,
		Grid:                 tilegrid,
		Cache:                cacheB,
		Locker:               locker,
		Identifier:           name,
		Format:               request_format_ext,
		Options:              opts,
		MinimizeMetaRequests: minimize_meta_requests,
		BulkMetaTiles:        bulk_meta_tiles,
		PreStoreFilter:       pre_store_filter,
		RescaleTiles:         rescale_tiles,
		CacheRescaledTiles:   cache_rescaled_tiles,
		MetaBuffer:           meta_buffer,
		MetaSize:             meta_size,
		ReprojectSrcSrs:      reprojectSrcSrs,
		ReprojectDstSrs:      reprojectDstSrs,
	}

	return cache.NewTileManager(topts)
}

func LoadCacheManager(c *CacheSource, globals *GlobalsSetting, instance ProxyInstance, fac CacheFactory, manager cache.Manager) {
	layers := []layer.Layer{}

	for i := range c.Sources {
		l := instance.GetSource(c.Sources[i])
		if l != nil {
			layers = append(layers, l)
		} else {
			l := instance.GetCacheSource(c.Sources[i])
			if l != nil {
				layers = append(layers, l)
			}
		}
	}

	manager.SetSources(layers)
}

func NewResolutionRange(conf *ScaleHints) *geo.ResolutionRange {
	if conf.MinRes != nil || conf.MaxRes != nil {
		return &geo.ResolutionRange{Min: conf.MinRes, Max: conf.MaxRes}
	}
	if conf.MinScale != nil || conf.MaxScale != nil {
		return geo.NewResolutionRangeScale(conf.MinScale, conf.MaxScale)
	}
	return nil
}

func ConvertLocalCache(opt *CacheInfo, opts tile.TileOptions) *cache.LocalCache {
	return cache.NewLocalCache(opt.Directory, opt.DirectoryLayout, cache.GetSourceCreater(opts))
}

func ConvertLocalStore(opt *StoreInfo) *resource.LocalStore {
	return resource.NewLocalStore(opt.Directory)
}

func ConvertMapboxTileLayer(l *MapboxTileLayer, globals *GlobalsSetting, instance ProxyInstance) *service.MapboxTileProvider {
	tp := service.MapboxVector
	if l.TileType != "" {
		tp = service.GetMapboxTileType(l.TileType)
	}
	tileManager := instance.GetCache(l.Source)

	tilesource := instance.GetSource(l.TileJSON)

	tilejsonSource, ok := tilesource.(layer.MapboxTileJSONLayer)

	metadata := &service.MapboxLayerMetadata{
		Name:        l.Name,
		Attribution: l.Attribution,
		Description: l.Description,
		Legend:      l.Legend,
		FillZoom:    l.FillZoom,
	}

	topts := &service.MapboxTileOptions{
		Name:           l.Name,
		Type:           tp,
		Metadata:       metadata,
		TileManager:    tileManager,
		TilejsonSource: nil,
		VectorLayers:   l.VectorLayers,
		ZoomRange:      l.ZoomRange,
	}

	if ok {
		topts.TilejsonSource = tilejsonSource
		return service.NewMapboxTileProvider(topts)
	}

	return service.NewMapboxTileProvider(topts)
}

func ConvertTileLayer(l *TileLayer, instance ProxyInstance) *service.TileProvider {
	dimensions := utils.NewDimensionsFromValues(l.Dimensions)

	tileManager := instance.GetCache(l.TileSource)

	infoSources := []layer.InfoLayer{}
	for _, info := range l.InfoSources {
		infoSources = append(infoSources, instance.GetInfoSource(info))
	}

	metadata := &service.TileProviderMetadata{Name: l.Name, Title: l.Title}

	tpopts := &service.TileProviderOptions{
		Name:         l.Name,
		Title:        l.Title,
		Metadata:     metadata,
		TileManager:  tileManager,
		InfoSources:  infoSources,
		Dimensions:   dimensions,
		ErrorHandler: &service.TMSExceptionHandler{},
	}

	return service.NewTileProvider(tpopts)
}

func ConvertWMSLayerMetadata(metadata *WMSLayerMetadata) *service.WMSLayerMetadata {
	if metadata == nil {
		return nil
	}
	ret := &service.WMSLayerMetadata{}
	ret.Abstract = metadata.Abstract

	if metadata.KeywordList != nil {
		ret.KeywordList = &wms130.Keywords{}
		copy(ret.KeywordList.Keyword, metadata.KeywordList.Keyword)
	}

	if metadata.AuthorityURL != nil {
		ret.AuthorityURL = &wms130.AuthorityURL{}
		ret.AuthorityURL.Name = metadata.AuthorityURL.Name
		ret.AuthorityURL.OnlineResource.Href = metadata.AuthorityURL.OnlineResource.Href
		ret.AuthorityURL.OnlineResource.Type = metadata.AuthorityURL.OnlineResource.Type
		ret.AuthorityURL.OnlineResource.Xlink = metadata.AuthorityURL.OnlineResource.Xlink
	}

	if metadata.Identifier != nil {
		ret.Identifier = &wms130.Identifier{}
		ret.Identifier.Authority = metadata.Identifier.Authority
		ret.Identifier.Value = metadata.Identifier.Value
	}

	for i := range metadata.MetadataURL {
		url := &wms130.MetadataURL{}
		url.Type = metadata.MetadataURL[i].Type
		url.Format = metadata.MetadataURL[i].Format

		url.Format = metadata.MetadataURL[i].Format
		url.Format = metadata.MetadataURL[i].Format

		url.OnlineResource.Href = metadata.MetadataURL[i].OnlineResource.Href
		url.OnlineResource.Type = metadata.MetadataURL[i].OnlineResource.Type
		url.OnlineResource.Xlink = metadata.MetadataURL[i].OnlineResource.Xlink

		ret.MetadataURL = append(ret.MetadataURL, url)
	}

	for i := range metadata.Style {
		stl := &wms130.Style{}
		stl.Name = metadata.Style[i].Name
		stl.Title = metadata.Style[i].Title
		stl.Abstract = metadata.Style[i].Abstract

		stl.LegendURL.Width = metadata.Style[i].LegendURL.Width
		stl.LegendURL.Height = metadata.Style[i].LegendURL.Height
		stl.LegendURL.Format = metadata.Style[i].LegendURL.Format

		stl.LegendURL.OnlineResource.Href = metadata.Style[i].LegendURL.OnlineResource.Href
		stl.LegendURL.OnlineResource.Type = metadata.Style[i].LegendURL.OnlineResource.Type
		stl.LegendURL.OnlineResource.Xlink = metadata.Style[i].LegendURL.OnlineResource.Xlink

		if metadata.Style[i].StyleSheetURL != nil {
			stl.StyleSheetURL = new(struct {
				Format         string                `xml:"Format" yaml:"format"`
				OnlineResource wms130.OnlineResource `xml:"OnlineResource" yaml:"onlineresource"`
			})
			stl.StyleSheetURL.Format = metadata.Style[i].StyleSheetURL.Format

			stl.StyleSheetURL.OnlineResource.Href = metadata.Style[i].StyleSheetURL.OnlineResource.Href
			stl.StyleSheetURL.OnlineResource.Type = metadata.Style[i].StyleSheetURL.OnlineResource.Type
			stl.StyleSheetURL.OnlineResource.Xlink = metadata.Style[i].StyleSheetURL.OnlineResource.Xlink
		}

		ret.Style = append(ret.Style, stl)
	}

	return ret
}

func ConvertWMSLayer(l *WMSLayer, instance ProxyInstance) service.WMSLayer {
	_range := &geo.ResolutionRange{Min: l.MinRes, Max: l.MaxRes}
	mapLayers := make(map[string]layer.Layer)
	infos := make(map[string]layer.InfoLayer)
	legends := make([]layer.LegendLayer, 0)

	for _, name := range l.MapSources {
		s := instance.GetSource(name)
		if s != nil {
			mapLayers[name] = s
		} else {
			s = instance.GetCacheSource(name)
			if s != nil {
				mapLayers[name] = s
			}
		}
	}

	for _, name := range l.FeatureinfoSources {
		infos[name] = instance.GetInfoSource(name)
	}

	for _, name := range l.LegendSources {
		legends = append(legends, instance.GetLegendSource(name))
	}

	nopts := &service.WMSNodeLayerOptions{
		Name:      l.Name,
		Title:     l.Title,
		MapLayers: mapLayers,
		Infos:     infos,
		Legends:   legends,
		ResRange:  _range,
		Metadata:  ConvertWMSLayerMetadata(l.Metadata),
	}

	return service.NewWMSNodeLayer(nopts)
}

func loadWMSRootLayer(l *WMSLayer, instance ProxyInstance) *service.WMSGroupLayer {
	thisLayer := ConvertWMSLayer(l, instance)
	layers := make(map[string]service.WMSLayer)
	for i := range l.Layers {
		layers[l.Layers[i].Name] = ConvertWMSLayer(&l.Layers[i], instance)
	}

	nopts := &service.WMSGroupLayerOptions{
		Name:     l.Name,
		Title:    l.Title,
		This:     thisLayer,
		Layers:   layers,
		Metadata: ConvertWMSLayerMetadata(l.Metadata),
	}

	return service.NewWMSGroupLayer(nopts)
}

func newSupportedSrs(supportedSrs []string, preferred geo.PreferredSrcSRS) *geo.SupportedSRS {
	srs := []geo.Proj{}
	for i := range supportedSrs {
		srs = append(srs, geo.NewProj(supportedSrs[i]))
	}
	return &geo.SupportedSRS{Srs: srs, Preferred: preferred}
}

func LoadWMSInfoSource(s *WMSSource, basePath string, globals *GlobalsSetting) *sources.WMSInfoSource {
	if s.Opts.FeatureInfo == nil || !*s.Opts.FeatureInfo {
		return nil
	}
	params := make(request.RequestParams)

	request_format := s.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	if s.Opts.FeatureinfoFormat != "" {
		params["info_format"] = []string{s.Opts.FeatureinfoFormat}
	}

	url := s.Url
	coverage := LoadCoverage(s.Coverage)

	var transformer *resource.XSLTransformer

	if s.Opts.FeatureinfoXslt != "" {
		fi_format := s.Opts.FeatureinfoOutFormat
		if fi_format == "" {
			fi_format = s.Opts.FeatureinfoFormat
		}

		transformer = resource.NewXSLTransformer(path.Join(basePath, s.Opts.FeatureinfoXslt), &fi_format)
	}

	fi_request := request.NewWMSFeatureInfoRequest(params, url, false, nil, false)

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}

	c := client.NewWMSInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, GetPreferredSrcSRS(&globals.Srs)), s.AccessToken, s.AccessTokenName, newCollectorContext(http))

	if s.Opts.Version == "1.1.1" {
		c.AdaptTo111 = true
	}

	return sources.NewWMSInfoSource(c, coverage, transformer)
}

func LoadWMSLegendsSource(s *WMSSource, globals *GlobalsSetting, fac CacheFactory) *sources.WMSLegendSource {
	if s.Opts.LegendGraphic == nil || !*s.Opts.LegendGraphic {
		return nil
	}
	params := make(request.RequestParams)

	request_format := s.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}
	if s.Transparent != nil && *s.Transparent {
		params["transparent"] = []string{"true"}
	}

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}

	url := s.Url
	lg_clients := make([]*client.WMSLegendClient, 0)
	for _, layer := range s.Layers {
		params["layer"] = []string{layer}
		lg_request := request.NewWMSLegendGraphicRequest(params, url, false, nil, false)
		lg_clients = append(lg_clients, client.NewWMSLegendClient(lg_request, s.AccessToken, s.AccessTokenName, newCollectorContext(http)))
	}

	var cache *resource.LegendCache
	if fac != nil {
		cache = resource.NewLegendCache(fac.CreateStore(s.Store))
	} else {
		cache = resource.NewLegendCache(ConvertLocalStore(s.Store))
	}

	return sources.NewWMSLegendSource(s.Opts.LegendID, lg_clients, cache)
}

func LoadWMSMapSource(s *WMSSource, instance ProxyInstance, globals *GlobalsSetting) *sources.WMSSource {
	if s.Opts.Map == nil || !*s.Opts.Map {
		return nil
	}

	params := make(request.RequestParams)

	request_format := s.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	image_opts := NewImageOptions(&s.Image.ImageOpts)
	res_range := NewResolutionRange(&s.ScaleHints)
	supported_srs := newSupportedSrs(s.SupportedSrs, GetPreferredSrcSRS(&globals.Srs))

	var coverage geo.Coverage
	if s.Coverage != nil {
		coverage = LoadCoverage(s.Coverage)
	}

	var transparent_color color.Color
	transparent_color_tolerance := s.Image.TransparentColorTolerance

	if s.Image.TransparentColor != nil {
		c := s.Image.TransparentColor
		transparent_color = color.RGBA{R: c[0], G: c[1], B: c[2], A: c[3]}
	}

	url := s.Url

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}

	req := request.NewWMSMapRequest(params, url, false, nil, false)
	c := client.NewWMSClient(req, s.AccessToken, s.AccessTokenName, newCollectorContext(http))

	if s.Opts.Version == "1.1.1" {
		c.AdaptTo111 = true
	}

	return sources.NewWMSSource(c, image_opts, coverage,
		res_range, transparent_color,
		transparent_color_tolerance, supported_srs, s.SupportedFormats,
		s.ForwardReqParams)
}

func newCollectorContext(httpOpts *HttpSetting) *client.CollectorContext {
	conf := client.Config{}
	if httpOpts.UserAgent != nil {
		conf.UserAgent = *httpOpts.UserAgent
	} else {
		conf.UserAgent = DefaultUserAgent
	}
	if httpOpts.RandomDelay != nil {
		conf.RandomDelay = *httpOpts.RandomDelay
	} else {
		conf.RandomDelay = DefaultRandomDelay
	}
	if httpOpts.DisableKeepAlives != nil {
		conf.DisableKeepAlives = *httpOpts.DisableKeepAlives
	}
	if httpOpts.Proxys != nil {
		conf.Proxys = httpOpts.Proxys
	}
	if httpOpts.RequestTimeout != nil {
		conf.RequestTimeout = *httpOpts.RequestTimeout
	} else {
		conf.RequestTimeout = time.Duration(DefaultRequestTimeout * int(time.Second))
	}
	if httpOpts.Threads != nil {
		conf.Threads = *httpOpts.Threads
	} else {
		conf.Threads = DefaultThreads
	}
	return client.NewCollectorContext(&conf)
}

func LoadTileSource(s *TileSource, globals *GlobalsSetting, instance ProxyInstance) *sources.TileSource {
	var opts tile.TileOptions
	switch o := s.Options.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}
	grid := instance.GetGrid(s.Grid)
	var coverage geo.Coverage
	if s.Coverage != nil {
		coverage = LoadCoverage(s.Coverage)
	}
	res_range := NewResolutionRange(&s.ScaleHints)

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}

	creater := cache.GetSourceCreater(opts)

	tpl := client.NewURLTemplate(s.URLTemplate, s.RequestFormat, s.Subdomains)

	c := client.NewTileClient(grid.(*geo.TileGrid), tpl, s.AccessToken, newCollectorContext(http))

	return sources.NewTileSource(grid.(*geo.TileGrid), c, coverage, opts, res_range, creater)
}

func LoadMapboxTileSource(s *MapboxTileSource, globals *GlobalsSetting, instance ProxyInstance, fac CacheFactory) *sources.MapboxTileSource {
	var opts tile.TileOptions
	switch o := s.Options.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}

	grid := instance.GetGrid(s.Grid)

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}

	var tcache *resource.TileJSONCache
	if fac != nil {
		tcache = resource.NewTileJSONCache(fac.CreateStore(s.TilejsonStore))
	} else {
		tcache = resource.NewTileJSONCache(ConvertLocalStore(s.TilejsonStore))
	}

	creater := cache.GetSourceCreater(opts)
	accessTokenName := "access_token"
	if s.AccessTokenName != "" {
		accessTokenName = s.AccessTokenName
	}
	c := client.NewMapboxTileClient(s.Url, s.TilejsonUrl, s.AccessToken, accessTokenName, newCollectorContext(http))

	return sources.NewMapboxTileSource(grid.(*geo.TileGrid), c, opts, creater, tcache)
}

func LoadArcGISSource(s *ArcGISSource, instance ProxyInstance, globals *GlobalsSetting) *sources.ArcGISSource {
	params := make(request.RequestParams)

	request_format := s.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	if s.Layers != nil {
		params["layers"] = s.Layers
	}

	if s.Transparent != nil {
		if *s.Transparent {
			params["transparent"] = []string{"true"}
		} else {
			params["transparent"] = []string{"false"}
		}
	}

	if s.PixelType != nil {
		params["pixelType"] = []string{*s.PixelType}
	} else {
		params["pixelType"] = []string{"UNKNOWN"}
	}

	if s.Dpi != nil {
		params["dpi"] = []string{strconv.Itoa(*s.Dpi)}
	}

	if s.LercVersion != nil {
		params["lercVersion"] = []string{strconv.Itoa(*s.LercVersion)}
	}

	if s.CompressionQuality != nil {
		params["compressionQuality"] = []string{strconv.Itoa(*s.CompressionQuality)}
	}

	image_opts := NewImageOptions(&s.Image.ImageOpts)
	res_range := NewResolutionRange(&s.ScaleHints)
	supported_srs := newSupportedSrs(s.SupportedSrs, GetPreferredSrcSRS(&globals.Srs))

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}
	url := s.Url

	coverage := LoadCoverage(s.Coverage)

	req := request.NewArcGISRequest(params, url)
	c := client.NewArcGISClient(req, newCollectorContext(http))

	return sources.NewArcGISSource(c, image_opts, coverage, res_range, supported_srs, s.SupportedFormats)
}

func LoadArcGISInfoSource(s *ArcGISSource, globals *GlobalsSetting) *sources.ArcGISInfoSource {
	params := make(request.RequestParams)

	request_format := s.Format
	if request_format != "" {
		params["format"] = []string{request_format}
	}

	if s.Layers != nil {
		params["layers"] = s.Layers
	}

	if s.Transparent != nil {
		if *s.Transparent {
			params["transparent"] = []string{"true"}
		} else {
			params["transparent"] = []string{"false"}
		}
	}

	if s.Dpi != nil {
		params["dpi"] = []string{strconv.Itoa(*s.Dpi)}
	}

	var tolerance int
	if s.Opts.FeatureinfoTolerance != nil {
		tolerance = *s.Opts.FeatureinfoTolerance
	} else {
		tolerance = 5
	}

	var return_geometries bool
	if s.Opts.FeatureinfoReturnGeometries != nil {
		return_geometries = *s.Opts.FeatureinfoReturnGeometries
	} else {
		return_geometries = false
	}

	var http *HttpSetting
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpSetting
	}
	url := s.Url

	fi_request := request.NewArcGISIdentifyRequest(params, url)

	c := client.NewArcGISInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, GetPreferredSrcSRS(&globals.Srs)), newCollectorContext(http), return_geometries, tolerance, s.Opts.AccessToken, s.Opts.AccessTokenName)
	return sources.NewArcGISInfoSource(c)
}

func LoadMapboxService(s *MapboxService, globals *GlobalsSetting, instance ProxyInstance, fac CacheFactory) *service.MapboxService {
	layers := make(map[string]service.Provider)
	metadata := &service.MapboxMetadata{Name: s.Name}

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertMapboxTileLayer(&tl, globals, instance)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}

	sopts := &service.MapboxServiceOptions{Tilesets: layers, Metadata: metadata, MaxTileAge: maxTileAge}

	return service.NewMapboxService(sopts)
}

func LoadTMSService(s *TMSService, instance ProxyInstance) *service.TileService {
	layers := make(map[string]service.Provider)
	metadata := &service.TileMetadata{Title: s.Title, Abstract: s.Abstract}

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}
	origin := s.Origin

	tsopts := &service.TileServiceOptions{Layers: layers, Metadata: metadata, MaxTileAge: maxTileAge, UseDimensionLayers: false, Origin: origin}

	return service.NewTileService(tsopts)
}

func ConvertWMTSServiceProvider(provider *WMTSServiceProvider) *wsc110.ServiceProvider {
	if provider == nil {
		return nil
	}
	ret := &wsc110.ServiceProvider{}
	ret.ProviderName = provider.ProviderName
	ret.ProviderSite.Type = provider.ProviderSite.Type
	ret.ProviderSite.Href = provider.ProviderSite.Href

	ret.ServiceContact.IndividualName = provider.ServiceContact.IndividualName
	ret.ServiceContact.PositionName = provider.ServiceContact.PositionName
	ret.ServiceContact.ContactInfo.Phone.Voice = provider.ServiceContact.ContactInfo.Phone.Voice
	ret.ServiceContact.ContactInfo.Phone.Facsimile = provider.ServiceContact.ContactInfo.Phone.Facsimile

	ret.ServiceContact.ContactInfo.Address.DeliveryPoint = provider.ServiceContact.ContactInfo.Address.DeliveryPoint
	ret.ServiceContact.ContactInfo.Address.City = provider.ServiceContact.ContactInfo.Address.City
	ret.ServiceContact.ContactInfo.Address.AdministrativeArea = provider.ServiceContact.ContactInfo.Address.AdministrativeArea
	ret.ServiceContact.ContactInfo.Address.PostalCode = provider.ServiceContact.ContactInfo.Address.PostalCode
	ret.ServiceContact.ContactInfo.Address.Country = provider.ServiceContact.ContactInfo.Address.Country
	ret.ServiceContact.ContactInfo.Address.ElectronicMailAddress = provider.ServiceContact.ContactInfo.Address.ElectronicMailAddress

	ret.ServiceContact.ContactInfo.OnlineResource.Type = provider.ServiceContact.ContactInfo.OnlineResource.Type
	ret.ServiceContact.ContactInfo.OnlineResource.Href = provider.ServiceContact.ContactInfo.OnlineResource.Href

	ret.ServiceContact.ContactInfo.HoursOfService = provider.ServiceContact.ContactInfo.HoursOfService
	ret.ServiceContact.ContactInfo.ContactInstructions = provider.ServiceContact.ContactInfo.ContactInstructions

	ret.ServiceContact.Role = provider.ServiceContact.Role

	return ret
}

func LoadWMTSService(s *WMTSService, instance ProxyInstance) *service.WMTSService {
	layers := make(map[string]service.Provider)
	metadata := &service.WMTSMetadata{
		Title:             s.Title,
		Abstract:          s.Abstract,
		KeywordList:       s.KeywordList,
		Fees:              s.Fees,
		AccessConstraints: s.AccessConstraints,
		Provider:          ConvertWMTSServiceProvider(s.Provider),
	}

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}

	info_formats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		info_formats[info.Suffix] = info.MimeType
	}

	wopts := &service.WMTSServiceOptions{Layers: layers, Metadata: metadata, MaxTileAge: maxTileAge, InfoFormats: info_formats}

	return service.NewWMTSService(wopts)
}

func LoadWMTSRestfulService(s *WMTSService, instance ProxyInstance) *service.WMTSRestService {
	if s.Restful == nil || !*s.Restful {
		return nil
	}

	layers := make(map[string]service.Provider)

	metadata := &service.WMTSMetadata{
		Title:             s.Title,
		Abstract:          s.Abstract,
		KeywordList:       s.KeywordList,
		Fees:              s.Fees,
		AccessConstraints: s.AccessConstraints,
		Provider:          ConvertWMTSServiceProvider(s.Provider),
	}

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}

	info_formats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		info_formats[info.Suffix] = info.MimeType
	}

	wopts := &service.WMTSRestServiceOptions{
		Layers:                     layers,
		Metadata:                   metadata,
		MaxTileAge:                 maxTileAge,
		RestfulTemplate:            s.RestfulTemplate,
		RestfulFeatureinfoTemplate: s.RestfulFeatureinfoTemplate,
		InfoFormats:                info_formats,
	}

	return service.NewWMTSRestService(wopts)
}

func loadXSLTransformer(featureinfoXslt map[string]string, basePath string) map[string]*resource.XSLTransformer {
	fi_transformers := make(map[string]*resource.XSLTransformer)
	for info_type, fi_xslt := range featureinfoXslt {
		fi_transformers[info_type] = resource.NewXSLTransformer(path.Join(basePath, fi_xslt), nil)
	}
	return fi_transformers
}

func extentsForSrs(bbox_srs []BBoxSrs) map[string]*geo.MapExtent {
	extents := make(map[string]*geo.MapExtent)
	for _, srs := range bbox_srs {
		srs, bbox := srs.Srs, srs.BBox
		e := &geo.MapExtent{BBox: vec2d.Rect{Min: vec2d.T{bbox[0], bbox[1]}, Max: vec2d.T{bbox[2], bbox[3]}}, Srs: geo.NewProj(srs)}
		extents[srs] = e
	}
	return extents
}

func LoadWMSService(s *WMSService, instance ProxyInstance, globals *GlobalsSetting, basePath string) *service.WMSService {
	md := &service.WMSMetadata{
		Title:       s.Title,
		Abstract:    s.Abstract,
		KeywordList: s.KeywordList,
		OnlineResource: struct {
			Xlink *string
			Type  *string
			Href  *string
		}{
			Xlink: s.OnlineResource.Xlink,
			Type:  s.OnlineResource.Type,
			Href:  s.OnlineResource.Href,
		},
		Fees:              s.Fees,
		AccessConstraints: s.AccessConstraints,
	}

	var rootLayer *service.WMSGroupLayer
	var layers map[string]service.WMSLayer
	if s.RootLayer != nil {
		rootLayer = loadWMSRootLayer(s.RootLayer, instance)
	} else {
		layers = make(map[string]service.WMSLayer)
		for i := range s.Layers {
			layers[s.Layers[i].Name] = ConvertWMSLayer(&s.Layers[i], instance)
		}
	}

	supportedSrs := newSupportedSrs(s.Srs, GetPreferredSrcSRS(&globals.Srs))

	imageFormats := make(map[string]*imagery.ImageOptions)
	for _, format := range s.ImageFormats {
		imageFormats[format] = &imagery.ImageOptions{Format: tile.TileFormat(format)}
	}

	infoFormats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		infoFormats[info.Suffix] = info.MimeType
	}
	strict := false
	if s.Strict != nil {
		strict = *s.Strict
	}
	maxOutputPixels := DefaultMaxOutputPixels
	if s.MaxOutputPixels != nil {
		maxOutputPixels = *s.MaxOutputPixels
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}

	ftransformers := loadXSLTransformer(s.FeatureinfoXslt, basePath)

	extents := extentsForSrs(s.BBoxSrs)

	md.Extended = ConvertExtendedCapabilities(s.ExtendedCapabilities)
	md.Contact = ConvertContactInformation(s.ContactInformation)

	wopts := &service.WMSServiceOptions{
		RootLayer:       rootLayer,
		Layers:          layers,
		Metadata:        md,
		Srs:             supportedSrs,
		SrsExtents:      extents,
		ImageFormats:    imageFormats,
		InfoFormats:     infoFormats,
		MaxOutputPixels: maxOutputPixels,
		MaxTileAge:      maxTileAge,
		Strict:          strict,
		Transformers:    ftransformers,
	}

	return service.NewWMSService(wopts)
}

func ConvertExtendedCapabilities(extendedCapabilities *WMSExtendedCapabilities) *wms130.ExtendedCapabilities {
	if extendedCapabilities == nil {
		return nil
	}

	ret := &wms130.ExtendedCapabilities{}

	ret.MetadataURL.URL = extendedCapabilities.MetadataURL.URL
	ret.MetadataURL.MediaType = extendedCapabilities.MetadataURL.MediaType

	ret.SupportedLanguages.DefaultLanguage.Language = extendedCapabilities.SupportedLanguages.DefaultLanguage.Language

	if extendedCapabilities.SupportedLanguages.SupportedLanguage != nil {
		ret.SupportedLanguages.SupportedLanguage = new([]struct {
			Language string "xml:\"inspire_common:Language\" yaml:\"language\""
		})
		for _, lan := range *extendedCapabilities.SupportedLanguages.SupportedLanguage {
			*ret.SupportedLanguages.SupportedLanguage = append(*ret.SupportedLanguages.SupportedLanguage, struct {
				Language string "xml:\"inspire_common:Language\" yaml:\"language\""
			}{Language: lan.Language})
		}
	}
	ret.ResponseLanguage.Language = extendedCapabilities.ResponseLanguage.Language
	return ret
}

func ConvertContactInformation(contactInformation *WMSContactInformation) *wms130.ContactInformation {
	if contactInformation == nil {
		return nil
	}

	ret := &wms130.ContactInformation{}

	ret.ContactPersonPrimary.ContactPerson = contactInformation.ContactPersonPrimary.ContactPerson
	ret.ContactPersonPrimary.ContactOrganization = contactInformation.ContactPersonPrimary.ContactOrganization

	ret.ContactPosition = contactInformation.ContactPosition

	ret.ContactAddress.AddressType = contactInformation.ContactAddress.AddressType
	ret.ContactAddress.Address = contactInformation.ContactAddress.Address
	ret.ContactAddress.City = contactInformation.ContactAddress.City
	ret.ContactAddress.StateOrProvince = contactInformation.ContactAddress.StateOrProvince
	ret.ContactAddress.PostCode = contactInformation.ContactAddress.PostCode
	ret.ContactAddress.Country = contactInformation.ContactAddress.Country

	ret.ContactVoiceTelephone = contactInformation.ContactVoiceTelephone
	ret.ContactFacsimileTelephone = contactInformation.ContactFacsimileTelephone
	ret.ContactElectronicMailAddress = contactInformation.ContactElectronicMailAddress

	return ret
}

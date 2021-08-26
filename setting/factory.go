package setting

import (
	"image/color"
	"path"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
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

func NewImageOptions(opt *ImageOpts) *imagery.ImageOptions {
	image_opt := &imagery.ImageOptions{}
	image_opt.Transparent = opt.Transparent
	image_opt.Format = tile.TileFormat(opt.Format)
	if opt.ResamplingMethod != "" {
		image_opt.Resampling = opt.ResamplingMethod
	} else {
		image_opt.Resampling = DefaultResamplingMethod
	}
	if opt.Colors != nil {
		image_opt.Colors = *opt.Colors
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
		conf[geo.TILEGRID_ORIGIN] = opt.Origin
	}
	if opt.TileSize != nil {
		conf[geo.TILEGRID_TILE_SIZE] = *opt.TileSize
	} else {
		conf[geo.TILEGRID_TILE_SIZE] = DefaultTileSize
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

func LoadCacheManager(c *Caches, globals *GlobalsSetting, instance ProxyInstance) cache.Manager {
	opts := make([]tile.TileOptions, len(c.Sources))
	layers := make([]layer.Layer, len(c.Sources))

	for i := range layers {
		l := instance.GetSource(c.Sources[i])
		layers = append(layers, l)
		opts = append(opts, l.GetOptions())
	}

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
		meta_buffer = globals.Cache.MetaBuffer
	}

	var meta_size [2]uint32
	if c.MetaSize != nil {
		meta_size = [2]uint32{c.MetaSize[0], c.MetaSize[1]}
	} else {
		meta_size = [2]uint32{globals.Cache.MetaSize[0], globals.Cache.MetaSize[1]}
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

	cache_rescaled_tiles := c.CacheRescaledTiles
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
	tile_opts := opts[0]

	var cacheB cache.Cache

	switch cinfo := c.CacheInfo.(type) {
	case *LocalCache:
		cacheB = ConvertLocalCache(cinfo, tile_opts)
	case *S3Cache:
		cacheB = ConvertS3Cache(cinfo, tile_opts)
	}

	var locker cache.TileLocker

	if c.LockDir != "" {
		locker = cache.NewFileTileLocker(c.LockDir, time.Duration(c.LockRetryDelay*int(time.Second)))
	} else {
		locker = &cache.DummyTileLocker{}
	}

	var pre_store_filter []cache.Filter

	if c.WaterMark != nil {
		pre_store_filter = append(pre_store_filter, newWaterMarkFilter(c.WaterMark))
	}

	return cache.NewTileManager(layers, grid.(*geo.TileGrid), cacheB, locker, name, request_format_ext, tile_opts, minimize_meta_requests, bulk_meta_tiles, pre_store_filter, rescale_tiles, *cache_rescaled_tiles, meta_buffer, meta_size)
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

func ConvertLocalCache(opt *LocalCache, opts tile.TileOptions) *cache.LocalCache {
	return cache.NewLocalCache(opt.Directory, opt.DirectoryLayout, cache.GetSourceCreater(opts))
}

func ConvertS3Cache(opt *S3Cache, opts tile.TileOptions) *cache.S3Cache {
	setting := cache.S3Options{}
	setting.Endpoint = opt.Endpoint
	setting.AccessKey = opt.AccessKey
	setting.SecretKey = opt.SecretKey
	setting.Secure = opt.Secure
	setting.SignV2 = opt.SignV2
	setting.Region = opt.Region
	setting.Bucket = opt.Bucket
	setting.Encrypt = opt.Encrypt
	setting.Trace = opt.Trace
	return cache.NewS3Cache(opt.Directory, opt.DirectoryLayout, setting, cache.GetSourceCreater(opts))
}

func ConvertLocalStore(opt *LocalStore) *resource.LocalStore {
	return resource.NewLocalStore(opt.Directory)
}

func ConvertS3Store(opt *S3Store) *resource.S3Store {
	setting := resource.S3Options{}
	setting.Endpoint = opt.Endpoint
	setting.AccessKey = opt.AccessKey
	setting.SecretKey = opt.SecretKey
	setting.Secure = opt.Secure
	setting.SignV2 = opt.SignV2
	setting.Region = opt.Region
	setting.Bucket = opt.Bucket
	setting.Encrypt = opt.Encrypt
	setting.Trace = opt.Trace
	return resource.NewS3Store(opt.Directory, setting)
}

func ConvertMapboxTileLayer(l *MapboxTileLayer, instance ProxyInstance) *service.MapboxTileProvider {
	tp := service.MapboxVector
	if l.TileType != "" {
		tp = service.GetMapboxTileType(l.TileType)
	}
	tileManager := instance.GetCache(l.Source)
	tilejsonSource := LoadTileJSONSource(&l.TileJSON)
	return service.NewMapboxTileProvider(l.Name, tp, l.Metadata, tileManager, tilejsonSource, l.VectorLayers, l.ZoomRange)
}

func ConvertTileLayer(l *TileLayer, instance ProxyInstance) *service.TileProvider {
	dimensions := utils.NewDimensionsFromValues(l.Dimensions)

	tileManager := instance.GetCache(l.TileSource)

	infoSources := []layer.InfoLayer{}
	for _, info := range l.InfoSources {
		infoSources = append(infoSources, instance.GetInfoSource(info))
	}

	return service.NewTileProvider(l.Name, l.Title, l.Metadata, tileManager, infoSources, dimensions, &service.TMSExceptionHandler{})
}

func ConvertWMSLayer(l *WMSLayer, instance ProxyInstance) service.WMSLayer {
	_range := &geo.ResolutionRange{Min: l.MinRes, Max: l.MaxRes}
	mapLayers := make(map[string]layer.Layer)
	infos := make(map[string]layer.InfoLayer)
	legends := make([]layer.LegendLayer, 0)

	for _, name := range l.MapSources {
		mapLayers[name] = instance.GetSource(name)
	}

	for _, name := range l.FeatureinfoSources {
		infos[name] = instance.GetInfoSource(name)
	}

	for _, name := range l.LegendSources {
		legends = append(legends, instance.GetLegendSource(name))
	}

	return service.NewWMSNodeLayer(l.Name, l.Title, mapLayers, infos, legends, _range, l.Metadata)
}

func loadWMSRootLayer(l *WMSLayer, instance ProxyInstance) *service.WMSGroupLayer {
	thisLayer := ConvertWMSLayer(l, instance)
	layers := make(map[string]service.WMSLayer)
	for i := range l.Layers {
		layers[l.Layers[i].Name] = ConvertWMSLayer(&l.Layers[i], instance)
	}
	return service.NewWMSGroupLayer(l.Name, l.Title, thisLayer, layers, l.Metadata)
}

func newSupportedSrs(supportedSrs []string, preferred geo.PreferredSrcSRS) *geo.SupportedSRS {
	srs := []geo.Proj{}
	for i := range supportedSrs {
		srs = append(srs, geo.NewSRSProj4(supportedSrs[i]))
	}
	return &geo.SupportedSRS{Srs: srs, Preferred: preferred}
}

func LoadWMSInfoSource(s *WMSSource, basePath string, preferred geo.PreferredSrcSRS) *sources.WMSInfoSource {
	if s.Opts.FeatureInfo == nil || !*s.Opts.FeatureInfo {
		return nil
	}
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	if s.Opts.FeatureinfoFormat != "" {
		params["info_format"] = []string{s.Opts.FeatureinfoFormat}
	}

	url := s.Request.Url
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

	c := client.NewWMSInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, preferred), newCollectorContext(&s.Http))
	return sources.NewWMSInfoSource(c, coverage, transformer)
}

func LoadWMSLegendsSource(s *WMSSource) *sources.WMSLegendSource {
	if s.Opts.LegendGraphic == nil || !*s.Opts.LegendGraphic {
		return nil
	}
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}
	if s.Request.Transparent != nil && *s.Request.Transparent {
		params["transparent"] = []string{"true"}
	}
	url := s.Request.Url
	lg_clients := make([]*client.WMSLegendClient, 0)
	for _, layer := range s.Request.Layers {
		params["layer"] = []string{layer}
		lg_request := request.NewWMSLegendGraphicRequest(params, url, false, nil, false)
		lg_clients = append(lg_clients, client.NewWMSLegendClient(lg_request, newCollectorContext(&s.Http)))
	}

	var cache *resource.LegendCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewLegendCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewLegendCache(ConvertLocalStore(s))
	}
	return sources.NewWMSLegendSource(s.Opts.LegendID, lg_clients, cache)
}

func LoadWMSMapSource(s *WMSSource, instance ProxyInstance, preferred geo.PreferredSrcSRS) *sources.WMSSource {
	if s.Opts.Map == nil || !*s.Opts.Map {
		return nil
	}

	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	image_opts := NewImageOptions(&s.Image.ImageOpts)
	res_range := NewResolutionRange(&s.ScaleHints)
	supported_srs := newSupportedSrs(s.SupportedSrs, preferred)
	coverage := LoadCoverage(s.Coverage)

	var transparent_color color.Color
	transparent_color_tolerance := s.Image.TransparentColorTolerance

	if s.Image.TransparentColor != nil {
		c := s.Image.TransparentColor
		transparent_color = color.RGBA{R: c[0], G: c[1], B: c[2], A: c[3]}
	}

	url := s.Request.Url

	req := request.NewWMSMapRequest(params, url, false, nil, false)
	c := client.NewWMSClient(req, newCollectorContext(&s.Http))
	return sources.NewWMSSource(c, image_opts, coverage,
		res_range, transparent_color,
		transparent_color_tolerance, supported_srs, s.SupportedFormats,
		s.ForwardReqParams)
}

func newCollectorContext(httpOpts *HttpOpts) *client.CollectorContext {
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
	return client.NewCollectorContext(&conf)
}

func LoadTileSource(s *TileSource, instance ProxyInstance) *sources.TileSource {
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
	coverage := LoadCoverage(s.Coverage)
	res_range := NewResolutionRange(&s.ScaleHints)

	creater := cache.GetSourceCreater(opts)
	tpl := client.NewURLTemplate(s.URLTemplate, s.RequestFormat, s.Subdomains)
	c := client.NewTileClient(grid.(*geo.TileGrid), tpl, newCollectorContext(&s.Http))
	return sources.NewTileSource(grid.(*geo.TileGrid), c, coverage, opts, res_range, creater)
}

func LoadMapboxTileSource(s *MapboxTileSource, instance ProxyInstance) *sources.MapboxTileSource {
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

	creater := cache.GetSourceCreater(opts)
	c := client.NewMapboxTileClient(s.Url, s.UserName, s.AccessToken, s.TilesetID, newCollectorContext(&s.Http))
	return sources.NewMapboxTileSource(grid.(*geo.TileGrid), c, opts, creater)
}

func LoadLuokuangTileSource(s *LuokuangTileSource, instance ProxyInstance) *sources.LuoKuangTileSource {
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

	creater := cache.GetSourceCreater(opts)
	c := client.NewLuoKuangTileClient(s.Url, s.AccessToken, s.TilesetID, newCollectorContext(&s.Http))
	return sources.NewLuoKuangTileSource(grid.(*geo.TileGrid), c, opts, creater)
}

func LoadArcGISSource(s *ArcGISSource, instance ProxyInstance, preferred geo.PreferredSrcSRS) *sources.ArcGISSource {
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	image_opts := NewImageOptions(&s.Image.ImageOpts)
	res_range := NewResolutionRange(&s.ScaleHints)
	supported_srs := newSupportedSrs(s.SupportedSrs, preferred)

	url := s.Request.Url

	coverage := LoadCoverage(s.Coverage)

	req := request.NewArcGISRequest(params, url, false, nil)
	c := client.NewArcGISClient(req, newCollectorContext(&s.Http))
	return sources.NewArcGISSource(c, image_opts, coverage, res_range, supported_srs, s.SupportedFormats)
}

func LoadArcGISInfoSource(s *ArcGISSource, preferred geo.PreferredSrcSRS) *sources.ArcGISInfoSource {
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
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

	url := s.Request.Url

	fi_request := request.NewArcGISIdentifyRequest(params, url, false, nil)

	c := client.NewArcGISInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, preferred), newCollectorContext(&s.Http), return_geometries, tolerance)
	return sources.NewArcGISInfoSource(c)
}

func LoadTileJSONSource(s *TileJSONSource) *sources.MapboxTileJSONSource {
	c := client.NewMapboxTileJSONClient(s.Url, s.UserName, s.AccessToken, newCollectorContext(&s.Http))
	var cache *resource.TileJSONCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewTileJSONCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewTileJSONCache(ConvertLocalStore(s))
	}
	return sources.NewMapboxTileJSONSource(c, cache)
}

func LoadStyleSource(s *StyleSource) *sources.MapboxStyleSource {
	c := client.NewMapboxStyleClient(s.Url, s.UserName, s.AccessToken, newCollectorContext(&s.Http))
	var cache *resource.StyleCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewStyleCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewStyleCache(ConvertLocalStore(s))
	}
	return sources.NewMapboxStyleSource(c, cache)
}

func LoadGlyphsSource(s *GlyphsSource) *sources.MapboxGlyphsSource {
	c := client.NewMapboxGlyphsClient(s.Url, s.UserName, s.AccessToken, newCollectorContext(&s.Http))
	var cache *resource.GlyphsCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewGlyphsCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewGlyphsCache(ConvertLocalStore(s))
	}
	return sources.NewMapboxGlyphsSource(c, cache)
}

func LoadMapboxService(s *MapboxService, instance ProxyInstance) *service.MapboxService {
	layers := make(map[string]service.Provider)
	styles := make(map[string]*service.StyleProvider)
	fonts := make(map[string]*service.GlyphProvider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertMapboxTileLayer(&tl, instance)
	}

	for _, st := range s.Styles {
		styles[st.StyleID] = service.NewStyleProvider(LoadStyleSource(&st))
	}

	for _, ft := range s.Fonts {
		fonts[ft.Font] = service.NewGlyphProvider(LoadGlyphsSource(&ft))
	}

	var max_tile_age *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		max_tile_age = &d
	}

	return service.NewMapboxService(layers, styles, fonts, s.Metadata, max_tile_age)
}

func LoadTMSService(s *TMSService, instance ProxyInstance) *service.TileService {
	layers := make(map[string]service.Provider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var max_tile_age *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		max_tile_age = &d
	}
	origin := s.Origin

	return service.NewTileService(layers, s.Metadata, max_tile_age, false, origin)
}

func LoadWMTSService(s *WMTSService, instance ProxyInstance) *service.WMTSService {
	if s.KVP == nil || !*s.KVP {
		return nil
	}

	layers := make(map[string]service.Provider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var max_tile_age *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		max_tile_age = &d
	}

	info_formats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		info_formats[info.Suffix] = info.MimeType
	}

	return service.NewWMTSService(layers, s.Metadata, max_tile_age, info_formats)
}

func LoadWMTSRestfulService(s *WMTSService, instance ProxyInstance) *service.WMTSRestService {
	if s.Restful == nil || !*s.Restful {
		return nil
	}

	layers := make(map[string]service.Provider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var max_tile_age *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		max_tile_age = &d
	}

	info_formats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		info_formats[info.Suffix] = info.MimeType
	}

	return service.NewWMTSRestService(layers, s.Metadata, max_tile_age, s.RestfulTemplate, s.RestfulFeatureinfoTemplate, info_formats)
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
		e := &geo.MapExtent{BBox: vec2d.Rect{Min: vec2d.T{bbox[0], bbox[1]}, Max: vec2d.T{bbox[2], bbox[3]}}, Srs: geo.NewSRSProj4(srs)}
		extents[srs] = e
	}
	return extents
}

func LoadWMSService(s *WMSService, instance ProxyInstance, basePath string, preferred geo.PreferredSrcSRS) *service.WMSService {
	md := s.Metadata

	rootLayer := loadWMSRootLayer(s.Layer, instance)
	supported_srs := newSupportedSrs(s.Srs, preferred)

	imageFormats := make(map[string]*imagery.ImageOptions)
	for _, format := range s.ImageFormats {
		imageFormats[format] = &imagery.ImageOptions{Format: tile.TileFormat(format)}
	}

	info_formats := make(map[string]string)
	for _, info := range s.FeatureinfoFormats {
		info_formats[info.Suffix] = info.MimeType
	}
	strict := false
	if s.Strict != nil {
		strict = *s.Strict
	}
	maxOutputPixels := DefaultMaxOutputPixels
	if s.MaxOutputPixels != nil {
		maxOutputPixels = *s.MaxOutputPixels
	}

	var max_tile_age *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		max_tile_age = &d
	}

	ftransformers := loadXSLTransformer(s.FeatureinfoXslt, basePath)

	extents := extentsForSrs(s.BBoxSrs)

	return service.NewWMSService(rootLayer, md, supported_srs, imageFormats, info_formats, extents, maxOutputPixels, max_tile_age, strict, ftransformers)
}

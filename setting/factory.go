package setting

import (
	"image/color"
	"path"
	"strconv"
	"strings"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/ogc-specifications/pkg/wsc110"

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
		conf[geo.TILEGRID_ORIGIN] = geo.OriginFromString(opt.Origin)
	}
	if opt.TileSize != nil {
		conf[geo.TILEGRID_TILE_SIZE] = (*opt.TileSize)[:]
	} else {
		conf[geo.TILEGRID_TILE_SIZE] = DefaultTileSize[:]
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

func LoadCacheManager(c *CacheSource, globals *GlobalsSetting, instance ProxyInstance) cache.Manager {
	opts := []tile.TileOptions{}
	layers := []layer.Layer{}

	for i := range c.Sources {
		l := instance.GetSource(c.Sources[i])
		if l != nil {
			layers = append(layers, l)
			opts = append(opts, l.GetOptions())
		} else {
			l := instance.GetCacheSource(c.Sources[i])
			if l != nil {
				layers = append(layers, l)
				opts = append(opts, l.GetOptions())
			}
		}
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

	if c.Filters != nil {
		for i := range c.Filters {
			pre_store_filter = append(pre_store_filter, LoadFilter(c.Filters[i]))
		}
	}

	tilegrid := grid.(*geo.TileGrid)

	return cache.NewTileManager(layers, tilegrid, cacheB, locker, name, request_format_ext, tile_opts, minimize_meta_requests, bulk_meta_tiles, pre_store_filter, rescale_tiles, cache_rescaled_tiles, meta_buffer, meta_size)
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

func ConvertMapboxTileLayer(l *MapboxTileLayer, globals *GlobalsSetting, instance ProxyInstance) *service.MapboxTileProvider {
	tp := service.MapboxVector
	if l.TileType != "" {
		tp = service.GetMapboxTileType(l.TileType)
	}
	tileManager := instance.GetCache(l.Source)

	tilesource := instance.GetSource(l.TileJSON)

	tilejsonSource, ok := tilesource.(layer.MapboxTileJSONLayer)
	if ok {
		return service.NewMapboxTileProvider(l.Name, tp, l.Metadata, tileManager, tilejsonSource, l.VectorLayers, l.ZoomRange)
	}

	return service.NewMapboxTileProvider(l.Name, tp, l.Metadata, tileManager, nil, l.VectorLayers, l.ZoomRange)
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

func ConvertWMSLayerMetadata(metadata *WMSLayerMetadata) *service.WMSLayerMetadata {
	return nil //TODO
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

	return service.NewWMSNodeLayer(l.Name, l.Title, mapLayers, infos, legends, _range, ConvertWMSLayerMetadata(l.Metadata))
}

func loadWMSRootLayer(l *WMSLayer, instance ProxyInstance) *service.WMSGroupLayer {
	thisLayer := ConvertWMSLayer(l, instance)
	layers := make(map[string]service.WMSLayer)
	for i := range l.Layers {
		layers[l.Layers[i].Name] = ConvertWMSLayer(&l.Layers[i], instance)
	}
	return service.NewWMSGroupLayer(l.Name, l.Title, thisLayer, layers, ConvertWMSLayerMetadata(l.Metadata))
}

func newSupportedSrs(supportedSrs []string, preferred geo.PreferredSrcSRS) *geo.SupportedSRS {
	srs := []geo.Proj{}
	for i := range supportedSrs {
		srs = append(srs, geo.NewProj(supportedSrs[i]))
	}
	return &geo.SupportedSRS{Srs: srs, Preferred: preferred}
}

func LoadWMSInfoSource(s *WMSSource, basePath string, globals *GlobalsSetting, preferred geo.PreferredSrcSRS) *sources.WMSInfoSource {
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

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}

	c := client.NewWMSInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, preferred), newCollectorContext(http))

	if s.Opts.Version == "1.1.1" {
		c.AdaptTo111 = true
	}

	return sources.NewWMSInfoSource(c, coverage, transformer)
}

func LoadWMSLegendsSource(s *WMSSource, globals *GlobalsSetting) *sources.WMSLegendSource {
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

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}

	url := s.Request.Url
	lg_clients := make([]*client.WMSLegendClient, 0)
	for _, layer := range s.Request.Layers {
		params["layer"] = []string{layer}
		lg_request := request.NewWMSLegendGraphicRequest(params, url, false, nil, false)
		lg_clients = append(lg_clients, client.NewWMSLegendClient(lg_request, newCollectorContext(http)))
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

func LoadWMSMapSource(s *WMSSource, instance ProxyInstance, globals *GlobalsSetting, preferred geo.PreferredSrcSRS) *sources.WMSSource {
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

	url := s.Request.Url

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}

	req := request.NewWMSMapRequest(params, url, false, nil, false)
	c := client.NewWMSClient(req, newCollectorContext(http))

	if s.Opts.Version == "1.1.1" {
		c.AdaptTo111 = true
	}

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
	if httpOpts.Threads != nil {
		conf.Threads = *httpOpts.Threads
	} else {
		conf.Threads = DefaultThreads
	}
	if httpOpts.MaxQueueSize != nil {
		conf.MaxQueueSize = *httpOpts.MaxQueueSize
	} else {
		conf.MaxQueueSize = DefaultMaxQueueSize
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

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}
	creater := cache.GetSourceCreater(opts)
	tpl := client.NewURLTemplate(s.URLTemplate, s.RequestFormat, s.Subdomains)
	c := client.NewTileClient(grid.(*geo.TileGrid), tpl, newCollectorContext(http))
	return sources.NewTileSource(grid.(*geo.TileGrid), c, coverage, opts, res_range, creater)
}

func LoadMapboxTileSource(s *MapboxTileSource, globals *GlobalsSetting, instance ProxyInstance) *sources.MapboxTileSource {
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

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}

	var tcache *resource.TileJSONCache
	switch s := s.TilejsonStore.(type) {
	case *S3Store:
		tcache = resource.NewTileJSONCache(ConvertS3Store(s))
	case *LocalStore:
		tcache = resource.NewTileJSONCache(ConvertLocalStore(s))
	}

	creater := cache.GetSourceCreater(opts)
	accessTokenName := "access_token"
	if s.AccessTokenName != "" {
		accessTokenName = s.AccessTokenName
	}
	c := client.NewMapboxTileClient(s.Url, s.TilejsonUrl, s.AccessToken, accessTokenName, newCollectorContext(http))
	return sources.NewMapboxTileSource(grid.(*geo.TileGrid), c, opts, creater, tcache)
}

func LoadArcGISSource(s *ArcGISSource, instance ProxyInstance, globals *GlobalsSetting, preferred geo.PreferredSrcSRS) *sources.ArcGISSource {
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format == "" {
		params["format"] = []string{request_format}
	}

	if s.Request.Layers != nil {
		params["layers"] = s.Request.Layers
	}

	if s.Request.Transparent != nil {
		if *s.Request.Transparent {
			params["transparent"] = []string{"true"}
		} else {
			params["transparent"] = []string{"false"}
		}
	}

	if s.Request.PixelType != nil {
		params["pixelType"] = []string{*s.Request.PixelType}
	} else {
		params["pixelType"] = []string{"UNKNOWN"}
	}

	if s.Request.Dpi != nil {
		params["dpi"] = []string{strconv.Itoa(*s.Request.Dpi)}
	}

	if s.Request.Time != nil {
		params["time"] = []string{strconv.FormatInt(*s.Request.Time, 10)}
	}

	if s.Request.LercVersion != nil {
		params["lercVersion"] = []string{strconv.Itoa(*s.Request.LercVersion)}
	}

	if s.Request.CompressionQuality != nil {
		params["compressionQuality"] = []string{strconv.Itoa(*s.Request.CompressionQuality)}
	}

	image_opts := NewImageOptions(&s.Image.ImageOpts)
	res_range := NewResolutionRange(&s.ScaleHints)
	supported_srs := newSupportedSrs(s.SupportedSrs, preferred)

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}
	url := s.Request.Url

	coverage := LoadCoverage(s.Coverage)

	req := request.NewArcGISRequest(params, url)
	c := client.NewArcGISClient(req, newCollectorContext(http))
	return sources.NewArcGISSource(c, image_opts, coverage, res_range, supported_srs, s.SupportedFormats)
}

func LoadArcGISInfoSource(s *ArcGISSource, globals *GlobalsSetting, preferred geo.PreferredSrcSRS) *sources.ArcGISInfoSource {
	params := make(request.RequestParams)

	request_format := s.Request.Format
	if request_format != "" {
		params["format"] = []string{request_format}
	}

	if s.Request.Layers != nil {
		params["layers"] = s.Request.Layers
	}

	if s.Request.Transparent != nil {
		if *s.Request.Transparent {
			params["transparent"] = []string{"true"}
		} else {
			params["transparent"] = []string{"false"}
		}
	}

	if s.Request.Dpi != nil {
		params["dpi"] = []string{strconv.Itoa(*s.Request.Dpi)}
	}

	if s.Request.Time != nil {
		params["time"] = []string{strconv.FormatInt(*s.Request.Time, 10)}
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

	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}
	url := s.Request.Url

	fi_request := request.NewArcGISIdentifyRequest(params, url)

	c := client.NewArcGISInfoClient(fi_request, newSupportedSrs(s.SupportedSrs, preferred), newCollectorContext(http), return_geometries, tolerance)
	return sources.NewArcGISInfoSource(c)
}

func LoadStyleSource(s *MapboxStyleLayer, globals *GlobalsSetting) (style *sources.MapboxStyleSource, glyphs *sources.MapboxGlyphsSource) {
	var http *HttpOpts
	if s.Http != nil {
		http = s.Http
	} else {
		http = &globals.Http.HttpOpts
	}
	accessTokenName := "access_token"
	if s.AccessTokenName != "" {
		accessTokenName = s.AccessTokenName
	}
	c := client.NewMapboxStyleClient(s.Url, s.AccessToken, accessTokenName, newCollectorContext(http))

	if s.StyleContentAttr != nil {
		c.StyleContentAttr = s.StyleContentAttr
	}

	csprite := client.NewMapboxStyleClient(s.Sprite, s.AccessToken, accessTokenName, newCollectorContext(http))
	cglyphs := client.NewMapboxStyleClient(s.Glyphs, s.AccessToken, accessTokenName, newCollectorContext(http))

	var cache *resource.StyleCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewStyleCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewStyleCache(ConvertLocalStore(s))
	}

	var gcache *resource.GlyphsCache
	switch s := s.GlyphsStore.(type) {
	case *S3Store:
		gcache = resource.NewGlyphsCache(ConvertS3Store(s))
	case *LocalStore:
		gcache = resource.NewGlyphsCache(ConvertLocalStore(s))
	}
	return sources.NewMapboxStyleSource(c, csprite, cache), sources.NewMapboxGlyphsSource(cglyphs, s.Fonts, gcache)
}

func LoadMapboxService(s *MapboxService, globals *GlobalsSetting, instance ProxyInstance) *service.MapboxService {
	layers := make(map[string]service.Provider)
	styles := make(map[string]*service.StyleProvider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertMapboxTileLayer(&tl, globals, instance)
	}

	for _, st := range s.Styles {
		sts, glys := LoadStyleSource(&st, globals)
		styles[st.StyleID] = service.NewStyleProvider(sts, glys)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}

	return service.NewMapboxService(layers, styles, s.Metadata, maxTileAge)
}

func LoadTMSService(s *TMSService, instance ProxyInstance) *service.TileService {
	layers := make(map[string]service.Provider)

	for _, tl := range s.Layers {
		layers[tl.Name] = ConvertTileLayer(&tl, instance)
	}

	var maxTileAge *time.Duration

	if s.MaxTileAge != nil {
		d := time.Duration(*s.MaxTileAge * int(time.Hour))
		maxTileAge = &d
	}
	origin := s.Origin

	return service.NewTileService(layers, s.Metadata, maxTileAge, false, origin)
}

func ConvertWMTSServiceProvider(provider *WMTSServiceProvider) *wsc110.ServiceProvider {
	return nil //TODO
}

func LoadWMTSService(s *WMTSService, instance ProxyInstance) *service.WMTSService {
	layers := make(map[string]service.Provider)

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

	return service.NewWMTSService(layers, s.Metadata, maxTileAge, info_formats, ConvertWMTSServiceProvider(s.Provider))
}

func LoadWMTSRestfulService(s *WMTSService, instance ProxyInstance) *service.WMTSRestService {
	if s.Restful == nil || !*s.Restful {
		return nil
	}

	layers := make(map[string]service.Provider)

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

	return service.NewWMTSRestService(layers, s.Metadata, maxTileAge, s.RestfulTemplate, s.RestfulFeatureinfoTemplate, info_formats)
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

func LoadWMSService(s *WMSService, instance ProxyInstance, basePath string, preferred geo.PreferredSrcSRS) *service.WMSService {
	md := s.Metadata

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

	supportedSrs := newSupportedSrs(s.Srs, preferred)

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

	return service.NewWMSService(rootLayer, layers, md, supportedSrs, imageFormats, infoFormats, extents, maxOutputPixels, maxTileAge, strict, ftransformers)
}

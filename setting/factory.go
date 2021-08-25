package setting

import (
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

func ConvertLocalCache(opt *LocalCache, opts tile.TileOptions) *cache.LocalCache {
	ret := cache.NewLocalCache(opt.Directory, opt.DirectoryLayout, cache.GetSourceCreater(opts))
	return ret
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
	ret := cache.NewS3Cache(opt.Directory, opt.DirectoryLayout, setting, cache.GetSourceCreater(opts))
	return ret
}

func ConvertLocalStore(opt *LocalStore) *resource.LocalStore {
	ret := resource.NewLocalStore(opt.Directory)
	return ret
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
	ret := resource.NewS3Store(opt.Directory, setting)
	return ret
}

func ConvertMapboxTileLayer(l *MapboxTileLayer, tileManager cache.Manager) *service.MapboxTileProvider {
	tp := service.MapboxVector
	if l.TileType != "" {
		tp = service.GetMapboxTileType(l.TileType)
	}
	tilejsonSource := LoadTileJSONSource(&l.TileJSON)
	provider := service.NewMapboxTileProvider(l.Name, tp, l.Metadata, tileManager, tilejsonSource, l.VectorLayers, l.ZoomRange)
	return provider
}

func ConvertTileLayer(l *TileLayer, infoSources []layer.InfoLayer, tileManager cache.Manager) *service.TileProvider {
	dimensions := utils.NewDimensionsFromValues(l.Dimensions)
	provider := service.NewTileProvider(l.Name, l.Title, l.Metadata, tileManager, infoSources, dimensions, &service.TMSExceptionHandler{})
	return provider
}

func ConvertWMSLayer(l *WMSLayer, mapLayers map[string]layer.Layer, infos map[string]layer.InfoLayer, legends []layer.LegendLayer) service.WMSLayer {
	_range := &geo.ResolutionRange{Min: l.MinRes, Max: l.MaxRes}
	provider := service.NewWMSNodeLayer(l.Name, l.Title, mapLayers, infos, legends, _range, l.Metadata)
	return provider
}

func LoadWMSRootLayer(l *WMSLayer, mapLayers map[string]layer.Layer, infos map[string]layer.InfoLayer, legends []layer.LegendLayer, layers map[string]service.WMSLayer) *service.WMSGroupLayer {
	thisLayer := ConvertWMSLayer(l, mapLayers, infos, legends)
	provider := service.NewWMSGroupLayer(l.Name, l.Title, thisLayer, layers, l.Metadata)
	return provider
}

func newSupportedSrs(supportedSrs []string, preferred geo.PreferredSrcSRS) *geo.SupportedSRS {
	srs := []geo.Proj{}
	for i := range supportedSrs {
		srs = append(srs, geo.NewSRSProj4(supportedSrs[i]))
	}
	return &geo.SupportedSRS{Srs: srs, Preferred: preferred}
}

func newWMSFeatureInfoRequest(s *WMSSource) *request.WMSFeatureInfoRequest {
	return nil
}

func LoadWMSInfoSource(s *WMSSource, coverage geo.Coverage, preferred geo.PreferredSrcSRS, transformer *resource.XSLTransformer) *sources.WMSInfoSource {
	c := client.NewWMSInfoClient(newWMSFeatureInfoRequest(s), newSupportedSrs(s.SupportedSrs, preferred), newCollectorContext(&s.Http))
	return sources.NewWMSInfoSource(c, coverage, transformer)
}

func newWMSLegendGraphicRequest(s *WMSSource) *request.WMSLegendGraphicRequest {
	return nil
}

func LoadWMSLegendsSource(s *WMSSource) *sources.WMSLegendSource {
	c := client.NewWMSLegendClient(newWMSLegendGraphicRequest(s), newCollectorContext(&s.Http))
	var cache *resource.LegendCache
	switch s := s.Store.(type) {
	case *S3Store:
		cache = resource.NewLegendCache(ConvertS3Store(s))
	case *LocalStore:
		cache = resource.NewLegendCache(ConvertLocalStore(s))
	}
	return sources.NewWMSLegendSource(s.Opts.LegendID, c, cache)
}

func LoadWMSMapSource(s *WMSSource, coverage geo.Coverage) *sources.WMSSource {
	return nil
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

func LoadTileSource(s *TileSource, grid *geo.TileGrid) *sources.TileSource {
	var opts tile.TileOptions
	switch o := s.Options.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}

	creater := cache.GetSourceCreater(opts)
	tpl := client.NewURLTemplate(s.URLTemplate, s.RequestFormat)
	c := client.NewTileClient(grid, tpl, newCollectorContext(&s.Http))
	return sources.NewTileSource(c, opts, creater)
}

func LoadMapboxTileSource(s *MapboxTileSource, grid *geo.TileGrid) *sources.MapboxTileSource {
	var opts tile.TileOptions
	switch o := s.Options.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}

	creater := cache.GetSourceCreater(opts)
	c := client.NewMapboxTileClient(s.Url, s.UserName, s.AccessToken, s.TilesetID, newCollectorContext(&s.Http))
	return sources.NewMapboxTileSource(grid, c, opts, creater)
}

func LoadLuokuangTileSource(s *LuokuangTileSource, grid *geo.TileGrid) *sources.LuoKuangTileSource {
	var opts tile.TileOptions
	switch o := s.Options.(type) {
	case *ImageOpts:
		opts = NewImageOptions(o)
	case *RasterOpts:
		opts = NewRasterOptions(o)
	case *VectorOpts:
		opts = NewVectorOptions(o)
	}

	creater := cache.GetSourceCreater(opts)
	c := client.NewLuoKuangTileClient(s.Url, s.AccessToken, s.TilesetID, newCollectorContext(&s.Http))
	return sources.NewLuoKuangTileSource(grid, c, opts, creater)
}

func LoadArcgisSource(s *ArcgisSource) *sources.ArcGISSource {
	return nil
}

func LoadArcGISInfoSource(s *ArcgisSource) *sources.ArcGISInfoSource {
	return nil
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
	return nil
}

func LoadTMSService(s *TMSService, instance ProxyInstance) *service.TileService {
	return nil
}

func LoadWMTSService(s *WMTSService, instance ProxyInstance) *service.WMTSService {
	return nil
}

func LoadWMSService(s *WMSService, instance ProxyInstance) *service.WMSService {
	return nil
}

package setting

var (
	DefaultServer       = []string{"wms", "tms", "wmts", "mapbox"}
	DefaultImageFormats = []string{"image/png", "image/jpeg", "image/gif", "image/GeoTIFF", "image/tiff"}
	DefaultSRS          = []string{"EPSG:4326", "EPSG:4258", "CRS:84", "EPSG:900913", "EPSG:3857", "EPSG:GCJ02", "EPSG:BD09", "EPSG:GCJ02MC", "EPSG:BDMC"}
)

const (
	DefaultStrict          = false
	DefaultMaxOutputPixels = 4000 * 4000
)

var (
	DefaultAxisOrderNE = []string{"EPSG:4326", "EPSG:4258", "EPSG:31466", "EPSG:31467", "EPSG:31468"}
	DefaultAxisOrderEN = []string{"CRS:84", "EPSG:900913", "EPSG:25831", "EPSG:25832", "EPSG:25833"}
)

const (
	DefaultResamplingMethod          = "bicubic"
	DefaultJpegQuality               = 90
	DefaultStretchFactor             = 1.15
	DefaultMaxShrinkFactor           = 4.0
	DefaultPaletted                  = true
	DefaultTransparentColorTolerance = 5
)

const (
	DefaultConcurrentTileCreators = 2
	DefaultMetaBuffer             = 80
	DefaultMinimizeMetaRequests   = false
)

var (
	DefaultMetaSize = [2]int{4, 4}
	DefaultTileSize = [2]uint32{256, 256}
)

var (
	DefaultGrids = []GridOpts{
		{Srs: "EPSG:4326", Origin: "sw", Name: "GLOBAL_GEODETIC"},
		{Srs: "EPSG:900913", Origin: "sw", Name: "GLOBAL_MERCATOR"},
		{Srs: "EPSG:3857", Origin: "nw", Name: "GLOBAL_WEBMERCATOR"},
		{Srs: "EPSG:GCJ02MC", Origin: "nw", Name: "GCJ02_GLOBAL_WEBMERCATOR"},
		{Srs: "EPSG:BDMC", Origin: "nw", Name: "BD_GLOBAL_WEBMERCATOR"},
	}
)

var (
	DefaultExpiresHours             = 72
	DefaultRequestTimeout           = 60
	DefaultAccessControlAllowOrigin = "*"
	DefaultUserAgent                = "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0"
	DefaultRandomDelay              = 0
	DefaultThreads                  = 4
)

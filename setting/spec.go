package setting

import (
	"time"

	"github.com/flywave/go-tileproxy/resource"
)

type GlobalsSetting struct {
	Image ImageSetting `json:"image,omitempty"`
	Http  struct {
		HttpOpts
		AccessControlAllowOrigin string `json:"access_control_allow_origin,omitempty"`
	} `json:"http,omitempty"`
	Cache struct {
		BaseDir              string   `json:"base_dir,omitempty"`
		LockDir              string   `json:"lock_dir,omitempty"`
		TileLockDir          string   `json:"tile_lock_dir,omitempty"`
		MetaSize             []uint32 `json:"meta_size,omitempty"`
		MetaBuffer           int      `json:"meta_buffer,omitempty"`
		BulkMetaTiles        bool     `json:"bulk_meta_tiles,omitempty"`
		MaxTileLimit         int      `json:"max_tile_limit,omitempty"`
		MinimizeMetaRequests bool     `json:"minimize_meta_requests,omitempty"`
	} `json:"cache,omitempty"`
	Grid struct {
		TileSize []uint32 `json:"tile_size,omitempty"`
	} `json:"grid,omitempty"`
	Srs   Srs   `json:"srs,omitempty"`
	Geoid Geoid `json:"geoid,omitempty"`
	Tiles struct {
		ExpiresHours int `json:"expires_hours,omitempty"`
	}
}

type Coverage struct {
	Polygons     string      `json:"polygons,omitempty"`
	PolygonsSrs  string      `json:"polygons_srs,omitempty"`
	BBox         *[4]float64 `json:"bbox,omitempty"`
	BBoxSrs      string      `json:"bbox_srs,omitempty"`
	Geometry     string      `json:"geometry,omitempty"`
	GeometrySrs  string      `json:"geometry_srs,omitempty"`
	ExpireTiles  [][3]int    `json:"expire_tiles,omitempty"`
	Union        []Coverage  `json:"union,omitempty"`
	Intersection []Coverage  `json:"intersection,omitempty"`
	Difference   []Coverage  `json:"difference,omitempty"`
	Clip         *bool       `json:"clip,omitempty"`
}

type RasterOpts struct {
	Format       string   `json:"format,omitempty"`
	Mode         *string  `json:"mode,omitempty"`
	MaxError     float64  `json:"max_error,omitempty"`
	Nodata       *float64 `json:"nodata,omitempty"`
	Interpolator *string  `json:"interpolator,omitempty"`
	DataType     *string  `json:"data_type,omitempty"`
}

type ImageOpts struct {
	Mode             string                 `json:"mode,omitempty"`
	Colors           *int                   `json:"colors,omitempty"`
	Transparent      *bool                  `json:"transparent,omitempty"`
	ResamplingMethod string                 `json:"resampling_method,omitempty"`
	Format           string                 `json:"format,omitempty"`
	EncodingOptions  map[string]interface{} `json:"encoding_options,omitempty"`
}

type VectorOpts struct {
	Format      string   `json:"format,omitempty"`
	Tolerance   *float64 `json:"tolerance,omitempty"`
	Extent      uint16   `json:"extent,omitempty"`
	Buffer      *uint16  `json:"buffer,omitempty"`
	LineMetrics *bool    `json:"line_metrics,omitempty"`
	MaxZoom     *uint8   `json:"max_zoom,omitempty"`
}

type GridOpts struct {
	Name                 string      `json:"name,omitempty"`
	Srs                  string      `json:"srs,omitempty"`
	BBox                 *[4]float64 `json:"bbox,omitempty"`
	BBoxSrs              string      `json:"bbox_srs,omitempty"`
	NumLevels            *int        `json:"num_levels,omitempty"`
	Resolutions          []float64   `json:"res,omitempty"`
	ResFactor            interface{} `json:"res_factor,omitempty"`
	MaxRes               *float64    `json:"max_res,omitempty"`
	MinRes               *float64    `json:"min_res,omitempty"`
	MaxStretchFactor     *float64    `json:"max_stretch_factor,omitempty"`
	MaxShrinkFactor      *float64    `json:"max_shrink_factor,omitempty"`
	AlignResolutionsWith string      `json:"align_resolutions_with,omitempty"`
	Origin               string      `json:"origin,omitempty"`
	TileSize             *[2]uint32  `json:"tile_size,omitempty"`
	ThresholdRes         []float64   `json:"threshold_res,omitempty"`
}

type ScaleHints struct {
	MaxScale *float64 `json:"max_scale,omitempty"`
	MinScale *float64 `json:"min_scale,omitempty"`
	MaxRes   *float64 `json:"max_res,omitempty"`
	MinRes   *float64 `json:"min_res,omitempty"`
}

type ImageSetting struct {
	ResamplingMethod string               `json:"resampling_method,omitempty"`
	Paletted         *bool                `json:"paletted,omitempty"`
	MaxStretchFactor *float64             `json:"max_stretch_factor,omitempty"`
	MaxShrinkFactor  *float64             `json:"max_shrink_factor,omitempty"`
	FontDir          *string              `json:"font_dir,omitempty"`
	Formats          map[string]ImageOpts `json:"formats,omitempty"`
}

type Geoid struct {
	GeoidDataDir string `json:"geoid_data_dir,omitempty"`
}

type Srs struct {
	AxisOrderNE      []string            `json:"axis_order_ne,omitempty"`
	AxisOrderEN      []string            `json:"axis_order_en,omitempty"`
	ProjDataDir      string              `json:"proj_data_dir,omitempty"`
	PreferredSrcProj map[string][]string `json:"preferred_src_proj,omitempty"`
}

type LocalCache struct {
	DirectoryLayout string `json:"directory_layout,omitempty"`
	UseGridNames    bool   `json:"use_grid_names,omitempty"`
	Directory       string `json:"directory,omitempty"`
	TileLockDir     string `json:"tile_lock_dir,omitempty"`
}

type S3Cache struct {
	DirectoryLayout string `json:"directory_layout,omitempty"`
	Directory       string `json:"directory,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	AccessKey       string `json:"access_key,omitempty"`
	SecretKey       string `json:"secret_key,omitempty"`
	Secure          bool   `json:"secure,omitempty"`
	SignV2          bool   `json:"signv2,omitempty"`
	Region          string `json:"region,omitempty"`
	Bucket          string `json:"bucket,omitempty"`
	Encrypt         bool   `json:"encrypt,omitempty"`
	Trace           bool   `json:"trace,omitempty"`
	TileLockDir     string `json:"tile_lock_dir,omitempty"`
}

type Caches struct {
	Sources              []string    `json:"sources,omitempty"`
	Name                 string      `json:"name,omitempty"`
	Grid                 string      `json:"grid,omitempty"`
	LockDir              string      `json:"lock_dir,omitempty"`
	LockRetryDelay       int         `json:"lock_retry_delay,omitempty"`
	CacheDir             string      `json:"cache_dir,omitempty"`
	MetaSize             []uint32    `json:"meta_size,omitempty"`
	MetaBuffer           *int        `json:"meta_buffer,omitempty"`
	BulkMetaTiles        *bool       `json:"bulk_meta_tiles,omitempty"`
	TileOptions          interface{} `json:"tile_options,omitempty"`
	MaxTileLimit         *int        `json:"max_tile_limit,omitempty"`
	MinimizeMetaRequests *bool       `json:"minimize_meta_requests,omitempty"`
	UseDirectFromLevel   *int        `json:"use_direct_from_level,omitempty"`
	UseDirectFromRes     *float64    `json:"use_direct_from_res,omitempty"`
	DisableStorage       *bool       `json:"disable_storage,omitempty"`
	Format               string      `json:"format,omitempty"`
	RequestFormat        string      `json:"request_format,omitempty"`
	CacheRescaledTiles   *bool       `json:"cache_rescaled_tiles,omitempty"`
	UpscaleTiles         *int        `json:"upscale_tiles,omitempty"`
	DownscaleTiles       *int        `json:"downscale_tiles,omitempty"`
	WaterMark            *WaterMark  `json:"watermark,omitempty"`
	CacheInfo            interface{} `json:"cache,omitempty"`
}

type HttpOpts struct {
	UserAgent         *string        `json:"user_agent,omitempty"`
	RandomDelay       *int           `json:"random_delay,omitempty"`
	DisableKeepAlives *bool          `json:"disable_keep_alives,omitempty"`
	Proxys            []string       `json:"proxys,omitempty"`
	RequestTimeout    *time.Duration `json:"request_timeout,omitempty"`
}

type WMSSourceOpts struct {
	Version              string `json:"version,omitempty"`
	Map                  *bool  `json:"map,omitempty"`
	FeatureInfo          *bool  `json:"featureinfo,omitempty"`
	LegendGraphic        *bool  `json:"legend_graphic,omitempty"`
	LegendURL            string `json:"legend_url,omitempty"`
	LegendID             string `json:"legend_id,omitempty"`
	FeatureinfoFormat    string `json:"featureinfo_format,omitempty"`
	FeatureinfoXslt      string `json:"featureinfo_xslt,omitempty"`
	FeatureinfoOutFormat string `json:"featureinfo_out_format,omitempty"`
}

type SourceCommons struct {
	ScaleHints
	ConcurrentRequests *int      `json:"concurrent_requests,omitempty"`
	Coverage           *Coverage `json:"coverage,omitempty"`
	SeedOnly           *bool     `json:"seed_only,omitempty"`
}

type WMSSource struct {
	SourceCommons
	Opts  WMSSourceOpts `json:"wms_opts"`
	Image struct {
		ImageOpts
		Opacity                   *float64  `json:"opacity,omitempty"`
		TransparentColor          *[4]uint8 `json:"transparent_color,omitempty"`
		TransparentColorTolerance *float64  `json:"transparent_color_tolerance,omitempty"`
	} `json:"image"`
	ForwardReqParams map[string]string `json:"forward_req_params,omitempty"`
	SupportedFormats []string          `json:"supported_formats,omitempty"`
	SupportedSrs     []string          `json:"supported_srs,omitempty"`
	Http             *HttpOpts         `json:"http,omitempty"`
	Request          struct {
		Url         string   `json:"url,omitempty"`
		Layers      []string `json:"layers"`
		Transparent *bool    `json:"transparent,omitempty"`
		Format      string   `json:"format,omitempty"`
	} `json:"req"`
	Store interface{} `json:"store"`
}

type TileSource struct {
	SourceCommons
	URLTemplate   string      `json:"url_template,omitempty"`
	Transparent   *bool       `json:"transparent,omitempty"`
	Options       interface{} `json:"options,omitempty"`
	Grid          string      `json:"grid,omitempty"`
	RequestFormat string      `json:"request_format,omitempty"`
	Subdomains    []string    `json:"subdomains,omitempty"`
	Origin        string      `json:"origin,omitempty"`
	Http          *HttpOpts   `json:"http,omitempty"`
}

type MapboxTileSource struct {
	SourceCommons
	Url         string      `json:"url,omitempty"`
	Version     string      `json:"version,omitempty"`
	TilesetID   string      `json:"tileset_id,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	Options     interface{} `json:"options,omitempty"`
	Grid        string      `json:"grid,omitempty"`
	Http        *HttpOpts   `json:"http,omitempty"`
}

type LuokuangTileSource struct {
	SourceCommons
	Url         string      `json:"url,omitempty"`
	TilesetID   string      `json:"tileset_id,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	Options     interface{} `json:"options,omitempty"`
	Grid        string      `json:"grid,omitempty"`
	Http        *HttpOpts   `json:"http,omitempty"`
}

type ArcGISSource struct {
	SourceCommons
	Image struct {
		ImageOpts
		Opacity                   *float64  `json:"opacity,omitempty"`
		TransparentColor          *[4]uint8 `json:"transparent_color,omitempty"`
		TransparentColorTolerance *float64  `json:"transparent_color_tolerance,omitempty"`
	} `json:"image"`
	Request struct {
		Url         string         `json:"url,omitempty"`
		Dpi         *int           `json:"dpi,omitempty"`
		Layers      []string       `json:"layers"`
		Transparent *bool          `json:"transparent,omitempty"`
		Time        *time.Duration `json:"time,omitempty"`
		Format      string         `json:"format,omitempty"`
	} `json:"req"`
	Opts struct {
		Featureinfo                 *bool `json:"featureinfo,omitempty"`
		FeatureinfoTolerance        *int  `json:"featureinfo_tolerance,omitempty"`
		FeatureinfoReturnGeometries *bool `json:"featureinfo_return_geometries,omitempty"`
	} `json:"opts"`
	SupportedFormats []string  `json:"supported_formats,omitempty"`
	SupportedSrs     []string  `json:"supported_srs,omitempty"`
	Http             *HttpOpts `json:"http,omitempty"`
}

type WaterMark struct {
	Text     string    `json:"text"`
	FontSize *int      `json:"font_size,omitempty"`
	Color    *[4]uint8 `json:"color,omitempty"`
	Opacity  *float64  `json:"opacity,omitempty"`
	Spacing  *string   `json:"spacing,omitempty"`
}

type MapboxService struct {
	Metadata   map[string]string `json:"metadata,omitempty"`
	Layers     []MapboxTileLayer `json:"layers,omitempty"`
	Styles     []StyleSource     `json:"styles,omitempty"`
	Fonts      []GlyphsSource    `json:"fonts,omitempty"`
	MaxTileAge *int              `json:"max_tile_age,omitempty"`
}

type TMSService struct {
	Metadata     map[string]string `json:"metadata,omitempty"`
	UseGridNames *bool             `json:"use_grid_names,omitempty"`
	Origin       string            `json:"origin,omitempty"`
	Layers       []TileLayer       `json:"layers,omitempty"`
	MaxTileAge   *int              `json:"max_tile_age,omitempty"`
}

type WMTSService struct {
	KVP                        *bool             `json:"kvp,omitempty"`
	Restful                    *bool             `json:"restful,omitempty"`
	RestfulTemplate            string            `json:"restful_template,omitempty"`
	RestfulFeatureinfoTemplate string            `json:"restful_featureinfo_template,omitempty"`
	Metadata                   map[string]string `json:"metadata,omitempty"`
	FeatureinfoFormats         []struct {
		MimeType string `json:"mimetype"`
		Suffix   string `json:"suffix,omitempty"`
	} `json:"featureinfo_formats,omitempty"`
	Layers     []TileLayer `json:"layers,omitempty"`
	MaxTileAge *int        `json:"max_tile_age,omitempty"`
}

type BBoxSrs struct {
	Srs  string     `json:"srs"`
	BBox [4]float64 `json:"bbox,omitempty"`
}

type WMSService struct {
	Srs                []string  `json:"srs,omitempty"`
	BBoxSrs            []BBoxSrs `json:"bbox_srs,omitempty"`
	ImageFormats       []string  `json:"image_formats,omitempty"`
	FeatureinfoFormats []struct {
		MimeType string `json:"mimetype"`
		Suffix   string `json:"suffix,omitempty"`
	} `json:"featureinfo_formats,omitempty"`
	FeatureinfoXslt map[string]string `json:"featureinfo_xslt,omitempty"`
	MaxOutputPixels *int              `json:"max_output_pixels,omitempty"`
	Strict          *bool             `json:"strict,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	InspireMetadata map[string]string `json:"inspire_metadata,omitempty"`
	Layer           *WMSLayer         `json:"layer,omitempty"`
	MaxTileAge      *int              `json:"max_tile_age,omitempty"`
}

type MapboxTileLayer struct {
	Source       string                  `json:"source"`
	Name         string                  `json:"name,omitempty"`
	Metadata     map[string]string       `json:"metadata,omitempty"`
	TileJSON     TileJSONSource          `json:"tilejson,omitempty"`
	VectorLayers []*resource.VectorLayer `json:"vector_layers,omitempty"`
	TileType     string                  `json:"tile_type,omitempty"`
	ZoomRange    *[2]int                 `json:"zoom_range,omitempty"`
}

type TileLayer struct {
	ScaleHints
	TileSource  string                   `json:"tile_source"`
	InfoSources []string                 `json:"info_sources"`
	Name        string                   `json:"name,omitempty"`
	Title       string                   `json:"title"`
	LegendURL   string                   `json:"legend_url,omitempty"`
	Metadata    map[string]string        `json:"metadata,omitempty"`
	Dimensions  map[string][]interface{} `json:"dimensions,omitempty"`
	LegendStore interface{}              `json:"legendstore"`
}

type WMSLayer struct {
	ScaleHints
	MapSources         []string                 `json:"map_source"`
	FeatureinfoSources []string                 `json:"featureinfo_sources,omitempty"`
	LegendSources      []string                 `json:"legend_sources,omitempty"`
	Name               string                   `json:"name,omitempty"`
	Title              string                   `json:"title"`
	LegendURL          string                   `json:"legend_url,omitempty"`
	Metadata           map[string]string        `json:"metadata,omitempty"`
	Layers             []WMSLayer               `json:"layers,omitempty"`
	Dimensions         map[string][]interface{} `json:"dimensions,omitempty"`
	LegendStore        interface{}              `json:"legendstore"`
}

type TimeSpec struct {
	Seconds int       `json:"seconds,omitempty"`
	Minutes int       `json:"minutes,omitempty"`
	Hours   int       `json:"hours,omitempty"`
	Days    int       `json:"days,omitempty"`
	Weeks   int       `json:"weeks,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

type Seed struct {
	Caches        []string  `json:"caches,omitempty"`
	Grids         []string  `json:"grids,omitempty"`
	Coverages     []string  `json:"coverages,omitempty"`
	Levels        []int     `json:"levels,omitempty"`
	Resolutions   []float64 `json:"resolutions,omitempty"`
	RefreshBefore TimeSpec  `json:"refresh_before,omitempty"`
}

type Cleanup struct {
	Caches       []string  `json:"caches,omitempty"`
	Grids        []string  `json:"grids,omitempty"`
	Coverages    []string  `json:"coverages,omitempty"`
	Levels       []int     `json:"levels,omitempty"`
	Resolutions  []float64 `json:"resolutions,omitempty"`
	RemoveBefore TimeSpec  `json:"remove_before,omitempty"`
	RemoveAll    bool      `json:"remove_all,omitempty"`
}

type LocalStore struct {
	Directory string `json:"directory,omitempty"`
}

type S3Store struct {
	Directory string `json:"directory,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Secure    bool   `json:"secure,omitempty"`
	SignV2    bool   `json:"signv2,omitempty"`
	Region    string `json:"region,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Encrypt   bool   `json:"encrypt,omitempty"`
	Trace     bool   `json:"trace,omitempty"`
}

type StyleSource struct {
	Url         string      `json:"url,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	StyleID     string      `json:"style_id,omitempty"`
	Store       interface{} `json:"store"`
	Http        *HttpOpts   `json:"http,omitempty"`
}

type GlyphsSource struct {
	Url         string      `json:"url,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	Font        string      `json:"font,omitempty"`
	Store       interface{} `json:"store"`
	Http        *HttpOpts   `json:"http,omitempty"`
}

type TileJSONSource struct {
	Url         string      `json:"url,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	TilesetID   string      `json:"tileset_id,omitempty"`
	Store       interface{} `json:"store"`
	Http        *HttpOpts   `json:"http,omitempty"`
}

package setting

import "time"

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

type ImageOpts struct {
	Mode             string            `json:"mode,omitempty"`
	Colors           *uint32           `json:"colors,omitempty"`
	Transparent      *bool             `json:"transparent,omitempty"`
	ResamplingMethod string            `json:"resampling_method,omitempty"`
	Format           string            `json:"format,omitempty"`
	EncodingOptions  map[string]string `json:"encoding_options,omitempty"`
}

type GridOpts struct {
	Base                 string      `json:"base,omitempty"`
	Name                 string      `json:"name,omitempty"`
	Srs                  string      `json:"srs,omitempty"`
	BBox                 *[4]float64 `json:"bbox,omitempty"`
	BBoxSrs              string      `json:"bbox_srs,omitempty"`
	NumLevels            *int        `json:"num_levels,omitempty"`
	Resolutions          []float64   `json:"res,omitempty"`
	ResFactor            interface{} `json:"res_factor,omitempty"`
	MaxRes               *float64    `json:"max_res,omitempty"`
	MinRes               *float64    `json:"min_res,omitempty"`
	StretchFactor        *float64    `json:"stretch_factor,omitempty"`
	MaxStretchFactor     *float64    `json:"max_shrink_factor,omitempty"`
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
	StretchFactor    *float64             `json:"stretch_factor,omitempty"`
	MaxStretchFactor *float64             `json:"max_shrink_factor,omitempty"`
	JpegQuality      *float64             `json:"jpeg_quality,omitempty"`
	FontDir          *string              `json:"font_dir,omitempty"`
	Formats          map[string]ImageOpts `json:"formats,omitempty"`
}

type Srs struct {
	AxisOrderNE      []string            `json:"axis_order_ne,omitempty"`
	AxisOrderEN      []string            `json:"axis_order_en,omitempty"`
	ProjDataDir      string              `json:"proj_data_dir,omitempty"`
	PreferredSrcProj map[string][]string `json:"preferred_src_proj,omitempty"`
}

type Caches struct {
	Name                   string                 `json:"name,omitempty"`
	Grids                  []string               `json:"grids,omitempty"`
	LockDir                string                 `json:"lock_dir,omitempty"`
	TileLockDir            string                 `json:"tile_lock_dir,omitempty"`
	CacheDir               string                 `json:"cache_dir,omitempty"`
	MetaSize               []int                  `json:"meta_size,omitempty"`
	MetaBuffer             *int                   `json:"meta_buffer,omitempty"`
	BulkMetaTiles          *bool                  `json:"bulk_meta_tiles,omitempty"`
	TileOptions            interface{}            `json:"tile_options,omitempty"`
	MaxTileLimit           *int                   `json:"max_tile_limit,omitempty"`
	MinimizeMetaRequests   *bool                  `json:"minimize_meta_requests,omitempty"`
	ConcurrentTileCreators *int                   `json:"concurrent_tile_creators,omitempty"`
	UseDirectFromLevel     *int                   `json:"use_direct_from_level,omitempty"`
	UseDirectFromRes       *float64               `json:"use_direct_from_res,omitempty"`
	DisableStorage         *bool                  `json:"disable_storage,omitempty"`
	Format                 string                 `json:"format,omitempty"`
	RequestFormat          string                 `json:"request_format,omitempty"`
	CacheRescaledTiles     *bool                  `json:"cache_rescaled_tiles,omitempty"`
	UpscaleTiles           *int                   `json:"upscale_tiles,omitempty"`
	DownscaleTiles         *int                   `json:"downscale_tiles,omitempty"`
	WaterMark              *WaterMark             `json:"watermark,omitempty"`
	CacheInfo              map[string]interface{} `json:"cache,omitempty"`
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
	LegendGraphic        *bool  `json:"legendgraphic,omitempty"`
	LegendURL            string `json:"legendurl,omitempty"`
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
	ForwardReqParams []string `json:"forward_req_params,omitempty"`
	SupportedFormats []string `json:"supported_formats,omitempty"`
	SupportedSrs     []string `json:"supported_srs,omitempty"`
	Http             HttpOpts `json:"http,omitempty"`
	Request          struct {
		Url         string   `json:"url,omitempty"`
		Layers      []string `json:"layers"`
		Transparent *bool    `json:"transparent,omitempty"`
	} `json:"req"`
}

type TileSource struct {
	SourceCommons
	Url           string      `json:"url,omitempty"`
	Transparent   *bool       `json:"transparent,omitempty"`
	Options       interface{} `json:"options,omitempty"`
	Grid          string      `json:"grid,omitempty"`
	RequestFormat string      `json:"request_format,omitempty"`
	Origin        string      `json:"origin,omitempty"`
	Http          HttpOpts    `json:"http,omitempty"`
}

type ArcgisSource struct {
	SourceCommons
	Request struct {
		Url         string         `json:"url,omitempty"`
		Dpi         *int           `json:"dpi,omitempty"`
		Layers      []string       `json:"layers"`
		Transparent *bool          `json:"transparent,omitempty"`
		Time        *time.Duration `json:"time,omitempty"`
	} `json:"req"`
	Opts struct {
		Featureinfo                 *bool    `json:"featureinfo,omitempty"`
		FeatureinfoTolerance        *float64 `json:"featureinfo_tolerance,omitempty"`
		FeatureinfoReturnGeometries *bool    `json:"featureinfo_return_geometries,omitempty"`
	} `json:"opts"`
	SupportedSrs []string `json:"supported_srs,omitempty"`
	Http         HttpOpts `json:"http,omitempty"`
}

type WaterMark struct {
	Text     string    `json:"text"`
	FontSize *int      `json:"font_size,omitempty"`
	Color    *[4]uint8 `json:"color,omitempty"`
	Opacity  *float64  `json:"opacity,omitempty"`
	Spacing  *string   `json:"spacing,omitempty"`
}

type MapboxService struct {
	Metadata map[string]string `json:"metadata,omitempty"`
	Layers   []TileLayer       `json:"layers,omitempty"`
}

type TMSService struct {
	UseGridNames *bool       `json:"use_grid_names,omitempty"`
	Origin       string      `json:"origin,omitempty"`
	Layers       []TileLayer `json:"layers,omitempty"`
}

type WMTSService struct {
	KVP                        *bool             `json:"kvp,omitempty"`
	Restful                    *bool             `json:"restful,omitempty"`
	RestfulTemplate            string            `json:"restful_template,omitempty"`
	RestfulFeatureinfoTemplate string            `json:"restful_featureinfo_template,omitempty"`
	Metadata                   map[string]string `json:"metadata,omitempty"`
	FeatureinfoFormats         struct {
		MimeType string `json:"mimetype"`
		Suffix   string `json:"suffix,omitempty"`
	} `json:"featureinfo_formats,omitempty"`
	Layers []TileLayer `json:"layers,omitempty"`
}

type WMSService struct {
	Srs              []string          `json:"srs,omitempty"`
	BBox             *[4]float64       `json:"bbox,omitempty"`
	BBoxSrs          string            `json:"bbox_srs,omitempty"`
	ImageFormats     []string          `json:"image_formats,omitempty"`
	FeatureinfoTypes []string          `json:"featureinfo_types,omitempty"`
	FeatureinfoXslt  map[string]string `json:"featureinfo_xslt,omitempty"`
	MaxOutputPixels  *int              `json:"max_output_pixels,omitempty"`
	Strict           *bool             `json:"strict,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	InspireMetadata  map[string]string `json:"inspire_metadata,omitempty"`
	Layers           []WMSLayer        `json:"layers,omitempty"`
}

type Dimension struct {
	Default interface{}   `json:"default,omitempty"`
	Values  []interface{} `json:"values"`
}

type TileLayer struct {
	ScaleHints
	Sources     []string             `json:"sources"`
	TileSources []string             `json:"tile_sources"`
	Name        string               `json:"name,omitempty"`
	Title       string               `json:"title"`
	LegendURL   string               `json:"legendurl,omitempty"`
	Metadata    map[string]string    `json:"metadata,omitempty"`
	Dimensions  map[string]Dimension `json:"dimensions,omitempty"`
}

type WMSLayer struct {
	ScaleHints
	MapSources         []string             `json:"map_sources"`
	FeatureinfoSources []string             `json:"featureinfo_sources,omitempty"`
	LegendSources      []string             `json:"legend_sources,omitempty"`
	Name               string               `json:"name,omitempty"`
	Title              string               `json:"title"`
	LegendURL          string               `json:"legendurl,omitempty"`
	Metadata           map[string]string    `json:"metadata,omitempty"`
	Layers             []WMSLayer           `json:"layers,omitempty"`
	Dimensions         map[string]Dimension `json:"dimensions,omitempty"`
}

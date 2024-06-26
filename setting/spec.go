package setting

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/flywave/go-tileproxy/resource"
)

type SourceType string

const (
	NONE_SOURCE       SourceType = "none"
	WMS_SOURCE        SourceType = "wms"
	TILE_SOURCE       SourceType = "tile"
	MAPBOXTILE_SOURCE SourceType = "mapbox"
	ARCGIS_SOURCE     SourceType = "arcgis"
	CESIUMTILE_SOURCE SourceType = "cesium"
)

type ServiceType string

const (
	NONE_SERVICE   ServiceType = "none"
	WMS_SERVICE    ServiceType = "wms"
	TMS_SERVICE    ServiceType = "tms"
	WMTS_SERVICE   ServiceType = "wmts"
	MAPBOX_SERVICE ServiceType = "mapbox"
	CESIUM_SERVICE ServiceType = "cesium"
)

type CacheType string

const (
	CACHE_TYPE_FILE   CacheType = "local"
	CACHE_TYPE_CUSTOM CacheType = "custom"
)

type FilterType string

const (
	FILTER_TYPE_WATERMARK FilterType = "watermark"
)

const (
	TILE_OPTION_RASTER = "raster"
	TILE_OPTION_VECTOR = "vector"
	TILE_OPTION_IMAGE  = "image"
)

const (
	DIRECTORY_LAYOUT_TC      = "tc"
	DIRECTORY_LAYOUT_MP      = "mp"
	DIRECTORY_LAYOUT_TMS     = "tms"
	DIRECTORY_LAYOUT_RE_TMS  = "reverse_tms"
	DIRECTORY_LAYOUT_QUADKEY = "quadkey"
	DIRECTORY_LAYOUT_ARCGIS  = "arcgis"
)

type ImageSetting struct {
	ResamplingMethod string               `json:"resampling_method,omitempty"`
	Paletted         *bool                `json:"paletted,omitempty"`
	MaxStretchFactor *float64             `json:"max_stretch_factor,omitempty"`
	MaxShrinkFactor  *float64             `json:"max_shrink_factor,omitempty"`
	FontDir          *string              `json:"font_dir,omitempty"`
	Formats          map[string]ImageOpts `json:"formats,omitempty"`
}

type HttpSetting struct {
	SiteURL           string         `json:"site_url,omitempty"`
	UserAgent         *string        `json:"user_agent,omitempty"`
	RandomDelay       *int           `json:"random_delay,omitempty"`
	DisableKeepAlives *bool          `json:"disable_keep_alives,omitempty"`
	Proxys            []string       `json:"proxys,omitempty"`
	RequestTimeout    *time.Duration `json:"request_timeout,omitempty"`
	Threads           *int           `json:"thread_size,omitempty"`
}

type GlobalsSetting struct {
	Image ImageSetting `json:"image,omitempty"`
	Http  struct {
		HttpSetting
		AccessControlAllowOrigin string `json:"access_control_allow_origin,omitempty"`
	} `json:"http,omitempty"`
	Cache struct {
		BaseDir              string   `json:"base_dir,omitempty"`
		MetaSize             []uint32 `json:"meta_size,omitempty"`
		MetaBuffer           int      `json:"meta_buffer,omitempty"`
		BulkMetaTiles        bool     `json:"bulk_meta_tiles,omitempty"`
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
	Name         string      `json:"name"`
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
	Type         string   `json:"type"`
	Format       string   `json:"format,omitempty"`
	Mode         *string  `json:"mode,omitempty"`
	MaxError     float64  `json:"max_error,omitempty"`
	Nodata       *float64 `json:"nodata,omitempty"`
	Interpolator *string  `json:"interpolator,omitempty"`
	DataType     *string  `json:"data_type,omitempty"`
	HeightModel  *string  `json:"height_model,omitempty"`
	HeightOffset *float64 `json:"height_offset,omitempty"`
}

type ImageOpts struct {
	Type             string                 `json:"type"`
	Mode             string                 `json:"mode,omitempty"`
	Transparent      *bool                  `json:"transparent,omitempty"`
	Opacity          *float64               `json:"opacity,omitempty"`
	ResamplingMethod string                 `json:"resampling_method,omitempty"`
	Format           string                 `json:"format,omitempty"`
	EncodingOptions  map[string]interface{} `json:"encoding_options,omitempty"`
	BgColor          *[4]uint8              `json:"bgcolor,omitempty"`
}

type VectorOpts struct {
	Type        string   `json:"type"`
	Format      string   `json:"format,omitempty"`
	Tolerance   *float64 `json:"tolerance,omitempty"`
	Extent      uint16   `json:"extent,omitempty"`
	Buffer      *uint16  `json:"buffer,omitempty"`
	LineMetrics *bool    `json:"line_metrics,omitempty"`
	MaxZoom     *uint8   `json:"max_zoom,omitempty"`
	Proto       *int     `json:"proto,omitempty"`
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
	InitialResMin        *bool       `json:"initial_res_min,omitempty"`
}

type ScaleHints struct {
	MaxScale *float64 `json:"max_scale,omitempty"`
	MinScale *float64 `json:"min_scale,omitempty"`
	MaxRes   *float64 `json:"max_res,omitempty"`
	MinRes   *float64 `json:"min_res,omitempty"`
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

type CacheInfo struct {
	Type            CacheType `json:"type,omitempty"`
	DirectoryLayout string    `json:"directory_layout,omitempty"`
	Directory       string    `json:"directory,omitempty"`
}

type Reproject struct {
	SrcSrs  string `json:"src_srs,omitempty"`
	DestSrs string `json:"dest_srs,omitempty"`
}

type CacheSource struct {
	Type                 SourceType    `json:"type,omitempty"`
	Sources              []string      `json:"sources,omitempty"`
	Name                 string        `json:"name,omitempty"`
	Grid                 string        `json:"grid,omitempty"`
	LockDir              string        `json:"lock_dir,omitempty"`
	LockRetryDelay       int           `json:"lock_retry_delay,omitempty"`
	MetaSize             []uint32      `json:"meta_size,omitempty"`
	MetaBuffer           *int          `json:"meta_buffer,omitempty"`
	BulkMetaTiles        *bool         `json:"bulk_meta_tiles,omitempty"`
	MinimizeMetaRequests *bool         `json:"minimize_meta_requests,omitempty"`
	Format               string        `json:"format,omitempty"`
	RequestFormat        string        `json:"request_format,omitempty"`
	CacheRescaledTiles   *bool         `json:"cache_rescaled_tiles,omitempty"`
	UpscaleTiles         *int          `json:"upscale_tiles,omitempty"`
	DownscaleTiles       *int          `json:"downscale_tiles,omitempty"`
	Filters              []interface{} `json:"filters,omitempty"`
	CacheInfo            *CacheInfo    `json:"cache,omitempty"`
	ReprojectSrs         *Reproject    `json:"reproject,omitempty"`
	QueryBuffer          *int          `json:"query_buffer,omitempty"`
	TileOptions          interface{}   `json:"tile_options,omitempty"`
}

func (c *CacheSource) FromJson(data []byte) error {
	err := json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	tmp := struct {
		TileOptions interface{} `json:"tile_options,omitempty"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.TileOptions != nil {
		opt, err := TileOptionUnmarshal(tmp.TileOptions)
		if err != nil {
			return err
		}
		c.TileOptions = opt
	}
	return nil
}

func TileOptionUnmarshal(obj interface{}) (interface{}, error) {
	bt, _ := json.Marshal(obj)
	mp := make(map[string]interface{})
	json.Unmarshal(bt, &mp)
	ty, ok := mp["type"]
	if !ok {
		return nil, errors.New("options lost type")
	}
	switch ty.(string) {
	case TILE_OPTION_IMAGE:
		opt := &ImageOpts{}
		json.Unmarshal(bt, opt)
		return opt, nil
	case TILE_OPTION_VECTOR:
		opt := &VectorOpts{}
		json.Unmarshal(bt, opt)
		return opt, nil
	case TILE_OPTION_RASTER:
		opt := &RasterOpts{}
		json.Unmarshal(bt, opt)
		return opt, nil
	}
	return nil, errors.New("not support type")
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

type Source interface {
	GetCoverage() *Coverage
}

type SourceCommons struct {
	ScaleHints
	Coverage *Coverage    `json:"coverage,omitempty"`
	Http     *HttpSetting `json:"http,omitempty"`
	Name     string       `json:"name"`
}

func (s *SourceCommons) GetCoverage() *Coverage {
	return s.Coverage
}

type WMSImageOpts struct {
	ImageOpts
	TransparentColor          *[4]uint8 `json:"transparent_color,omitempty"`
	TransparentColorTolerance *float64  `json:"transparent_color_tolerance,omitempty"`
}

type WMSSource struct {
	SourceCommons
	Type             SourceType        `json:"type,omitempty"`
	Opts             WMSSourceOpts     `json:"wms_opts"`
	Image            WMSImageOpts      `json:"image"`
	ForwardReqParams map[string]string `json:"forward_req_params,omitempty"`
	SupportedFormats []string          `json:"supported_formats,omitempty"`
	SupportedSrs     []string          `json:"supported_srs,omitempty"`
	Url              string            `json:"url,omitempty"`
	Layers           []string          `json:"layers"`
	Transparent      *bool             `json:"transparent,omitempty"`
	Format           string            `json:"format,omitempty"`
	Store            *StoreInfo        `json:"store"`
	AccessToken      *string           `json:"access_token,omitempty"`
	AccessTokenName  *string           `json:"access_token_name,omitempty"`
}

type TileSource struct {
	SourceCommons
	Type          SourceType  `json:"type,omitempty"`
	URLTemplate   string      `json:"url_template,omitempty"`
	AccessToken   *string     `json:"access_token,omitempty"`
	Grid          string      `json:"grid,omitempty"`
	RequestFormat string      `json:"request_format,omitempty"`
	Subdomains    []string    `json:"subdomains,omitempty"`
	Options       interface{} `json:"options,omitempty"`
}

func (c *TileSource) FromJson(data []byte) error {
	err := json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	tmp := struct {
		Options interface{} `json:"options,omitempty"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	opt, err := TileOptionUnmarshal(tmp.Options)
	if err != nil {
		return err
	}
	c.Options = opt
	return nil
}

type MapboxTileSource struct {
	SourceCommons
	Type            SourceType  `json:"type,omitempty"`
	Url             string      `json:"url,omitempty"`
	Tiles           []string    `json:"tiles,omitempty"`
	TileStatsUrl    string      `json:"tile_stats_url,omitempty"`
	Sku             string      `json:"sku,omitempty"`
	AccessToken     string      `json:"access_token,omitempty"`
	AccessTokenName string      `json:"access_token_name,omitempty"`
	Grid            string      `json:"grid,omitempty"`
	ResourceStore   *StoreInfo  `json:"resource_store"`
	Options         interface{} `json:"options,omitempty"`
}

func (c *MapboxTileSource) FromJson(data []byte) error {
	err := json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	tmp := struct {
		Options interface{} `json:"options,omitempty"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.Options != nil {
		opt, err := TileOptionUnmarshal(tmp.Options)
		if err != nil {
			return err
		}
		c.Options = opt
	}
	return nil
}

type CesiumTileSource struct {
	SourceCommons
	Type           SourceType  `json:"type,omitempty"`
	AuthUrl        string      `json:"auth_url,omitempty"`
	Url            string      `json:"url,omitempty"`
	Version        string      `json:"version,omitempty"`
	AssetId        int         `json:"asset_id,omitempty"`
	AccessToken    string      `json:"access_token,omitempty"`
	Grid           string      `json:"grid,omitempty"`
	LayerjsonStore *StoreInfo  `json:"layerjson_store"`
	Options        interface{} `json:"options,omitempty"`
}

func (c *CesiumTileSource) FromJson(data []byte) error {
	err := json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	tmp := struct {
		Options interface{} `json:"options,omitempty"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.Options != nil {
		opt, err := TileOptionUnmarshal(tmp.Options)
		if err != nil {
			return err
		}
		c.Options = opt
	}
	return nil
}

type ArcGISSourceOpts struct {
	Featureinfo                 *bool   `json:"featureinfo,omitempty"`
	FeatureinfoTolerance        *int    `json:"featureinfo_tolerance,omitempty"`
	FeatureinfoReturnGeometries *bool   `json:"featureinfo_return_geometries,omitempty"`
	AccessToken                 *string `json:"access_token,omitempty"`
	AccessTokenName             *string `json:"access_token_name,omitempty"`
}

type ArcGISSource struct {
	SourceCommons
	Type               SourceType       `json:"type,omitempty"`
	Image              WMSImageOpts     `json:"image"`
	Url                string           `json:"url,omitempty"`
	Dpi                *int             `json:"dpi,omitempty"`
	Layers             []string         `json:"layers"`
	Transparent        *bool            `json:"transparent,omitempty"`
	Format             string           `json:"format,omitempty"`
	LercVersion        *int             `json:"lerc_version,omitempty"`
	PixelType          *string          `json:"pixel_type,omitempty"`
	CompressionQuality *int             `json:"compression_quality,omitempty"`
	Opts               ArcGISSourceOpts `json:"opts"`
	SupportedFormats   []string         `json:"supported_formats,omitempty"`
	SupportedSrs       []string         `json:"supported_srs,omitempty"`
}

type WaterMark struct {
	Type     string    `json:"type"`
	Text     string    `json:"text"`
	FontSize *int      `json:"font_size,omitempty"`
	Color    *[4]uint8 `json:"color,omitempty"`
	Opacity  *float64  `json:"opacity,omitempty"`
	Spacing  *string   `json:"spacing,omitempty"`
}

type MapboxTileLayer struct {
	Source       string                  `json:"source"`
	Name         string                  `json:"name,omitempty"`
	Title        string                  `json:"title"`
	VectorLayers []*resource.VectorLayer `json:"vector_layers,omitempty"`
	TileType     string                  `json:"tile_type,omitempty"`
	ZoomRange    *[2]int                 `json:"zoom_range,omitempty"`
	TileJSON     string                  `json:"tilejson,omitempty"`
	Attribution  *string                 `json:"attribution,omitempty"`
	Description  *string                 `json:"description,omitempty"`
	Legend       *string                 `json:"legend,omitempty"`
	FillZoom     *uint32                 `json:"fill_zoom,omitempty"`
}

type MapboxService struct {
	Type       string            `json:"type,omitempty"`
	Layers     []MapboxTileLayer `json:"layers,omitempty"`
	MaxTileAge *int              `json:"max_tile_age,omitempty"`
}

type CesiumTileLayer struct {
	Source      string  `json:"source"`
	Name        string  `json:"name,omitempty"`
	Title       string  `json:"title"`
	ZoomRange   *[2]int `json:"zoom_range,omitempty"`
	LayerJSON   string  `json:"layerjson,omitempty"`
	Attribution *string `json:"attribution,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CesiumService struct {
	Type       string            `json:"type,omitempty"`
	Layers     []CesiumTileLayer `json:"layers,omitempty"`
	MaxTileAge *int              `json:"max_tile_age,omitempty"`
}

type TileLayer struct {
	Name        string                   `json:"name,omitempty"`
	Title       string                   `json:"title,omitempty"`
	Description string                   `json:"description,omitempty"`
	Source      string                   `json:"source"`
	InfoSources []string                 `json:"info_sources"`
	Dimensions  map[string][]interface{} `json:"dimensions,omitempty"`
}

type TMSService struct {
	Type       string      `json:"type,omitempty"`
	Title      string      `json:"title,omitempty"`
	Abstract   string      `json:"abstract,omitempty"`
	Origin     string      `json:"origin,omitempty"`
	Layers     []TileLayer `json:"layers,omitempty"`
	MaxTileAge *int        `json:"max_tile_age,omitempty"`
}

type WMTSServiceProvider struct {
	ProviderName string `json:"providername"`
	ProviderSite struct {
		Type string `json:"type"`
		Href string `json:"href"`
	} `json:"providersite"`
	ServiceContact struct {
		IndividualName string `json:"individualname"`
		PositionName   string `json:"positionname"`
		ContactInfo    struct {
			Phone struct {
				Voice     string `json:"voice"`
				Facsimile string `json:"facsimile"`
			} `json:"phone"`
			Address struct {
				DeliveryPoint         string `json:"deliverypoint"`
				City                  string `json:"city"`
				AdministrativeArea    string `json:"administrativearea"`
				PostalCode            string `json:"postalcode"`
				Country               string `json:"country"`
				ElectronicMailAddress string `json:"electronicmailaddress"`
			} `json:"address"`
			OnlineResource *struct {
				Type string `json:"type"`
				Href string `json:"href"`
			} `json:"onlineresource,omitempty"`
			HoursOfService      string `json:"hoursofservice"`
			ContactInstructions string `json:"contactinstructions"`
		} `json:"contactinfo"`
		Role string `json:"role"`
	} `json:"servicecontact"`
}

type WMTSService struct {
	Type                       string               `json:"type,omitempty"`
	Restful                    *bool                `json:"restful,omitempty"`
	RestfulTemplate            string               `json:"restful_template,omitempty"`
	RestfulFeatureinfoTemplate string               `json:"restful_featureinfo_template,omitempty"`
	Title                      string               `json:"title,omitempty"`
	Abstract                   string               `json:"abstract,omitempty"`
	KeywordList                []string             `json:"keyword_list,omitempty"`
	Fees                       *string              `json:"fees,omitempty"`
	AccessConstraints          *string              `json:"access_constraints,omitempty"`
	Provider                   *WMTSServiceProvider `json:"provider,omitempty"`
	FeatureinfoFormats         []FeatureinfoFormat  `json:"featureinfo_formats,omitempty"`
	Layers                     []TileLayer          `json:"layers,omitempty"`
	MaxTileAge                 *int                 `json:"max_tile_age,omitempty"`
}

type BBoxSrs struct {
	Srs  string     `json:"srs"`
	BBox [4]float64 `json:"bbox,omitempty"`
}

type FeatureinfoFormat struct {
	MimeType string `json:"mimetype"`
	Suffix   string `json:"suffix,omitempty"`
}

type WMSExtendedCapabilities struct {
	MetadataURL struct {
		URL       string `json:"url,omitempty"`
		MediaType string `json:"mediatype,omitempty"`
	} `json:"metadataurl,omitempty"`
	SupportedLanguages struct {
		DefaultLanguage struct {
			Language string `json:"language,omitempty"`
		} `json:"defaultlanguage,omitempty"`
		SupportedLanguage *[]struct {
			Language string `json:"language,omitempty"`
		} `json:"supportedlanguage,omitempty"`
	} `json:"supportedlanguages,omitempty"`
	ResponseLanguage struct {
		Language string `json:"language,omitempty"`
	} `json:"responselanguage,omitempty"`
}

type WMSContactInformation struct {
	ContactPersonPrimary struct {
		ContactPerson       string `json:"contactperson,omitempty"`
		ContactOrganization string `json:"contactorganization,omitempty"`
	} `json:"contactpersonprimary,omitempty"`
	ContactPosition string `json:"contactposition,omitempty"`
	ContactAddress  struct {
		AddressType     string `json:"addresstype,omitempty"`
		Address         string `json:"address,omitempty"`
		City            string `json:"city,omitempty"`
		StateOrProvince string `json:"stateorprovince,omitempty"`
		PostCode        string `json:"postalcode,omitempty"`
		Country         string `json:"country,omitempty"`
	} `json:"contactaddress,omitempty"`
	ContactVoiceTelephone        string `json:"contactvoicetelephone,omitempty"`
	ContactFacsimileTelephone    string `json:"contactfacsimiletelephone,omitempty"`
	ContactElectronicMailAddress string `json:"contactelectronicmailaddress,omitempty"`
}

type WMSService struct {
	Type               string              `json:"type,omitempty"`
	Srs                []string            `json:"srs,omitempty"`
	BBoxSrs            []BBoxSrs           `json:"bbox_srs,omitempty"`
	ImageFormats       []string            `json:"image_formats,omitempty"`
	FeatureinfoFormats []FeatureinfoFormat `json:"featureinfo_formats,omitempty"`
	FeatureinfoXslt    map[string]string   `json:"featureinfo_xslt,omitempty"`
	MaxOutputPixels    *int                `json:"max_output_pixels,omitempty"`
	Strict             *bool               `json:"strict,omitempty"`
	Title              string              `json:"title,omitempty"`
	Abstract           string              `json:"abstract,omitempty"`
	KeywordList        []string            `json:"keyword_list,omitempty"`
	OnlineResource     struct {
		Xlink *string `json:"xlink,omitempty"`
		Type  *string `json:"type,omitempty"`
		Href  *string `json:"href,omitempty"`
	} `json:"online_resource"`
	Fees                 *string                  `json:"fees,omitempty"`
	AccessConstraints    *string                  `json:"access_constraints,omitempty"`
	RootLayer            *WMSLayer                `json:"root_layer,omitempty"`
	Layers               []WMSLayer               `json:"layers,omitempty"`
	MaxTileAge           *int                     `json:"max_tile_age,omitempty"`
	ExtendedCapabilities *WMSExtendedCapabilities `json:"extended_capabilities,omitempty"`
	ContactInformation   *WMSContactInformation   `json:"contact_information,omitempty"`
}

type WMSKeywords struct {
	Keyword []string `json:"keyword,omitempty"`
}

type WMSOnlineResource struct {
	Xlink *string `json:"xlink,omitempty"`
	Type  *string `json:"type,omitempty"`
	Href  *string `json:"href,omitempty"`
}

type WMSAuthorityURL struct {
	Name           string            `json:"name,omitempty"`
	OnlineResource WMSOnlineResource `json:"onlineresource,omitempty"`
}

type WMSIdentifier struct {
	Authority string `json:"authority,omitempty"`
	Value     string `json:"value,omitempty"`
}

type WMSMetadataURL struct {
	Type           *string           `json:"type,omitempty"`
	Format         *string           `json:"format,omitempty"`
	OnlineResource WMSOnlineResource `json:"onlineresource,omitempty"`
}

type WMSStyle struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
	LegendURL struct {
		Width          int               `json:"width"`
		Height         int               `json:"height"`
		Format         string            `json:"format"`
		OnlineResource WMSOnlineResource `json:"onlineresource"`
	} `json:"legendurl"`
	StyleSheetURL *struct {
		Format         string            `json:"format"`
		OnlineResource WMSOnlineResource `json:"onlineresource"`
	} `json:"stylesheeturl"`
}

type WMSLayerMetadata struct {
	Abstract     string            `json:"abstract"`
	KeywordList  *WMSKeywords      `json:"keyword_list,omitempty"`
	AuthorityURL *WMSAuthorityURL  `json:"authority_url,omitempty"`
	Identifier   *WMSIdentifier    `json:"identifier,omitempty"`
	MetadataURL  []*WMSMetadataURL `json:"metadata_url,omitempty"`
	Style        []*WMSStyle       `json:"style,omitempty"`
}

type WMSLayer struct {
	ScaleHints
	Sources            []string                 `json:"sources"`
	FeatureinfoSources []string                 `json:"featureinfo_sources,omitempty"`
	LegendSources      []string                 `json:"legend_sources,omitempty"`
	Name               string                   `json:"name,omitempty"`
	Title              string                   `json:"title"`
	Description        string                   `json:"description,omitempty"`
	LegendURL          string                   `json:"legend_url,omitempty"`
	Metadata           *WMSLayerMetadata        `json:"metadata,omitempty"`
	Layers             []WMSLayer               `json:"layers,omitempty"`
	Dimensions         map[string][]interface{} `json:"dimensions,omitempty"`
	LegendStore        interface{}              `json:"legendstore"`
}

type StoreInfo struct {
	Type      CacheType `json:"type,omitempty"`
	Directory string    `json:"directory,omitempty"`
}

func DefaultStoreInfo() *StoreInfo {
	return &StoreInfo{
		Type:      CACHE_TYPE_FILE,
		Directory: "./cache",
	}
}

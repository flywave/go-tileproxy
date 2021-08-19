package setting

import "time"

type Coverage struct {
	Polygons    *string
	PolygonsSrs *string
	BBox        *[4]float64
	BBoxSrs     *string
	Geometry    *string
	GeometrySrs *string
	ExpireTiles [][3]int
	Clip        bool
}

type Union []Coverage

type Intersection []Coverage

type Difference []Coverage

type ImageOpts struct {
	Mode             string
	Colors           uint32
	Transparent      bool
	ResamplingMethod string
	Format           string
	EncodingOptions  map[string]string
}

type GridOpts struct {
	Base                 string
	Name                 *string
	Srs                  *string
	BBox                 *[4]float64
	BBoxSrs              *string
	NumLevels            int
	Resolutions          []float64
	ResFactor            interface{}
	MaxRes               float64
	MinRes               float64
	StretchFactor        float64
	MaxStretchFactor     float64
	AlignResolutionsWith string
	Origin               string
	TileSize             [2]uint32
	ThresholdRes         []float64
}

type ScaleHints struct {
	MaxScale float64
	MinScale float64
	MaxRes   float64
	MinRes   float64
}

type ImageSetting struct {
	ResamplingMethod string
	Paletted         bool
	StretchFactor    float64
	MaxStretchFactor float64
	JpegQuality      float64
	FontDir          string
	Formats          map[string]ImageOpts
}

type Srs struct {
	AxisOrderNE      []string
	AxisOrderEN      []string
	ProjDataDir      string
	PreferredSrcProj map[string][]string
}

type Cache struct {
	Name                   string
	Grids                  []string
	LockDir                string
	CacheDir               string
	MetaSize               []int
	MetaBuffer             int
	BulkMetaTiles          bool
	Options                interface{}
	MaxTileLimit           int
	MinimizeMetaRequests   bool
	ConcurrentTileCreators int
	UseDirectFromLevel     int
	UseDirectFromRes       float64
	DisableStorage         bool
	Format                 string
	RequestFormat          string
	CacheRescaledTiles     bool
	UpscaleTiles           int
	DownscaleTiles         int
	WaterMark              *WaterMark
	CacheInfo              interface{}
}

type HttpOpts struct {
	UserAgent         string
	RandomDelay       int
	DisableKeepAlives bool
	Proxys            []string
	RequestTimeout    time.Duration
}

type WMSSourceOpts struct {
	Version              string
	Map                  bool
	FeatureInfo          bool
	LegendGraphic        bool
	LegendURL            string
	FeatureinfoFormat    string
	FeatureinfoXslt      string
	FeatureinfoOutFormat string
}

type SourceCommons struct {
	ScaleHints
	ConcurrentRequests int
	Coverage           *Coverage
	SeedOnly           bool
}

type WMSSource struct {
	SourceCommons
	Opts  WMSSourceOpts
	Image struct {
		ImageOpts
		Opacity                   float64
		TransparentColor          [4]uint8
		TransparentColorTolerance float64
	}
	ForwardReqParams map[string]string
	SupportedFormats []string
	SupportedSrs     []string
	Http             HttpOpts
	Request          struct {
		Url         string
		Layers      []string
		Transparent bool
	}
}

type TileSource struct {
	SourceCommons
	Url           string
	Transparent   bool
	Image         ImageOpts
	Grid          string
	RequestFormat string
	Origin        string
	Http          HttpOpts
}

type ArcgisSource struct {
	SourceCommons
	Request struct {
		Url         string
		Dpi         int
		Layers      []string
		Transparent bool
		Time        string
	}
	Opts struct {
		Featureinfo                 bool
		FeatureinfoTolerance        float64
		FeatureinfoReturnGeometries bool
	}
	SupportedSrs []string
	Http         HttpOpts
}

type WaterMark struct {
	Text     string
	FontSize int
	Color    [4]uint8
	Opacity  float64
	Spacing  string
}

type MapboxService struct {
}

type TMSService struct {
	UseGridNames bool
	Origin       string
}

type WMTSService struct {
	KVP                        bool
	Restful                    bool
	RestfulTemplate            string
	RestfulFeatureinfoTemplate string
	Metadata                   map[string]string
	FeatureinfoFormats         struct {
		MimeType string
		Suffix   string
	}
}

type WMSService struct {
	Srs              []string
	BBox             [4]float64
	BBoxSrs          string
	ImageFormats     []string
	FeatureinfoTypes []string
	FeatureinfoXslt  map[string]string
	MaxOutputPixels  int
	Strict           bool
	Metadata         map[string]string
	InspireMetadata  map[string]string
}

type Dimension struct {
	Default string
	Values  []string
}

type TileLayer struct {
	ScaleHints
	Sources     []string
	TileSources []string
	Name        string
	Title       string
	LegendURL   string
	Metadata    map[string]string
	Dimensions  map[string]Dimension
}

type WMSLayer struct {
	ScaleHints
	MapSources         []string
	FeatureinfoSources []string
	LegendSources      []string
	Name               string
	Title              string
	LegendURL          string
	Metadata           map[string]string
	Layers             []WMSLayer
	Dimensions         map[string]Dimension
}

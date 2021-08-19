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

type ImageSetting struct {
	ResamplingMethod string
	Paletted         bool
	StretchFactor    float64
	MaxStretchFactor float64
	JpegQuality      float64
	FontDir          string
	Formats          map[string]ImageOpts
}

type SRS struct {
	AxisOrderNE      []string
	AxisOrderEN      []string
	ProjDataDir      string
	PreferredSrcProj map[string][]string
}

type HttpCollector struct {
	UserAgent         string
	RandomDelay       int
	DisableKeepAlives bool
	Proxys            []string
	RequestTimeout    time.Duration
}

type WMSSourceOpts struct {
	Version       string
	Map           bool
	FeatureInfo   bool
	LegendGraphic bool
	LegendURL     string
}

type WMSSource struct {
	Type             string
	Opts             WMSSourceOpts
	ForwardReqParams map[string]string
	SupportedFormats []string
	SupportedSrs     []string
	Request          struct {
		Url         string
		Layers      []string
		Transparent bool
	}
	Coverage *Coverage
}

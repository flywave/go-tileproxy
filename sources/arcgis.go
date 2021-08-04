package sources

import (
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type ArcGISSource struct {
	WMSSource
}

func NewArcGISSource(client *client.WMSClient, image_opts *images.ImageOptions, coverage geo.Coverage, res_range *geo.ResolutionRange, supported_srs *geo.SupportedSRS, supported_formats []string, error_handler HTTPSourceErrorHandler) *ArcGISSource {
	return &ArcGISSource{WMSSource: WMSSource{Client: client, MapLayer: layer.MapLayer{Options: image_opts, Coverage: coverage, ResRange: res_range, SupportMetaTiles: false}, TransparentColor: nil, TransparentColorTolerance: nil, SupportedSRS: supported_srs, SupportedFormats: supported_formats, ExtReqParams: nil, ErrorHandler: error_handler}}
}

type ArcGISInfoSource struct {
	WMSInfoSource
}

func NewArcGISInfoSource(c client.Client) *ArcGISInfoSource {
	return &ArcGISInfoSource{WMSInfoSource: WMSInfoSource{Client: c.(*client.WMSInfoClient)}}
}

func (c *ArcGISInfoSource) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
	agis := c.GetClient().(*client.ArcGISInfoClient)
	doc := agis.GetInfo(query)
	return doc
}

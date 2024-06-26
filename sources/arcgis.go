package sources

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type ArcGISSource struct {
	WMSSource
}

func NewArcGISSource(
	client *client.ArcGISClient,
	image_opts *imagery.ImageOptions,
	coverage geo.Coverage,
	res_range *geo.ResolutionRange,
	supported_srs *geo.SupportedSRS,
	supported_formats []string) *ArcGISSource {
	return &ArcGISSource{
		WMSSource: WMSSource{
			Client: client,
			MapLayer: layer.MapLayer{
				Options:          image_opts,
				Coverage:         coverage,
				ResRange:         res_range,
				SupportMetaTiles: false,
			},
			TransparentColor:          nil,
			TransparentColorTolerance: nil,
			SupportedSRS:              supported_srs,
			SupportedFormats:          supported_formats,
			ExtReqParams:              nil,
		},
	}
}

type ArcGISInfoSource struct {
	Client *client.ArcGISInfoClient
}

func (s *ArcGISInfoSource) GetClient() *client.ArcGISInfoClient {
	return s.Client
}

func (c *ArcGISInfoSource) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
	agis := c.GetClient()
	doc := agis.GetInfo(query)
	return doc
}

func NewArcGISInfoSource(c *client.ArcGISInfoClient) *ArcGISInfoSource {
	return &ArcGISInfoSource{Client: c}
}

package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
)

type ArcGISClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
}

func (c *ArcGISClient) Retrieve(*layer.MapQuery, *images.ImageFormat) []byte {
	return nil
}

func (c *ArcGISClient) queryURL(query *layer.MapQuery, format *images.ImageFormat) string {
	req := c.RequestTemplate
	params := request.NewArcGISExportRequestParams(req.GetParams())
	params.SetFormat(*format)
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetBBOxSrs(query.Srs.GetDef())
	params.SetImageSrs(query.Srs.GetDef())
	params.SetTransparent(query.Transparent)
	return req.CompleteUrl()
}

type ArcGISInfoClient struct {
	WMSInfoClient
}

func (c *ArcGISInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	return nil
}

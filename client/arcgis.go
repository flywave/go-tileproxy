package client

import (
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
)

type ArcGISClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
}

func NewArcGISClient() *ArcGISClient {
	ret := &ArcGISClient{}

	return ret
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
	ReturnGeometries bool
	Tolerance        int
}

func NewArcGISInfoClient() *ArcGISInfoClient {
	ret := &ArcGISInfoClient{}

	return ret
}

func (c *ArcGISInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	b, _ := geo.ContainsSrs(query.Srs.GetDef(), c.SupportedSrs)
	if c.SupportedSrs != nil && !b {
		query = c.GetTransformedQuery(query)
	}
	resp := c.retrieve(query)
	return resource.CreateFeatureinfoDoc(resp, query.InfoFormat)
}

func (c *ArcGISInfoClient) queryURL(query *layer.InfoQuery) string {
	req := c.RequestTemplate
	params := request.NewArcGISIdentifyRequestParams(req.GetParams())
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetPos(query.Pos)
	params.SetSrs(query.Srs.GetDef())

	if strings.HasPrefix(query.InfoFormat, "text/html") {
		req.GetParams().Set("f", []string{"html"})
	} else {
		req.GetParams().Set("f", []string{"json"})
	}

	req.GetParams().Set("tolerance", []string{strconv.FormatInt(int64(c.Tolerance), 10)})
	if c.ReturnGeometries {
		req.GetParams().Set("returnGeometry", []string{"true"})
	} else {
		req.GetParams().Set("returnGeometry", []string{"false"})
	}

	return req.CompleteUrl()
}

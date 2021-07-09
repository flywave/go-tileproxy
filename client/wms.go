package client

import (
	"strings"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
)

type WMSClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
	HttpMethod      string
	FWDReqParams    map[string]string
}

func (c *WMSClient) Retrieve(query *layer.MapQuery, format *images.ImageFormat) []byte {
	var request_method string
	if c.HttpMethod == "POST" {
		request_method = "POST"
	} else if c.HttpMethod == "GET" {
		request_method = "GET"
	} else {
		if _, ok := c.RequestTemplate.GetParams().Get("sld_body"); ok {
			request_method = "POST"
		} else {
			request_method = "GET"
		}
	}
	var url string
	var data []byte
	if request_method == "POST" {
		url, data = c.queryData(query, format)
	} else {
		url = c.queryURL(query, format)
		data = nil
	}
	var resp *crawler.Response
	resp = c.Client.Open(url, data)
	return resp.Body
}

func (c *WMSClient) queryData(query *layer.MapQuery, format *images.ImageFormat) (url string, data []byte) {
	req := c.queryReq(query, format)
	ind := strings.Index(req.Url, "?")
	if ind != -1 {
		url = req.Url[ind:]
	} else {
		url = req.Url[:]
	}
	return url, []byte(req.QueryString())
}

func (c *WMSClient) queryURL(query *layer.MapQuery, format *images.ImageFormat) string {
	return c.queryReq(query, format).CompleteUrl()
}

func (c *WMSClient) queryReq(query *layer.MapQuery, format *images.ImageFormat) *request.ArcGISRequest {
	req := c.RequestTemplate
	params := request.NewWMTSTileRequestParams(req.GetParams())
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetSrs(query.Srs.GetDef())
	params.SetFormat(*format)
	params.Update(query.DimensionsForParams(c.FWDReqParams))
	return req
}

func (c *WMSClient) CombinedClient(other *WMSClient, query *layer.MapQuery) *WMSClient {
	if c.RequestTemplate.Url != other.RequestTemplate.Url {
		return nil
	}

	new_req := *c.RequestTemplate
	params := request.NewWMTSTileRequestParams(new_req.GetParams())
	other_params := request.NewWMTSTileRequestParams(other.RequestTemplate.Params)

	params.SetLayer(params.GetLayer() + other_params.GetLayer())

	return &WMSClient{RequestTemplate: &new_req, Client: c.Client, HttpMethod: c.HttpMethod, FWDReqParams: c.FWDReqParams}
}

type WMSInfoClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
	SupportedSrs    []geo.Proj
}

func (c *WMSInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	return nil
}

func (c *WMSInfoClient) GetTransformedQuery(query *layer.InfoQuery) *layer.InfoQuery {
	return nil
}

func (c *WMSInfoClient) retrieve(query *layer.InfoQuery) []byte {
	return nil
}

func (c *WMSInfoClient) queryURL(query *layer.InfoQuery) string {
	return ""
}

type WMSLegendClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
}

func (c *WMSLegendClient) GetLegend(query *layer.LegendQuery) *resource.Legend {
	return nil
}

func (c *WMSLegendClient) retrieve(query *layer.LegendQuery) {

}

func (c *WMSLegendClient) queryURL(query *layer.LegendQuery) string {
	return ""
}

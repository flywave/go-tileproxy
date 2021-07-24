package client

import (
	"bytes"
	"math"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type WMSClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
	HttpMethod      string
	FWDReqParams    map[string]string
}

func (c *WMSClient) Retrieve(query *layer.MapQuery, format *tile.TileFormat) []byte {
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

func (c *WMSClient) queryData(query *layer.MapQuery, format *tile.TileFormat) (url string, data []byte) {
	req := c.queryReq(query, format)
	ind := strings.Index(req.Url, "?")
	if ind != -1 {
		url = req.Url[ind:]
	} else {
		url = req.Url[:]
	}
	return url, []byte(req.QueryString())
}

func (c *WMSClient) queryURL(query *layer.MapQuery, format *tile.TileFormat) string {
	return c.queryReq(query, format).CompleteUrl()
}

func (c *WMSClient) queryReq(query *layer.MapQuery, format *tile.TileFormat) *request.ArcGISRequest {
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
	SupportedSrs    *geo.SupportedSRS
}

func (c *WMSInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	b, _ := geo.ContainsSrs(query.Srs.GetDef(), c.SupportedSrs.Srs)
	if c.SupportedSrs != nil && !b {
		query = c.GetTransformedQuery(query)
	}
	resp := c.retrieve(query)
	var info_format string

	ifs, ok := c.RequestTemplate.Params["info_format"]
	if !ok {
		info_format = c.RequestTemplate.Params.GetOne("Content-type", "")
	} else {
		info_format = ifs[0]
	}
	if info_format == "" {
		info_format = query.InfoFormat
	}

	return resource.CreateFeatureinfoDoc(resp, info_format)
}

func (c *WMSInfoClient) GetTransformedQuery(query *layer.InfoQuery) *layer.InfoQuery {
	req_srs := query.Srs
	req_bbox := query.BBox
	req_coord := geo.MakeLinTransf(vec2d.Rect{Min: vec2d.T{float64(0), float64(0)}, Max: vec2d.T{float64(query.Size[0]), float64(query.Size[1])}}, req_bbox)(query.Pos[:])

	info_srs, _ := c.SupportedSrs.BestSrs(req_srs)
	info_bbox := req_srs.TransformRectTo(info_srs, req_bbox, 16)

	info_aratio := (info_bbox.Max[1] - info_bbox.Min[1]) / (info_bbox.Max[0] - info_bbox.Min[0])
	info_size := [2]uint32{query.Size[0], uint32(info_aratio * float64(query.Size[0]))}

	info_coord := req_srs.TransformTo(info_srs, []vec2d.T{{req_coord[0], req_coord[1]}})
	info_pos := geo.MakeLinTransf(info_bbox, vec2d.Rect{Min: vec2d.T{float64(0), float64(0)}, Max: vec2d.T{float64(info_size[0]), float64(info_size[1])}})(info_coord[0][:])
	info_pos2 := [2]float64{math.Round(info_pos[0]), math.Round(info_pos[1])}

	info_query := &layer.InfoQuery{
		BBox:         info_bbox,
		Size:         info_size,
		Srs:          info_srs,
		Pos:          info_pos2,
		InfoFormat:   query.InfoFormat,
		FeatureCount: query.FeatureCount,
	}
	return info_query
}

func (c *WMSInfoClient) retrieve(query *layer.InfoQuery) []byte {
	url := c.queryURL(query)
	return c.Client.Get(url).Body
}

func (c *WMSInfoClient) queryURL(query *layer.InfoQuery) string {
	req := c.RequestTemplate
	params := request.NewWMTSFeatureInfoRequestParams(req.GetParams())
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetPos(query.Pos)
	params.SetSrs(query.Srs.GetDef())

	if query.FeatureCount > 0 {
		fc := strconv.FormatInt(int64(query.FeatureCount), 10)
		req.GetParams().Set("feature_count", []string{fc})
	}
	req.GetParams().Set("query_layers", []string{req.GetParams().GetOne("layers", "")})
	if _, ok := req.GetParams()["info_format"]; !ok && query.InfoFormat != "" {
		req.GetParams().Set("info_format", []string{query.InfoFormat})
	}

	if query.Format != "" {
		params.SetFormat(tile.TileFormat(query.Format))
	} else {
		params.SetFormat("image/png")
	}

	return req.CompleteUrl()
}

type WMSLegendClient struct {
	Client
	RequestTemplate *request.ArcGISRequest
}

func (c *WMSLegendClient) GetLegend(query *layer.LegendQuery) *resource.Legend {
	resp := c.retrieve(query)
	format := request.SplitMimeType(query.Format)[1]

	src := &images.ImageSource{Options: &images.ImageOptions{Format: tile.TileFormat(format)}}
	src.SetSource(bytes.NewBuffer(resp))

	return &resource.Legend{Source: src, Scale: query.Scale}
}

func (c *WMSLegendClient) retrieve(query *layer.LegendQuery) []byte {
	url := c.queryURL(query)
	return c.Client.Get(url).Body
}

func (c *WMSLegendClient) queryURL(query *layer.LegendQuery) string {
	req := *c.RequestTemplate
	params := request.NewWMTSLegendRequestParams(req.GetParams())

	if query.Format != "" {
		params.SetFormat(tile.TileFormat(query.Format))
	} else {
		params.SetFormat("image/png")
	}
	if query.Scale == -1 {
		params.SetScale(query.Scale)
	}
	return req.CompleteUrl()
}

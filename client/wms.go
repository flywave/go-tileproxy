package client

import (
	"bytes"
	"math"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type WMSClient struct {
	BaseClient
	RequestTemplate *request.WMSMapRequest
	HttpMethod      string
	FWDReqParams    map[string]string
	AdaptTo111      bool
	AccessToken     *string
	AccessTokenName *string
}

func NewWMSClient(req *request.WMSMapRequest, accessToken *string, accessTokenName *string, ctx Context) *WMSClient {
	return &WMSClient{RequestTemplate: req, BaseClient: BaseClient{ctx: ctx}, AccessToken: accessToken, AccessTokenName: accessTokenName}
}

func (c *WMSClient) Retrieve(query *layer.MapQuery, format *tile.TileFormat) []byte {
	var requestMethod string
	switch c.HttpMethod {
	case "POST":
		requestMethod = "POST"
	case "GET":
		requestMethod = "GET"
	default:
		if _, ok := c.RequestTemplate.GetParams().Get("sld_body"); ok {
			requestMethod = "POST"
		} else {
			requestMethod = "GET"
		}
	}

	var url string
	var data []byte
	if requestMethod == "POST" {
		url, data = c.queryData(query, format)
	} else {
		url = c.queryURL(query, format)
		data = nil
	}
	status, resp := c.httpClient().Open(url, data, nil)
	if status == 200 {
		return resp
	}
	return nil
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
	req := c.queryReq(query, format)
	if c.AccessToken != nil {
		if c.AccessTokenName != nil {
			req.GetParams().Set(*c.AccessTokenName, []string{*c.AccessToken})
		} else {
			req.GetParams().Set("access_token", []string{*c.AccessToken})
		}
	}
	return req.CompleteUrl()
}

func (c *WMSClient) queryReq(query *layer.MapQuery, format *tile.TileFormat) *request.WMSMapRequest {
	req := *c.RequestTemplate
	params := request.NewWMSMapRequestParams(req.GetParams())
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetCrs(query.Srs.GetSrsCode())
	params.SetFormat(*format)
	params.Update(query.DimensionsForParams(c.FWDReqParams))
	if c.AdaptTo111 {
		req.AdaptToWMS111()
	}
	return &req
}

func (c *WMSClient) CombinedClient(other MapClient, query *layer.MapQuery) MapClient {
	oc := other.(*WMSClient)
	if c.RequestTemplate.Url != oc.RequestTemplate.Url {
		return nil
	}

	new_req := *c.RequestTemplate
	params := request.NewWMSMapRequestParams(new_req.GetParams())
	other_params := request.NewWMSMapRequestParams(oc.RequestTemplate.Params)

	layers := params.GetLayers()
	layers = append(layers, other_params.GetLayers()...)
	params.AddLayers(layers)

	return &WMSClient{RequestTemplate: &new_req, BaseClient: c.BaseClient, HttpMethod: c.HttpMethod, FWDReqParams: c.FWDReqParams}
}

type WMSInfoClient struct {
	BaseClient
	RequestTemplate *request.WMSFeatureInfoRequest
	SupportedSrs    *geo.SupportedSRS
	AdaptTo111      bool
	AccessToken     *string
	AccessTokenName *string
}

func NewWMSInfoClient(req *request.WMSFeatureInfoRequest, supported_srs *geo.SupportedSRS, accessToken *string, accessTokenName *string, ctx Context) *WMSInfoClient {
	return &WMSInfoClient{RequestTemplate: req, SupportedSrs: supported_srs, BaseClient: BaseClient{ctx: ctx}, AccessToken: accessToken, AccessTokenName: accessTokenName}
}

func (c *WMSInfoClient) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
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
	status, resp := c.httpClient().Open(url, nil, nil)
	if status == 200 {
		return resp
	}
	return nil
}

func (c *WMSInfoClient) queryURL(query *layer.InfoQuery) string {
	req := c.RequestTemplate
	params := request.NewWMSFeatureInfoRequestParams(req.GetParams())
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetPos(query.Pos)
	params.SetCrs(query.Srs.GetDef())

	if query.FeatureCount != nil {
		fc := strconv.FormatInt(int64(*query.FeatureCount), 10)
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

	if c.AdaptTo111 {
		req.AdaptToWMS111()
	}

	if c.AccessToken != nil {
		if c.AccessTokenName != nil {
			req.GetParams().Set(*c.AccessTokenName, []string{*c.AccessToken})
		} else {
			req.GetParams().Set("access_token", []string{*c.AccessToken})
		}
	}
	return req.CompleteUrl()
}

type WMSLegendClient struct {
	BaseClient
	RequestTemplate *request.WMSLegendGraphicRequest
	AccessToken     *string
	AccessTokenName *string
}

func NewWMSLegendClient(req *request.WMSLegendGraphicRequest, accessToken *string, accessTokenName *string, ctx Context) *WMSLegendClient {
	return &WMSLegendClient{RequestTemplate: req, BaseClient: BaseClient{ctx: ctx}, AccessToken: accessToken, AccessTokenName: accessTokenName}
}

func (c *WMSLegendClient) GetLegend(query *layer.LegendQuery) *resource.Legend {
	resp := c.retrieve(query)
	format := request.SplitMimeType(query.Format)[1]

	src := &imagery.ImageSource{Options: &imagery.ImageOptions{Format: tile.TileFormat(format)}}
	src.SetSource(bytes.NewBuffer(resp))

	id := c.RequestTemplate.Url

	return &resource.Legend{BaseResource: resource.BaseResource{StoreID: id}, Source: src, Scale: query.Scale}
}

func (c *WMSLegendClient) retrieve(query *layer.LegendQuery) []byte {
	url := c.queryURL(query)
	states, resp := c.httpClient().Open(url, nil, nil)
	if states == 200 {
		return resp
	}
	return nil
}

func (c *WMSLegendClient) queryURL(query *layer.LegendQuery) string {
	req := *c.RequestTemplate
	params := request.NewWMSLegendGraphicRequestParams(req.GetParams())

	if query.Format != "" {
		params.SetFormat(tile.TileFormat(query.Format))
	} else {
		params.SetFormat("image/png")
	}
	if query.Scale > 0 {
		params.SetScale(query.Scale)
	}

	if c.AccessToken != nil {
		if c.AccessTokenName != nil {
			req.GetParams().Set(*c.AccessTokenName, []string{*c.AccessToken})
		} else {
			req.GetParams().Set("access_token", []string{*c.AccessToken})
		}
	}
	return req.CompleteUrl()
}

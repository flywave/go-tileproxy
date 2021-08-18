package client

import (
	"math"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type ArcGISClient struct {
	BaseClient
	RequestTemplate *request.ArcGISRequest
}

func NewArcGISClient(req *request.ArcGISRequest, ctx Context) *ArcGISClient {
	ret := &ArcGISClient{RequestTemplate: req, BaseClient: BaseClient{ctx: ctx}}
	return ret
}

func (c *ArcGISClient) Retrieve(query *layer.MapQuery, format *tile.TileFormat) []byte {
	url := c.queryURL(query, format)
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}

func (c *ArcGISClient) CombinedClient(other MapClient, query *layer.MapQuery) MapClient {
	return nil
}

func (c *ArcGISClient) queryURL(query *layer.MapQuery, format *tile.TileFormat) string {
	req := c.RequestTemplate
	params := request.NewArcGISExportRequestParams(req.GetParams())
	params.SetFormat(*format)
	params.SetBBox(query.BBox)
	params.SetSize(query.Size)
	params.SetBBoxSrs(query.Srs.GetDef())
	params.SetImageSrs(query.Srs.GetDef())
	params.SetTransparent(query.Transparent)
	return req.CompleteUrl()
}

type ArcGISInfoClient struct {
	BaseClient
	RequestTemplate  *request.ArcGISIdentifyRequest
	SupportedSrs     *geo.SupportedSRS
	ReturnGeometries bool
	Tolerance        int
}

func NewArcGISInfoClient(req *request.ArcGISIdentifyRequest, supported_srs *geo.SupportedSRS, ctx Context, return_geometries bool, tolerance int) *ArcGISInfoClient {
	ret := &ArcGISInfoClient{BaseClient: BaseClient{ctx: ctx}, RequestTemplate: req, SupportedSrs: supported_srs, ReturnGeometries: return_geometries, Tolerance: tolerance}
	return ret
}

func (c *ArcGISInfoClient) GetTransformedQuery(query *layer.InfoQuery) *layer.InfoQuery {
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

func (c *ArcGISInfoClient) retrieve(query *layer.InfoQuery) []byte {
	url := c.queryURL(query)
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}

func (c *ArcGISInfoClient) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
	b, _ := geo.ContainsSrs(query.Srs.GetDef(), c.SupportedSrs.Srs)
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

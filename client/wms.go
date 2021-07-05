package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type WMSClient struct {
	Client
}

func (c *WMSClient) Retrieve(*layer.MapQuery, images.ImageFormat) []byte {
	return nil
}

func (c *WMSClient) CombinedClient(other Client, query *layer.MapQuery) Client {
	return nil
}

type WMSInfoClient struct {
	Client
}

func (c *WMSInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	return nil
}

type WMSLegendClient struct {
	Client
}

func (c *WMSLegendClient) GetLegend(query *layer.LegendQuery) *resource.Legend {
	return nil
}

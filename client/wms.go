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

func (c *WMSClient) GetInfo(query *layer.MapQuery) *resource.FeatureInfo {
	return nil
}

func (c *WMSClient) GetLegend(query *layer.MapQuery) *resource.Legend {
	return nil
}

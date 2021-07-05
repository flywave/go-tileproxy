package client

import (
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type ArcGISClient struct {
	Client
}

func (c *ArcGISClient) Retrieve(*layer.MapQuery, images.ImageFormat) []byte {
	return nil
}

type ArcGISInfoClient struct {
	WMSInfoClient
}

func (c *ArcGISInfoClient) GetInfo(query *layer.InfoQuery) *resource.FeatureInfo {
	return nil
}

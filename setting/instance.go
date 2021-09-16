package setting

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
)

type ProxyInstance interface {
	GetGrid(name string) geo.Grid
	GetSource(name string) layer.Layer
	GetCache(name string) cache.Manager
	GetCacheSource(name string) layer.Layer
	GetInfoSource(name string) layer.InfoLayer
	GetLegendSource(name string) layer.LegendLayer
}

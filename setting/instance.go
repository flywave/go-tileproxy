package setting

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type ProxyInstance interface {
	GetGrid(name string) geo.Grid
	GetSource(name string) layer.Layer
	GetCache(name string) cache.Manager
	GetCacheSource(name string, opt tile.TileOptions) layer.Layer
	GetInfoSource(name string) layer.InfoLayer
	GetLegendSource(name string) layer.LegendLayer
}

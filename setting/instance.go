package setting

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
)

type ProxyInstance interface {
	GetCache(name string) layer.Layer
	GetGrid(name string) geo.Grid
	GetCoverage(name string) geo.Coverage
	GetSource(name string) sources.Source
}
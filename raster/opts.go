package raster

import (
	"github.com/flywave/go-tileproxy/tile"
)

type RasterOptions struct {
	NoData          float64
	Format          tile.TileFormat
	EncodingOptions map[string]interface{}
}

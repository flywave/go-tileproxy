package layer

import "github.com/flywave/go-tileproxy/geo"

type TileManager interface {
	GetGrid() *geo.TileGrid
	GetMetaGrid() *geo.MetaGrid
	Cleanup() bool
}

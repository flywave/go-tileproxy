package layer

import "github.com/flywave/go-tileproxy/geo"

type TileManager interface {
	GetMetaGrid() *geo.MetaGrid
}

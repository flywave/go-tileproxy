package vector

import "github.com/flywave/go-tileproxy/tile"

type VectorOptions struct {
	Format tile.TileFormat
}

func (o *VectorOptions) GetFormat() tile.TileFormat {
	return o.Format
}

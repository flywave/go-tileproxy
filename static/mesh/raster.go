package mesh

import "github.com/flywave/go-tileproxy/terrain"

type RasterSource struct {
}

func (s *RasterSource) GetRaster() *terrain.TileData {
	return nil
}

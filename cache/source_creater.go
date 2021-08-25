package cache

import (
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

func GetSourceCreater(opts tile.TileOptions) tile.SourceCreater {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return &imagery.ImageSourceCreater{Opt: opt}
	case *terrain.RasterOptions:
		return &terrain.RasterSourceCreater{Opt: opt}
	case *vector.VectorOptions:
		return &vector.VectorSourceCreater{Opt: opt}
	}
	return nil
}

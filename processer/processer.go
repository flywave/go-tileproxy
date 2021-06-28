package processer

import "github.com/flywave/go-tileproxy/tile"

type TileProcesser interface {
	Process(p tile.Tile)
	Finish()
}

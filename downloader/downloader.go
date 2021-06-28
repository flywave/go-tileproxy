package downloader

import (
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type Downloader interface {
	Download(req *request.Request) tile.Tile
}

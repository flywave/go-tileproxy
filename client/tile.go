package client

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
)

type TileClient struct {
	Client
}

func (c *TileClient) GetTile(tile_coord [3]int, format *images.ImageFormat) []byte {
	return nil
}

func tilecachePath(tile_coord [3]int) string {
	return ""
}

func quadKey(tile_coord [3]int) string {
	return ""
}

func tmsPath(tile_coord [3]int) string {
	return ""
}

func arcgisCachePath(tile_coord [3]int) string {
	return ""
}

func bbox(tile_coord [3]int, grid geo.TileGrid) string {
	return ""
}

type TileURLTemplate struct{}

func (t *TileURLTemplate) substitute() {}

func (t *TileURLTemplate) ToString() string {
	return ""
}

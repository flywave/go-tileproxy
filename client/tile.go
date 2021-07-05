package client

import "github.com/flywave/go-tileproxy/images"

type TileClient struct {
	Client
}

func (c *TileClient) GetTile(tile_coord [3]int, format *images.ImageFormat) []byte {
	return nil
}

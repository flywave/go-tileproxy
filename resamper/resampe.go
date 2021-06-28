package resamper

import "github.com/flywave/go-tileproxy/tile"

type Resamper interface {
	Process(in []tile.Tile) tile.Tile
}

type TileResamper struct {
	Resamper
	tiles []tile.Tile
}

func (t *TileResamper) GetCollected() []tile.Tile {
	return t.tiles
}

func NewTileResamper() *TileResamper {
	return nil
}

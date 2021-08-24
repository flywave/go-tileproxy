package cache

import (
	"github.com/flywave/go-tileproxy/tile"
)

type TileCreater interface {
	Creater(data []byte, location string) tile.Source
}

type Cache interface {
	LoadTile(tile *Tile, withMetadata bool) error
	LoadTiles(tiles *TileCollection, withMetadata bool) error
	StoreTile(tile *Tile) error
	StoreTiles(tiles *TileCollection) error
	RemoveTile(tile *Tile) error
	RemoveTiles(tile *TileCollection) error
	IsCached(tile *Tile) bool
	LoadTileMetadata(tile *Tile) error
	LevelLocation(level int) string
}

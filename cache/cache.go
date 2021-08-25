package cache

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

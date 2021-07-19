package cache

type Cache interface {
	LoadTile(tile *Tile, withMetadata bool) error
	LoadTiles(tiles []*Tile, withMetadata bool) error
	StoreTile(tile *Tile) error
	StoreTiles(tiles []*Tile) error
	RemoveTile(tile *Tile) error
	RemoveTiles(tile []*Tile) error
	IsCached(tile *Tile) error
	LoadTileMetadata(tile *Tile) error
}

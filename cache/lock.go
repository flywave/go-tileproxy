package cache

type TileLocker interface {
	Lock(tile *Tile) error
}

package cache

type TileLocker interface {
	Lock(tile *Tile, run func() error) error
}

package cache

type Filter interface {
	Apply(*Tile) (*Tile, error)
}

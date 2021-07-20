package cache

type Filter interface {
	Apply(*Tile) *Tile
}

type ImageFilter interface {
	Filter
}

type RasterFilter interface {
	Filter
}

type VectorFilter interface {
	Filter
}

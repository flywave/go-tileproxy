package filters

type Filter interface{}

type ImageFilter interface {
	Filter
}

type RasterFilter interface {
	Filter
}

type VectorFilter interface {
	Filter
}

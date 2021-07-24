package tile

type Source interface {
	GetSource() interface{}
	SetSource(src interface{})
	GetFileName() string
	GetSize() [2]uint32
	GetBuffer(format *TileFormat, in_tile_opts TileOptions) []byte
	GetTile() interface{}
	GetCacheable() bool
	SetCacheable(c bool)
	SetTileOptions(options TileOptions)
	GetTileOptions() TileOptions
}

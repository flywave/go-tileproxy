package tile

type TileType uint8

const (
	TILE_IMAGERY   = 0
	TILE_DEMRASTER = 1
	TILE_VECTOR    = 2
)

type Source interface {
	GetType() TileType
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

package backend

type Source interface {
	MaxZoom() int
	MinZoom() int
	TileSize() int
	CacheSize() int
	CrossOrigin() string
	URLFormat() string
	Srs() (bool, string)
	ReProjection() (bool, string)
	Attributions() map[string]string
	Copyright() string
	MimeType() string
	Extension() string
}

type Layer interface {
	Name() string
	Title() string
	Type() string
	Path() string
	Source() Source
}

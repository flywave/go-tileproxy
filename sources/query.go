package sources

import (
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type Query interface{}

type MapQuery struct {
	Query
	BBox        vec2d.Rect
	Size        [2]uint32
	Srs         string
	Format      string
	Transparent bool
	TiledOnly   bool
}

type InfoQuery struct {
	Query
	BBox         vec2d.Rect
	Size         [2]uint32
	Srs          string
	Pos          [2]int
	InfoFormat   string
	Format       string
	FeatureCount int
}

type LegendQuery struct {
	Query
	Format string
	Scale  float32
}

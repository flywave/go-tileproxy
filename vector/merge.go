package vector

import (
	"github.com/flywave/go-geom"
	"github.com/flywave/go-mbgeom/vtile"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type FeatureBuilder struct {
	dst      [3]int
	coverage geo.Coverage
	Builder  *LayerBuilder
}

type LayerBuilder struct {
	name     string
	dst      [3]int
	coverage geo.Coverage
	result   []*geom.Feature
}

type TileBuilder struct {
	dst      [3]int
	coverage geo.Coverage
	result   map[string][]*geom.Feature
}

type VectorMerger struct {
	tile.Merger
	Layers    []tile.Source
	Coverages []geo.Coverage
	Cacheable *tile.CacheInfo
}

func (l *VectorMerger) getTiles() []*vtile.TileObject {
	return nil
}

func (l *VectorMerger) AddSource(src tile.Source, cov geo.Coverage) {

}

func (l *VectorMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	return nil
}

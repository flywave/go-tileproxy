package vector

import (
	"github.com/flywave/go-mbgeom/vtile"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type PBFMerger struct {
	tile.Merger
	Layers    []tile.Source
	Coverages []geo.Coverage
	Cacheable *tile.CacheInfo
	tileBaton *vtile.TileBaton
}

func (l *PBFMerger) getTiles() []*vtile.TileObject {
	return nil
}

func (l *PBFMerger) AddSource(src tile.Source, cov geo.Coverage) {

}

func (l *PBFMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	tiles := l.getTiles()
	l.tileBaton = vtile.NewTileBaton(len(l.Layers))
	for i := range tiles {
		l.tileBaton.AddTile(tiles[i])
	}
	vtile.Composite(l.tileBaton)
	return nil
}

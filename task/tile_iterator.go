package task

import (
	"github.com/flywave/go-geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type TileIterator struct {
	Subtiles    [][3]int
	AllSubtiles bool
	current     int
	grid        *geo.MetaGrid
	task        Task
}

func (it *TileIterator) HasNext() bool {
	return it.current < len(it.Subtiles)
}

type IntersectionType int

const (
	NONE       IntersectionType = -1
	CONTAINS   IntersectionType = 1
	INTERSECTS IntersectionType = 2
)

func (it *TileIterator) Next() (index int, rsubtile []int, rsub_bbox *vec2d.Rect, intersection IntersectionType) {
	subtile := it.Subtiles[it.current]
	defer func() {
		it.current++
	}()

	if subtile[0] == -1 || subtile[1] == -1 || subtile[2] == -1 {
		return it.current, nil, nil, NONE
	} else {
		metatile := it.grid.GetMetaTile(subtile)
		sub_bbox := metatile.GetBBox()
		if it.AllSubtiles {
			intersection = CONTAINS
		} else {
			intersection = it.task.Intersects(sub_bbox)
		}
		if intersection != NONE {
			index = it.current
			rsubtile = subtile[:]
			rsub_bbox = &sub_bbox
			return
		} else {
			return it.current, nil, nil, NONE
		}
	}
}

func (t *TileWalker) filterSubtiles(subtiles [][3]int, all_subtiles bool) *TileIterator {
	return &TileIterator{
		Subtiles:    subtiles,
		AllSubtiles: all_subtiles,
		current:     0,
		grid:        t.grid,
		task:        t.task,
	}
}

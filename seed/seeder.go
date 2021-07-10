package seed

import (
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type TileWalker struct {
	Ctx Context
}

func (t *TileWalker) Walk() {}

func (t *TileWalker) walk(cur_bbox vec2d.Rect, levels []int, current_level int, all_subtiles bool) {}

func (t *TileWalker) reportProgress(level int, bbox vec2d.Rect) {}

func (t *TileWalker) filterSubtiles(subtiles int, all_subtiles []int) {}

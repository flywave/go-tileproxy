package seed

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type TileWalker struct {
	Ctx                    Context
	manager                layer.TileManager
	task                   Task
	workOnMetatiles        bool
	handleStale            bool
	handleUncached         bool
	skipGeomsForLastLevels int
	reportTillLevel        int
	tilesPerMetatile       int
	seedProgress           *SeedProgress
	grid                   geo.MetaGrid
	count                  int
	seededTiles            []int
}

func NewTileWalker(task Task, ctx Context, handle_stale bool, handle_uncached bool,
	work_on_metatiles bool, skip_geoms_for_last_levels int,
	seed_progress *SeedProgress) *TileWalker {
	ret := &TileWalker{Ctx: ctx, task: task, manager: task.GetManager(), handleStale: handle_stale,
		handleUncached: handle_uncached, workOnMetatiles: work_on_metatiles,
		skipGeomsForLastLevels: skip_geoms_for_last_levels, seedProgress: seed_progress}

	num_seed_levels := len(task.GetLevels())
	if num_seed_levels >= 4 {
		ret.reportTillLevel = task.GetLevels()[num_seed_levels-2]
	} else {
		ret.reportTillLevel = task.GetLevels()[num_seed_levels-1]
	}
	var meta_size [2]uint32
	if ret.manager.GetMetaGrid() == nil {
		meta_size = [2]uint32{1, 1}
	} else {
		meta_size = ret.manager.GetMetaGrid().MetaSize
	}
	ret.tilesPerMetatile = int(meta_size[0] * meta_size[1])
	//ret.grid = &MetaGrid{self.tile_mgr.grid, meta_size=meta_size, meta_buffer=0}
	ret.count = 0
	if seed_progress != nil {
		ret.seedProgress = seed_progress
	} else {
		ret.seedProgress = &SeedProgress{}
	}

	//ret.seededTiles = {l: deque(maxlen=64) for l in task.levels}

	return ret
}

func (t *TileWalker) Walk() {}

func (t *TileWalker) walk(cur_bbox vec2d.Rect, levels []int, current_level int, all_subtiles bool) {}

func (t *TileWalker) reportProgress(level int, bbox vec2d.Rect) {}

func (t *TileWalker) filterSubtiles(subtiles int, all_subtiles []int) {}

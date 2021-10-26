package task

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/utils"
)

type TileWalker struct {
	manager                cache.Manager
	task                   Task
	workOnMetatiles        bool
	skipGeomsForLastLevels int
	reportTillLevel        int
	tilesPerMetatile       int
	taskProgress           *TaskProgress
	grid                   *geo.MetaGrid
	count                  int
	processedTiles         map[int]*utils.Deque
	progressLogger         ProgressLogger
	handleStale            bool
	handleUncached         bool
	pool                   WorkerPool
}

func NewTileWalker(task Task, tileWorkerPool WorkerPool, workOnMetatiles bool, skipGeomsForLastLevels int, progressLogger ProgressLogger, taskProgress *TaskProgress, handleStale, handleUNCached bool) *TileWalker {
	ret := &TileWalker{
		pool:                   tileWorkerPool,
		task:                   task,
		manager:                task.GetManager(),
		workOnMetatiles:        workOnMetatiles,
		skipGeomsForLastLevels: skipGeomsForLastLevels,
		taskProgress:           taskProgress,
		progressLogger:         progressLogger,
		handleStale:            handleStale,
		handleUncached:         handleUNCached,
	}

	num_seed_levels := len(task.GetLevels())
	if num_seed_levels >= 4 {
		ret.reportTillLevel = task.GetLevels()[num_seed_levels-2]
	} else {
		ret.reportTillLevel = task.GetLevels()[num_seed_levels-1]
	}
	var metaSize [2]uint32
	if ret.manager.GetMetaGrid() == nil {
		metaSize = [2]uint32{1, 1}
	} else {
		metaSize = ret.manager.GetMetaGrid().MetaSize
	}
	ret.tilesPerMetatile = int(metaSize[0] * metaSize[1])
	ret.grid = &geo.MetaGrid{TileGrid: *ret.manager.GetGrid(), MetaSize: metaSize, MetaBuffer: 0}
	ret.count = 0
	if taskProgress != nil {
		ret.taskProgress = taskProgress
	} else {
		ret.taskProgress = NewTaskProgress(nil)
	}

	ret.processedTiles = make(map[int]*utils.Deque)

	for _, l := range ret.task.GetLevels() {
		ret.processedTiles[l] = utils.NewDeque(64)
	}

	return ret
}

func (t *TileWalker) Walk() {
	bbox := t.task.GetCoverage().GetExtent().BBoxFor(t.manager.GetGrid().Srs)
	if t.taskProgress.AlreadyProcessed() {
		t.taskProgress.StepForward(1)
	} else {
		t.walk(bbox, t.task.GetLevels(), 0, false)
	}
	t.reportProgress(t.task.GetLevels()[0], t.task.GetCoverage().GetBBox())
}

func levelInLevels(level int, levels []int) bool {
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}

func limitSubBBox(bbox, sub_bbox vec2d.Rect) *vec2d.Rect {
	minx := math.Max(bbox.Min[0], sub_bbox.Min[0])
	miny := math.Max(bbox.Min[1], sub_bbox.Min[1])
	maxx := math.Min(bbox.Max[0], sub_bbox.Max[0])
	maxy := math.Min(bbox.Max[1], sub_bbox.Max[1])
	return &vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

func filterIsCached(manager cache.Manager, handle_tiles [][3]int) [][3]int {
	ret := make([][3]int, 0, len(handle_tiles))
	for i := range handle_tiles {
		if !manager.IsCached(handle_tiles[i], nil) {
			ret = append(ret, handle_tiles[i])
		}
	}
	return ret
}

func filterIsStale(manager cache.Manager, handle_tiles [][3]int) [][3]int {
	ret := make([][3]int, 0, len(handle_tiles))
	for i := range handle_tiles {
		if manager.IsStale(handle_tiles[i], nil) {
			ret = append(ret, handle_tiles[i])
		}
	}
	return ret
}

func (t *TileWalker) walk(cur_bbox vec2d.Rect, levels []int, currentLevel int, allSubtiles bool) bool {
	_, tiles, subtiles := t.grid.GetAffectedLevelTiles(cur_bbox, currentLevel)
	totalSubtiles := int(tiles[0] * tiles[1])
	if len(levels) < t.skipGeomsForLastLevels {
		allSubtiles = true
	}
	subtilesIt := t.filterSubtiles(subtiles, allSubtiles)

	if levelInLevels(currentLevel, levels) && currentLevel <= t.reportTillLevel {
		t.reportProgress(currentLevel, cur_bbox)
	}

	if !t.taskProgress.Running() {
		if levelInLevels(currentLevel, levels) {
			t.reportProgress(currentLevel, cur_bbox)
		}
		t.manager.Cleanup()
		return false
	}

	process := false
	if levelInLevels(currentLevel, levels) {
		levels = levels[1:]
		process = true
	}

	for subtilesIt.HasNext() {
		i, subtile, sub_bbox, intersection := subtilesIt.Next()

		if len(subtile) == 0 {
			t.taskProgress.StepForward(totalSubtiles)
			continue
		}

		if len(levels) > 0 {
			sub_bbox = limitSubBBox(cur_bbox, *sub_bbox)
			if intersection == CONTAINS {
				allSubtiles = true
			} else {
				allSubtiles = false
			}

			if !t.taskProgress.StepDown(i, totalSubtiles, func() bool {
				if t.taskProgress.AlreadyProcessed() {
					t.taskProgress.StepForward(1)
				} else {
					if !t.walk(*sub_bbox, levels, currentLevel+1, allSubtiles) {
						return false
					}
				}
				return true
			}) {
				return false
			}
		}

		if !process {
			continue
		}

		if t.processedTiles[currentLevel].Contains(subtile, func(a, b interface{}) bool {
			aa := a.([]int)
			bb := b.([]int)
			if aa[0] == bb[0] && aa[1] == bb[1] && aa[2] == bb[2] {
				return true
			}
			return false
		}) {
			if len(levels) == 0 {
				t.taskProgress.StepForward(totalSubtiles)
			}
			continue
		}
		t.processedTiles[currentLevel].PushFront(subtile)

		var handleTiles [][3]int

		if !t.workOnMetatiles {
			handleTiles = t.grid.TileList([3]int{subtile[0], subtile[1], subtile[2]})
		} else {
			handleTiles = append(handleTiles, [3]int{subtile[0], subtile[1], subtile[2]})
		}

		if t.handleUncached {
			handleTiles = filterIsCached(t.manager, handleTiles)
		} else if t.handleStale {
			handleTiles = filterIsStale(t.manager, handleTiles)
		}

		if handleTiles != nil {
			t.count += 1
			if !t.pool.Process(t.task.NewWork(handleTiles), t.taskProgress) {
				return false
			}
		}

		if len(levels) == 0 {
			t.taskProgress.StepForward(totalSubtiles)
		}
	}

	if len(levels) >= 4 {
		t.manager.Cleanup()
	}
	return true
}

func (t *TileWalker) reportProgress(level int, bbox vec2d.Rect) {
	if t.progressLogger != nil {
		t.progressLogger.LogProgress(t.taskProgress, level, bbox, t.count*t.tilesPerMetatile)
	}
}

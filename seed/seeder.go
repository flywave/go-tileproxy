package seed

import (
	"errors"
	"math"
	"reflect"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/utils"
)

type TileWalker struct {
	manager                cache.Manager
	task                   Task
	workOnMetatiles        bool
	skipGeomsForLastLevels int
	reportTillLevel        int
	tilesPerMetatile       int
	seedProgress           *SeedProgress
	grid                   *geo.MetaGrid
	count                  int
	seededTiles            map[int]*utils.Deque
	progressLogger         ProgressLogger
	handleStale            bool
	handleUncached         bool
	pool                   *TileWorkerPool
}

func NewTileWalker(task Task, tileWorkerPool *TileWorkerPool,
	workOnMetatiles bool, skipGeomsForLastLevels int, progressLogger ProgressLogger,
	seedProgress *SeedProgress, handleStale, handleUNCached bool) *TileWalker {
	ret := &TileWalker{pool: tileWorkerPool, task: task, manager: task.GetManager(), workOnMetatiles: workOnMetatiles,
		skipGeomsForLastLevels: skipGeomsForLastLevels, seedProgress: seedProgress, progressLogger: progressLogger, handleStale: handleStale, handleUncached: handleUNCached}

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
	if seedProgress != nil {
		ret.seedProgress = seedProgress
	} else {
		ret.seedProgress = NewSeedProgress()
	}

	ret.seededTiles = make(map[int]*utils.Deque)

	for _, l := range ret.task.GetLevels() {
		ret.seededTiles[l] = utils.NewDeque(64)
	}

	return ret
}

func (t *TileWalker) Walk() {
	bbox := t.task.GetCoverage().GetExtent().BBoxFor(t.manager.GetGrid().Srs)
	if t.seedProgress.AlreadyProcessed() {
		t.seedProgress.StepForward(1)
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

func (t *TileWalker) walk(cur_bbox vec2d.Rect, levels []int, current_level int, all_subtiles bool) bool {
	_, tiles, subtiles := t.grid.GetAffectedLevelTiles(cur_bbox, current_level)
	total_subtiles := int(tiles[0] * tiles[1])
	if len(levels) < t.skipGeomsForLastLevels {
		all_subtiles = true
	}
	subtiles_it := t.filterSubtiles(subtiles, all_subtiles)

	if levelInLevels(current_level, levels) && current_level <= t.reportTillLevel {
		t.reportProgress(current_level, cur_bbox)
	}

	if !t.seedProgress.Running() {
		if levelInLevels(current_level, levels) {
			t.reportProgress(current_level, cur_bbox)
		}
		t.manager.Cleanup()
		return false
	}

	process := false
	if levelInLevels(current_level, levels) {
		levels = levels[1:]
		process = true
	}
	i := 0
	for subtiles_it.HasNext() {
		subtile, sub_bbox, intersection := subtiles_it.Next()

		if subtile == nil {
			t.seedProgress.StepForward(total_subtiles)
			continue
		}

		if len(levels) > 0 {
			sub_bbox = limitSubBBox(cur_bbox, *sub_bbox)
			if intersection == CONTAINS {
				all_subtiles = true
			} else {
				all_subtiles = false
			}

			if !t.seedProgress.StepDown(i, total_subtiles, func() bool {
				if t.seedProgress.AlreadyProcessed() {
					t.seedProgress.StepForward(1)
				} else {
					if !t.walk(*sub_bbox, levels, current_level+1, all_subtiles) {
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

		if t.seededTiles[current_level].Contains(subtile, func(a, b interface{}) bool {
			aa := a.([]int)
			bb := b.([]int)
			if aa[0] == bb[0] && aa[1] == bb[1] && aa[2] == bb[2] {
				return true
			}
			return false
		}) {
			if len(levels) == 0 {
				t.seedProgress.StepForward(total_subtiles)
			}
			continue
		}
		t.seededTiles[current_level].PushFront(subtile)

		var handle_tiles [][3]int

		if !t.workOnMetatiles {
			handle_tiles = t.grid.TileList([3]int{subtile[0], subtile[1], subtile[2]})
		} else {
			handle_tiles = append(handle_tiles, [3]int{subtile[0], subtile[1], subtile[2]})
		}

		if t.handleUncached {
			handle_tiles = filterIsCached(t.manager, handle_tiles)
		} else if t.handleStale {
			handle_tiles = filterIsStale(t.manager, handle_tiles)
		}

		if handle_tiles != nil {
			t.count += 1
			t.pool.Process(&TileSeedWorker{task: t.task, manager: t.manager, tiles: handle_tiles}, t.seedProgress)
		}

		if levels == nil {
			t.seedProgress.StepForward(total_subtiles)
		}

		i++
	}

	if len(levels) >= 4 {
		t.manager.Cleanup()
	}
	return true
}

func (t *TileWalker) reportProgress(level int, bbox vec2d.Rect) {
	if t.progressLogger != nil {
		t.progressLogger.LogProgress(t.seedProgress, level, bbox, t.count*t.tilesPerMetatile)
	}
}

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

func (it *TileIterator) Next() (rsubtile []int, rsub_bbox *vec2d.Rect, intersection IntersectionType) {
	subtile := it.Subtiles[it.current]
	defer func() {
		it.current++
	}()

	if subtile[0] == -1 || subtile[1] == -1 || subtile[2] == -1 {
		return nil, nil, NONE
	} else {
		metatile := it.grid.GetMetaTile(subtile)
		sub_bbox := metatile.GetBBox()
		if it.AllSubtiles {
			intersection = CONTAINS
		} else {
			intersection = it.task.Intersects(sub_bbox)
		}
		if intersection != NONE {
			rsubtile = subtile[:]
			rsub_bbox = &sub_bbox
			return
		} else {
			return nil, nil, NONE
		}
	}
}

func (t *TileWalker) filterSubtiles(subtiles [][3]int, all_subtiles bool) *TileIterator {
	return &TileIterator{Subtiles: subtiles, AllSubtiles: all_subtiles, current: 0, grid: t.grid, task: t.task}
}

func seedTask(task *TileSeedTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, seedProgress *SeedProgress) error {
	if task.GetCoverage() == nil {
		return errors.New("task coverage is null!")
	}

	task.GetManager().SetMinimizeMetaRequests(false)

	var work_on_metatiles bool
	if task.GetManager().GetRescaleTiles() != 0 {
		work_on_metatiles = false
	}

	tile_worker_pool := NewTileWorkerPool(concurrency, task, progress_logger)
	tile_walker := NewTileWalker(task, tile_worker_pool, work_on_metatiles, skipGeomsForLastLevels, progress_logger, seedProgress, false, true)
	tile_walker.Walk()

	return nil
}

func reverse(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)

	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	swp := reflect.Swapper(s.Interface())
	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		swp(i, j)
	}
	return slice
}

func Seed(tasks []*TileSeedTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, progress_store ProgressStore, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	active_tasks := tasks[:]
	reverse(active_tasks)
	for len(active_tasks) > 0 {
		task := active_tasks[len(active_tasks)-1]
		md := task.GetMetadata()
		if err := cache_locker.Lock(md["cache_name"], func() error {
			var start_progress [][2]int
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskID(task.GetID())
				start_progress = progress_store.Get(task.GetID())
			} else {
				start_progress = nil
			}
			seed_progress := &SeedProgress{oldLevelProgresses: start_progress}
			return seedTask(task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileSeedTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

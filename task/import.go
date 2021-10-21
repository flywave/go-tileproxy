package task

import (
	"errors"

	"github.com/flywave/go-tileproxy/imports"
)

func importTask(task *TileImportTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, seedProgress *TaskProgress) error {
	if task.GetCoverage() == nil {
		return errors.New("task coverage is null!")
	}

	tile_worker_pool := NewTileWorkerPool(concurrency, task, progress_logger)
	tile_walker := NewTileWalker(task, tile_worker_pool, false, skipGeomsForLastLevels, progress_logger, seedProgress, false, true)
	tile_walker.Walk()

	return nil
}

func Import(io imports.ImportProvider, tasks []*TileImportTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, progress_store ProgressStore, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	active_tasks := tasks[:]
	active_tasks = reverse(active_tasks).([]*TileImportTask)
	for len(active_tasks) > 0 {
		task := active_tasks[len(active_tasks)-1]
		md := task.GetMetadata()
		if err := cache_locker.Lock(md["cache_name"], func() error {
			var start_progress []interface{}
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskId(task.GetID())
				start_progress = progress_store.Get(task.GetID()).([]interface{})
			} else {
				start_progress = nil
			}
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			return importTask(task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileImportTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

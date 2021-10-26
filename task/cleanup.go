package task

import (
	"context"
	"sync"
)

func Cleanup(ctx context.Context, tasks []*TileCleanupTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	progress_store := progress_logger.GetStore()
	active_tasks := tasks[:]
	active_tasks = reverse(active_tasks).([]*TileCleanupTask)
	for len(active_tasks) > 0 {
		task := active_tasks[len(active_tasks)-1]
		md := task.GetMetadata()
		if err := cache_locker.Lock(md["cache_name"], func() error {
			var start_progress [][2]int
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskId(task.GetID())
				start_progress = progress_store.Get(task.GetID()).([][2]int)
			} else {
				start_progress = nil
			}
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			return cleanupTask(ctx, task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileCleanupTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}

		task.GetManager().Cleanup()
	}
}

func cleanupTask(ctx context.Context, task *TileCleanupTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, seed_progress *TaskProgress) error {
	task.GetManager().SetExpireTimestamp(&task.RemoveTimestamp)
	task.GetManager().SetMinimizeMetaRequests(false)

	tile_worker_pool := NewTileWorkerPool(ctx, concurrency, task, progress_logger)

	var wg sync.WaitGroup

	wg.Add(1)
	var err error
	go func() {
		err = tile_worker_pool.Queue.Run()
		wg.Done()
	}()

	tile_walker := NewTileWalker(task, tile_worker_pool, false, skipGeomsForLastLevels, progress_logger, seed_progress, true, false)
	tile_walker.Walk()

	if tile_worker_pool.Queue.IsRuning() {
		tile_worker_pool.Queue.Stop()
	}

	wg.Wait()

	return err
}

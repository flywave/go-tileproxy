package task

import (
	"context"
	"sync"
)

func seedTask(cancel context.CancelFunc, task *TileSeedTask, concurrency int, progress_logger ProgressLogger, seedProgress *TaskProgress) {
	if task.GetCoverage() == nil {
		return
	}
	if task.RefreshTimestamp != nil {
		task.GetManager().SetExpireTimestamp(task.RefreshTimestamp)
	}

	task.GetManager().SetMinimizeMetaRequests(false)

	var work_on_metatiles bool
	if task.GetManager().GetRescaleTiles() != 0 {
		work_on_metatiles = false
	}

	tile_worker_pool := NewTileWorkerPool(cancel, concurrency, task, progress_logger)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		tile_worker_pool.Queue.Run()
		wg.Done()
	}()

	tile_walker := NewTileWalker(task, tile_worker_pool, work_on_metatiles, progress_logger, seedProgress, false, true)
	tile_walker.Walk()

	if tile_worker_pool.Queue.IsRuning() {
		tile_worker_pool.Queue.Stop()
	}

	wg.Wait()
}

func Seed(cancel context.CancelFunc, tasks []*TileSeedTask, concurrency int, progress_logger ProgressLogger, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	var progress_store ProgressStore
	if progress_logger != nil {
		progress_store = progress_logger.GetStore()
	}
	active_tasks := tasks[:]
	active_tasks = reverse(active_tasks).([]*TileSeedTask)
	for len(active_tasks) > 0 {
		task := active_tasks[len(active_tasks)-1]
		if task == nil {
			continue
		}
		md := task.GetMetadata()
		var cacheName string
		if md != nil {
			if cn, ok := md["cache_name"]; ok {
				cacheName = cn.(string)
			}
		} else {
			cacheName = "default"
		}
		if err := cache_locker.Lock(cacheName, func() {
			var start_progress [][2]int
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskId(task.GetID())
				data := progress_store.Get(task.GetID())
				if data != nil {
					start_progress = data.([][2]int)
				}
			} else {
				start_progress = nil
			}
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			seedTask(cancel, task, concurrency, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileSeedTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

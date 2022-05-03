package task

import (
	"context"
	"sync"

	"github.com/flywave/go-tileproxy/imports"
)

func importTask(cancel context.CancelFunc, task *TileImportTask, concurrency int, progress_logger ProgressLogger, seedProgress *TaskProgress) {
	if task.GetCoverage() == nil {
		return
	}

	tile_worker_pool := NewTileWorkerPool(cancel, concurrency, task, progress_logger)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		tile_worker_pool.Queue.Run()
		wg.Done()
	}()

	tile_walker := NewTileWalker(task, tile_worker_pool, false, progress_logger, seedProgress, false, true)
	tile_walker.Walk()

	if tile_worker_pool.Queue.IsRuning() {
		tile_worker_pool.Queue.Stop()
	}

	wg.Wait()
}

func Import(cancel context.CancelFunc, io imports.Import, tasks []*TileImportTask, concurrency int, progress_logger ProgressLogger, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	progress_store := progress_logger.GetStore()
	active_tasks := tasks[:]
	active_tasks = reverse(active_tasks).([]*TileImportTask)
	for len(active_tasks) > 0 {
		task := active_tasks[len(active_tasks)-1]
		md := task.GetMetadata()
		var cacheName string
		if cn, ok := md["cache_name"]; ok {
			cacheName = cn.(string)
		} else {
			cacheName = "default"
		}
		if err := cache_locker.Lock(cacheName, func() {
			var start_progress [][2]int
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskId(task.GetID())
				start_progress = progress_store.Get(task.GetID()).([][2]int)
			} else {
				start_progress = nil
			}
			task.io = io
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			importTask(cancel, task, concurrency, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileImportTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

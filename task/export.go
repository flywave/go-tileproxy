package task

import (
	"context"
	"sync"

	"github.com/flywave/go-tileproxy/exports"
)

func exportTask(cancel context.CancelFunc, task *TileExportTask, concurrency int, progress_logger ProgressLogger, seedProgress *TaskProgress) {
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

	tile_walker := NewTileWalker(task, tile_worker_pool, work_on_metatiles, progress_logger, seedProgress, false, false)
	tile_walker.Walk()

	if tile_worker_pool.Queue.IsRuning() {
		tile_worker_pool.Queue.Stop()
	}

	wg.Wait()
}

func Export(cancel context.CancelFunc, io exports.Export, tasks []*TileExportTask, concurrency int, progress_logger ProgressLogger, cache_locker CacheLocker) {
	if cache_locker == nil {
		cache_locker = &DummyCacheLocker{}
	}

	var progress_store ProgressStore
	if progress_logger != nil {
		progress_store = progress_logger.GetStore()
	}
	active_tasks := tasks[:]
	active_tasks = reverse(active_tasks).([]*TileExportTask)
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
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			task.io = io
			exportTask(cancel, task, concurrency, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileExportTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

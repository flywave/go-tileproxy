package task

import (
	"context"
	"errors"
	"sync"

	"github.com/flywave/go-tileproxy/imports"
)

func importTask(ctx context.Context, task *TileImportTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, seedProgress *TaskProgress) error {
	if task.GetCoverage() == nil {
		return errors.New("task coverage is null")
	}

	tile_worker_pool := NewTileWorkerPool(ctx, concurrency, task, progress_logger)

	var wg sync.WaitGroup

	wg.Add(1)

	var err error

	go func() {
		err = tile_worker_pool.Queue.Run()
		wg.Done()
	}()

	tile_walker := NewTileWalker(task, tile_worker_pool, false, skipGeomsForLastLevels, progress_logger, seedProgress, false, true)
	tile_walker.Walk()

	if tile_worker_pool.Queue.IsRuning() {
		tile_worker_pool.Queue.Stop()
	}

	wg.Wait()

	return err
}

func Import(ctx context.Context, io imports.Import, tasks []*TileImportTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, cache_locker CacheLocker) {
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
		if err := cache_locker.Lock(cacheName, func() error {
			var start_progress [][2]int
			if progress_logger != nil && progress_store != nil {
				progress_logger.SetCurrentTaskId(task.GetID())
				start_progress = progress_store.Get(task.GetID()).([][2]int)
			} else {
				start_progress = nil
			}
			seed_progress := &TaskProgress{oldLevelProgresses: start_progress}
			return importTask(ctx, task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)
		}); err != nil {
			active_tasks = append([]*TileImportTask{task}, active_tasks[:len(active_tasks)-1]...)
		} else {
			active_tasks = active_tasks[:len(active_tasks)-1]
		}
	}
}

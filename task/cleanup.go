package task

func Cleanup(tasks []*TileCleanupTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger) {
	progress_store := progress_logger.GetStore()
	for _, task := range tasks {
		if task.GetCoverage() == nil {
			continue
		}

		var seed_progress *TaskProgress
		if progress_logger != nil && progress_store != nil {
			progress_logger.SetCurrentTaskId(task.GetID())
			start_progress := progress_store.Get(task.GetID())
			seed_progress = &TaskProgress{oldLevelProgresses: start_progress.([]interface{})}
		}

		tileWalkerCleanup(task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)

		task.GetManager().Cleanup()
	}
}

func tileWalkerCleanup(task *TileCleanupTask, concurrency int, skipGeomsForLastLevels int,
	progress_logger ProgressLogger, seed_progress *TaskProgress) {
	task.GetManager().SetExpireTimestamp(&task.RemoveTimestamp)
	task.GetManager().SetMinimizeMetaRequests(false)

	tile_worker_pool := NewTileWorkerPool(concurrency, task, progress_logger)
	tile_walker := NewTileWalker(task, tile_worker_pool, false, skipGeomsForLastLevels, progress_logger, seed_progress, true, false)
	tile_walker.Walk()
}

package seed

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/utils"
)

func Cleanup(tasks []*TileCleanupTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, progress_store ProgressStore) {
	for _, task := range tasks {
		if task.GetCoverage() == nil {
			continue
		}

		var seed_progress *SeedProgress
		var cleanup_progress *DirectoryCleanupProgress
		if progress_logger != nil && progress_store != nil {
			progress_logger.SetCurrentTaskID(task.GetID())
			start_progress := progress_store.Get(task.GetID())
			seed_progress = &SeedProgress{oldLevelProgresses: start_progress}
			cleanup_progress = &DirectoryCleanupProgress{oldDir: start_progress}
		}

		if task.CompleteExtent {
			if false {
				simpleCleanup(task, progress_logger, cleanup_progress)
				task.GetManager().Cleanup()
				continue
			} else if true {
				cacheCleanup(task, progress_logger)
				task.GetManager().Cleanup()
				continue
			}
		}

		tileWalkerCleanup(task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)

		task.GetManager().Cleanup()
	}
}

func simpleCleanup(task *TileCleanupTask, progress_logger ProgressLogger, cleanup_progress *DirectoryCleanupProgress) {
	for _, level := range task.GetLevels() {
		level_dir := task.GetManager().GetCache().LevelLocation(level)
		if progress_logger != nil {
			progress_logger.LogMessage(fmt.Sprintf("removing old tiles in %s", normpath(level_dir)))
			if progress_logger.GetStore() != nil {
				cleanup_progress.stepDir(level_dir)
				if cleanup_progress.AlreadyProcessed() {
					continue
				}
				progress_logger.GetStore().Store(
					task.GetID(),
					cleanup_progress.CurrentProgressIdentifier(),
				)
				progress_logger.GetStore().Save()
			}
		}
		cleanupDirectory(level_dir, task.RemoveTimestamp, true)
	}
}

func cleanupDirectory(directory string, before_timestamp time.Time, remove_empty_dirs bool) {
	if before_timestamp.IsZero() && remove_empty_dirs && utils.FileExists(directory) {
		os.RemoveAll(directory)
		return
	}

	if utils.FileExists(directory) {
		dirs, _ := ioutil.ReadDir(directory)
		for _, fi := range dirs {
			if fi.IsDir() {
				cleanupDirectory(path.Join(directory, fi.Name()), before_timestamp, remove_empty_dirs)
				continue
			} else {
				filename := path.Join(directory, fi.Name())
				if before_timestamp.IsZero() {
					os.Remove(filename)
				}
				st, _ := os.Stat(filename)
				if st.ModTime().Before(before_timestamp) {
					os.Remove(filename)
				}
			}
		}
		if remove_empty_dirs {
			os.Remove(directory)
		}
	}
}

func cacheCleanup(task *TileCleanupTask, progress_logger ProgressLogger) {
	for _, level := range task.GetLevels() {
		if progress_logger != nil {
			progress_logger.LogMessage(fmt.Sprintf("removing old tiles for level %d", level))
		}
		task.GetManager().GetCache().RemoveLevelTilesBefore(level, task.RemoveTimestamp)
		task.GetManager().Cleanup()
	}
}

func normpath(path string) string {
	if strings.HasPrefix(path, "\\") {
		return path
	}

	currentwd, _ := os.Getwd()

	path, _ = filepath.Rel(path, currentwd)

	if strings.HasPrefix(path, "../../") {
		path, _ = filepath.Abs(path)
	}
	return path
}

type DirectoryCleanupProgress struct {
	oldDir     []int
	currentDir []int
}

func (p *DirectoryCleanupProgress) stepDir(dir string) {
	tiles := strings.Split(dir, "/")
	p.currentDir = make([]int, len(tiles))
	for i := range tiles {
		p.currentDir[i], _ = strconv.Atoi(tiles[i])
	}
}

func (p *DirectoryCleanupProgress) AlreadyProcessed() bool {
	return p.canSkip(p.oldDir, p.currentDir)
}

func (p *DirectoryCleanupProgress) CurrentProgressIdentifier() []int {
	if p.AlreadyProcessed() || p.currentDir == nil {
		return p.oldDir
	}
	return p.currentDir
}

func (p *DirectoryCleanupProgress) Running() bool {
	return true
}

func (p *DirectoryCleanupProgress) canSkip(old_progress, current_progress []int) bool {
	if old_progress == nil {
		return false
	}
	if current_progress == nil {
		return false
	}

	zips := iziplongest(-1, old_progress, current_progress)

	for i := range zips {
		old := zips[i][0]
		current := zips[i][1]
		if old == -1 {
			return false
		}
		if current == -1 {
			return false

		}
		if old < current {
			return false
		}
		if old > current {
			return true
		}
	}
	return false
}

func tileWalkerCleanup(task *TileCleanupTask, concurrency int, skipGeomsForLastLevels int,
	progress_logger ProgressLogger, seed_progress *SeedProgress) {
	task.GetManager().SetExpireTimestamp(&task.RemoveTimestamp)
	task.GetManager().SetMinimizeMetaRequests(false)

	tile_worker_pool := NewTileWorkerPool(concurrency, task, progress_logger)
	tile_walker := NewTileWalker(task, tile_worker_pool, false, skipGeomsForLastLevels, progress_logger, seed_progress, true, false)
	tile_walker.Walk()
}

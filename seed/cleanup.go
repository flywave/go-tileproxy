package seed

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/utils"
)

type CacheIO interface {
	Remove(filename string) error
	RemoveDirIfEmtpy(directory string) error
	EnsureDirectory(directory string) error
	Stat(file string) (os.FileInfo, error)
	FileExists(file string) bool
	ReadDir(dirname string) ([]fs.DirEntry, error)
}

type LocalCacheIO struct {
	CacheIO
}

func (i *LocalCacheIO) Remove(filename string) error {
	return os.Remove(filename)
}

func (i *LocalCacheIO) RemoveDirIfEmtpy(directory string) error {
	return os.Remove(directory)
}

func (i *LocalCacheIO) EnsureDirectory(directory string) error {
	return os.RemoveAll(directory)
}

func (i *LocalCacheIO) Stat(file string) (os.FileInfo, error) {
	return os.Stat(file)
}

func (i *LocalCacheIO) FileExists(file string) bool {
	return utils.FileExists(file)
}

func (i *LocalCacheIO) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirname)
}

func Cleanup(io CacheIO, tasks []*TileCleanupTask, concurrency int, skipGeomsForLastLevels int, progress_logger ProgressLogger, progress_store ProgressStore) {
	for _, task := range tasks {
		if task.GetCoverage() == nil {
			continue
		}

		var seed_progress *SeedProgress
		var cleanup_progress *DirectoryCleanupProgress
		if progress_logger != nil && progress_store != nil {
			progress_logger.SetCurrentTaskId(task.GetID())
			start_progress := progress_store.Get(task.GetID())
			seed_progress = &SeedProgress{oldLevelProgresses: start_progress.([]interface{})}
			cleanup_progress = newDirectoryCleanupProgress(start_progress.([]interface{}))
		}

		if task.CompleteExtent {
			simpleCleanup(io, task, progress_logger, cleanup_progress)
			task.GetManager().Cleanup()
			continue
		}

		tileWalkerCleanup(task, concurrency, skipGeomsForLastLevels, progress_logger, seed_progress)

		task.GetManager().Cleanup()
	}
}

func simpleCleanup(io CacheIO, task *TileCleanupTask, progress_logger ProgressLogger, cleanup_progress *DirectoryCleanupProgress) {
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
		cleanupDirectory(io, level_dir, task.RemoveTimestamp, true)
	}
}

func cleanupDirectory(io CacheIO, directory string, before_timestamp time.Time, remove_empty_dirs bool) error {
	if before_timestamp.IsZero() && remove_empty_dirs && io.FileExists(directory) {
		return io.EnsureDirectory(directory)
	}

	if io.FileExists(directory) {
		dirs, _ := io.ReadDir(directory)
		for _, fi := range dirs {
			if fi.IsDir() {
				cleanupDirectory(io, path.Join(directory, fi.Name()), before_timestamp, remove_empty_dirs)
				continue
			} else {
				filename := path.Join(directory, fi.Name())
				if before_timestamp.IsZero() {
					err := io.Remove(filename)
					if err != nil {
						return err
					}
				}
				st, _ := os.Stat(filename)
				if st.ModTime().Before(before_timestamp) {
					err := io.Remove(filename)
					if err != nil {
						return err
					}
				}
			}
		}
		if remove_empty_dirs {
			io.RemoveDirIfEmtpy(directory)
		}
	}
	return nil
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
	oldDir     string
	currentDir string
}

func newDirectoryCleanupProgress(start_progress []interface{}) *DirectoryCleanupProgress {
	dir := ""
	for _, p := range start_progress {
		cold := p.([2]int)
		dir = path.Join(dir, strconv.Itoa(cold[0]))
	}
	ret := &DirectoryCleanupProgress{oldDir: dir}
	return ret
}

func (p *DirectoryCleanupProgress) stepDir(dir string) {
	p.currentDir = dir
}

func (p *DirectoryCleanupProgress) AlreadyProcessed() bool {
	return p.canSkip(p.oldDir, p.currentDir)
}

func (p *DirectoryCleanupProgress) CurrentProgressIdentifier() interface{} {
	if p.AlreadyProcessed() || p.currentDir == "" {
		return p.oldDir
	}
	return p.currentDir
}

func (p *DirectoryCleanupProgress) Running() bool {
	return true
}

func (p *DirectoryCleanupProgress) canSkip(old_dir, current_dir string) bool {
	if old_dir == "" {
		return false
	}
	if current_dir == "" {
		return false
	}

	var _old_progress []interface{}
	var _current_progress []interface{}

	if old_dir != "" {
		dirs := strings.Split(old_dir, "/")
		_old_progress = make([]interface{}, len(dirs))
		for i := range dirs {
			_old_progress[i], _ = strconv.Atoi(dirs[i])
		}
	}
	if current_dir != "" {
		dirs := strings.Split(current_dir, "/")
		_current_progress = make([]interface{}, len(dirs))
		for i := range dirs {
			_current_progress[i], _ = strconv.Atoi(dirs[i])
		}
	}

	zips := izip_longest(-1, _old_progress, _current_progress)

	for i := range zips {
		old := zips[i][0]
		current := zips[i][1]
		if old == nil {
			return false
		}
		if current == nil {
			return false

		}
		cold := old.(int)
		ccurrent := current.(int)
		if cold < ccurrent {
			return false
		}
		if cold > ccurrent {
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

package task

import (
	"fmt"
	"strconv"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/exports"
	"github.com/flywave/go-tileproxy/imports"
)

type Task interface {
	GetID() string
	GetMetadata() map[string]interface{}
	GetManager() cache.Manager
	GetLevels() []int
	GetCoverage() geo.Coverage
	Intersects(vec2d.Rect) IntersectionType
	NewWork(handle_tiles [][3]int) Work
}

type BaseTask struct {
	Metadata map[string]interface{}
	Manager  cache.Manager
	Coverage geo.Coverage
	Grid     *geo.TileGrid
	Levels   []int
}

func (t *BaseTask) GetMetadata() map[string]interface{} {
	return t.Metadata
}

func (t *BaseTask) GetManager() cache.Manager {
	return t.Manager
}

func (t *BaseTask) GetLevels() []int {
	return t.Levels
}

func (t *BaseTask) GetCoverage() geo.Coverage {
	return t.Coverage
}

func (t *BaseTask) Intersects(bbox vec2d.Rect) IntersectionType {
	if t.Coverage == nil || t.Coverage.Contains(bbox, t.Grid.Srs) {
		return CONTAINS
	}
	if t.Coverage.Intersects(bbox, t.Grid.Srs) {
		return INTERSECTS
	}
	return NONE
}

type TileSeedTask struct {
	BaseTask
	RefreshTimestamp *time.Time
}

func (t *TileSeedTask) NewWork(handle_tiles [][3]int) Work {
	return &SeedWorker{task: t, manager: t.Manager, tiles: handle_tiles, done: make(chan struct{})}
}

func NewTileSeedTask(md map[string]interface{}, manager cache.Manager, levels []int, refresh_timestamp *time.Time, coverage geo.Coverage) *TileSeedTask {
	return &TileSeedTask{
		BaseTask: BaseTask{
			Metadata: md,
			Manager:  manager,
			Coverage: coverage,
			Grid:     manager.GetGrid(),
			Levels:   levels,
		},
		RefreshTimestamp: refresh_timestamp,
	}
}

func (t *TileSeedTask) GetID() string {
	l := "level"
	for _, level := range t.Levels {
		l += "-" + strconv.Itoa(level)
	}
	return fmt.Sprintf("%s %s %s %s", t.Metadata["name"], t.Metadata["cache_name"], t.Metadata["grid_name"], l)
}

type TileCleanupTask struct {
	BaseTask
	RemoveTimestamp time.Time
}

func (t *TileCleanupTask) NewWork(handle_tiles [][3]int) Work {
	return &CleanupWorker{task: t, manager: t.Manager, tiles: handle_tiles, done: make(chan struct{})}
}

func NewTileCleanupTask(md map[string]interface{}, manager cache.Manager, levels []int, remove_timestamp time.Time, coverage geo.Coverage) *TileCleanupTask {
	return &TileCleanupTask{
		BaseTask: BaseTask{
			Metadata: md,
			Manager:  manager,
			Coverage: coverage,
			Grid:     manager.GetGrid(),
			Levels:   levels,
		},
		RemoveTimestamp: remove_timestamp,
	}
}

func (t *TileCleanupTask) GetID() string {
	return fmt.Sprintf("cleanup %s %s %s", t.Metadata["name"], t.Metadata["cache_name"], t.Metadata["grid_name"])
}

type TileExportTask struct {
	BaseTask
	RefreshTimestamp *time.Time
	io               exports.Export
}

func (t *TileExportTask) NewWork(handle_tiles [][3]int) Work {
	return &ExportWorker{task: t, manager: t.Manager, io: t.io, tiles: handle_tiles, done: make(chan struct{})}
}

func NewTileExportTask(md map[string]interface{}, manager cache.Manager, levels []int, coverage geo.Coverage) *TileExportTask {
	return &TileExportTask{
		BaseTask: BaseTask{
			Metadata: md,
			Manager:  manager,
			Coverage: coverage,
			Grid:     manager.GetGrid(),
			Levels:   levels,
		},
	}
}

func (t *TileExportTask) GetID() string {
	return fmt.Sprintf("export %s %s %s", t.Metadata["name"], t.Metadata["cache_name"], t.Metadata["grid_name"])
}

type TileImportTask struct {
	BaseTask
	ForceOverwrite bool
	io             imports.Import
}

func (t *TileImportTask) NewWork(handle_tiles [][3]int) Work {
	return &ImportWorker{task: t, manager: t.Manager, io: t.io, tiles: handle_tiles, done: make(chan struct{}), force_overwrite: t.ForceOverwrite}
}

func NewTileImportTask(md map[string]interface{}, manager cache.Manager, levels []int, coverage geo.Coverage) *TileImportTask {
	return &TileImportTask{
		BaseTask: BaseTask{
			Metadata: md,
			Manager:  manager,
			Coverage: coverage,
			Grid:     manager.GetGrid(),
			Levels:   levels,
		},
	}
}

func (t *TileImportTask) GetID() string {
	return fmt.Sprintf("import %s %s %s", t.Metadata["name"], t.Metadata["cache_name"], t.Metadata["grid_name"])
}

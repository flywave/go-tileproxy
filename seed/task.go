package seed

import (
	"fmt"
	"strconv"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
)

type Task interface {
	GetID() string
	GetManager() cache.Manager
	GetLevels() []int
	GetCoverage() geo.Coverage
	Intersects(vec2d.Rect) IntersectionType
}

type BaseTask struct {
	Metadata map[string]string
	Manager  cache.Manager
	Coverage geo.Coverage
	Grid     *geo.TileGrid
	Levels   []int
}

func (t *BaseTask) GetManager() cache.Manager {
	return t.Manager
}

func (t *BaseTask) GetLevels() []int {
	return t.Levels
}

func (t *BaseTask) GetGrid() *geo.TileGrid {
	return t.Grid
}

func (t *BaseTask) GetCoverage() geo.Coverage {
	return t.Coverage
}

func (t *BaseTask) Intersects(bbox vec2d.Rect) IntersectionType {
	if t.Coverage.Contains(bbox, t.Grid.Srs) {
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

func NewTileSeedTask(md map[string]string, manager cache.Manager, levels []int, refresh_timestamp *time.Time, coverage geo.Coverage) *TileSeedTask {
	return &TileSeedTask{BaseTask: BaseTask{Metadata: md, Manager: manager, Coverage: coverage, Grid: manager.GetGrid(), Levels: levels}, RefreshTimestamp: refresh_timestamp}
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
	CompleteExtent  bool
}

func NewTileCleanupTask(md map[string]string, manager cache.Manager, levels []int, remove_timestamp time.Time, coverage geo.Coverage, complete_extent bool) *TileCleanupTask {
	return &TileCleanupTask{BaseTask: BaseTask{Metadata: md, Manager: manager, Coverage: coverage, Grid: manager.GetGrid(), Levels: levels}, RemoveTimestamp: remove_timestamp, CompleteExtent: complete_extent}
}

func (t *TileCleanupTask) GetID() string {
	return fmt.Sprintf("cleanup %s %s %s", t.Metadata["name"], t.Metadata["cache_name"], t.Metadata["grid_name"])
}

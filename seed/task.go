package seed

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type Task interface {
	Run() error
	GetID() string
	GetManager() cache.Manager
	GetLevels() []int
	GetCoverage() geo.Coverage
	Intersects(vec2d.Rect) IntersectionType
}

type BaseTask struct {
	Conf     map[string]string
	Manager  cache.Manager
	Coverage geo.Coverage
	Grid     *geo.MetaGrid
	Levels   []int
}

func (t *BaseTask) GetLevels() []int {
	return t.Levels
}

func (t *BaseTask) GetGrid() *geo.MetaGrid {
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
	RefreshTimestamp time.Time
}

func (t *TileSeedTask) GetID() string {
	l := "level"
	for _, level := range t.Levels {
		l += "-" + strconv.Itoa(level)
	}
	return fmt.Sprintf("%s %s %s %s", t.Conf["name"], t.Conf["cache_name"], t.Conf["grid_name"], l)
}

type TileCleanupTask struct {
	BaseTask
	RemoveTimestamp time.Time
	CompleteExtent  bool
}

func (t *TileCleanupTask) GetID() string {
	return fmt.Sprintf("cleanup %s %s %s", t.Conf["name"], t.Conf["cache_name"], t.Conf["grid_name"])
}

package seed

import (
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
	Task
}

type TileCleanupTask struct {
	Task
}

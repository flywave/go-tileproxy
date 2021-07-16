package seed

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
)

type Task interface {
	Run()
	GetManager() layer.TileManager
	GetLevels() []int
	GetCoverage() geo.Coverage
}

type TileSeedTask struct {
	Task
}

type TileCleanupTask struct {
	Task
}

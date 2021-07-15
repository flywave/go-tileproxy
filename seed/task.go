package seed

import "github.com/flywave/go-tileproxy/layer"

type Task interface {
	Run()
	GetManager() layer.TileManager
	GetLevels() []int
}

type TileSeedTask struct {
	Task
}

type TileCleanupTask struct {
	Task
}

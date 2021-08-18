package seed

import "github.com/flywave/go-tileproxy/utils"

type TileWorkerPool struct {
	TileQueue *utils.Deque
	Logger    ProgressLogger
	Task      Task
}

func NewTileWorkerPool(task Task, logger ProgressLogger) *TileWorkerPool {
	return &TileWorkerPool{TileQueue: utils.NewDeque(10), Logger: logger, Task: task}
}

func (p *TileWorkerPool) Process(tiles [][3]int, progress *SeedProgress) {
	p.TileQueue.PushBack(tiles)

	if p.Logger != nil {
		p.Logger.LogStep(progress)
	}
}

package backend

import (
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type FetcherStatus int32

const (
	FETCHER_INIT   FetcherStatus = 0
	FETCHER_RUNING FetcherStatus = 1
	FETCHER_ERROR  FetcherStatus = 2
	FETCHER_DONE   FetcherStatus = 3
)

type TileKey struct {
	Z uint
	X uint
	Y uint
}

type TileArray []TileKey

type Fetcher interface {
	Bounds() vec2d.Rect
	LevelRange() [2]int
	Layer() string
	User() string
	ForceUpdate() bool
	Status() FetcherStatus
	UpdateStatus(tp FetcherStatus) bool
	CreateTime() time.Time
	CompleteTime() time.Time
	Progress() (finished, total int)
	Finished() TileArray
	Errored() TileArray
}

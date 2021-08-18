package seed

import vec2d "github.com/flywave/go3d/float64/vec2"

type ProgressStore interface {
	Store(id string, progress []int)
	Get(id string) []int
	Load() map[string][]int
	Save() error
	Remove() error
}

type ProgressLogger interface {
	LogMessage(msg string)
	LogStep(progress *SeedProgress)
	LogProgress(seed *SeedProgress, level int, bbox vec2d.Rect, tiles int)
	SetCurrentTaskID(id string)
	GetStore() ProgressStore
}

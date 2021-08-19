package seed

import vec2d "github.com/flywave/go3d/float64/vec2"

type ProgressStore interface {
	Store(id string, progress interface{})
	Get(id string) interface{}
	Load() map[string]interface{}
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

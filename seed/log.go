package seed

import vec2d "github.com/flywave/go3d/float64/vec2"

type ProgressStore interface {
	Store(id string, progress [][2]int)
	Get(id string) [][2]int
	Load() map[string][][2]int
	Save() error
	Remove() error
}

type ProgressLogger interface {
	LogMessage(msg string)
	LogStep(progress *SeedProgress)
	LogProgress(seed *SeedProgress, level int, bbox vec2d.Rect, tiles int)
	SetCurrentTaskID(id string)
}

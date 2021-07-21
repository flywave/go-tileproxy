package seed

import vec2d "github.com/flywave/go3d/float64/vec2"

type ProgressStore interface {
	StoreProgress(id string, progress [][2]int)
	LoadProgress(id string) [][2]int
}

type ProgressLogger interface {
	LogMessage()
	LogStep(progress *SeedProgress)
	LogProgress(seed *SeedProgress, level int, bbox vec2d.Rect, tiles int)
	SetCurrentTaskID(id string)
}

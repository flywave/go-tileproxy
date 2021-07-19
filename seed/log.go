package seed

import vec2d "github.com/flywave/go3d/float64/vec2"

type ProgressLogger struct {
	currentTaskID string
	progressStore map[string][][2]int
}

func (p *ProgressLogger) LogMessage() {

}

func (p *ProgressLogger) LogStep() {

}

func (p *ProgressLogger) LogProgress(seed *SeedProgress, level int, bbox vec2d.Rect, tiles int) {

}

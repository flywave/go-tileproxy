package task

import (
	"math"
	"strings"
)

type TaskProgress struct {
	progress                 float32
	levelProgressPercentages []float32
	levelProgresses          []interface{}
	levelProgressesLevel     int
	progressStrParts         []string
	oldLevelProgresses       []interface{}
}

func NewTaskProgress(oldLevelProgresses []interface{}) *TaskProgress {
	return &TaskProgress{progress: 0.0, levelProgressPercentages: []float32{1.0}, levelProgressesLevel: 0, progressStrParts: []string{}, oldLevelProgresses: oldLevelProgresses}
}

func (p *TaskProgress) StepForward(subtiles int) {
	p.progress += float32(p.levelProgressPercentages[len(p.levelProgressPercentages)-1]) / float32(subtiles)
}

func (p *TaskProgress) ToString() string {
	return strings.Join(p.progressStrParts, "")
}

func statusSymbol(i, total int) string {
	symbols := string(" .oO0")
	i += 1
	if 0 < i && i > total {
		return "X"
	} else {
		x := uint32(math.Ceil(float64(i) / float64(total/4)))
		return string(symbols[x])
	}
}

func (p *TaskProgress) StepDown(i, subtiles int, task func() bool) bool {
	if p.levelProgresses == nil {
		p.levelProgresses = []interface{}{}
	}
	p.levelProgresses = p.levelProgresses[:p.levelProgressesLevel]
	p.levelProgresses = append(p.levelProgresses, [2]int{i, subtiles})
	p.levelProgressesLevel += 1
	p.progressStrParts = append(p.progressStrParts, statusSymbol(i, subtiles))
	p.levelProgressPercentages = append(p.levelProgressPercentages, p.levelProgressPercentages[len(p.levelProgressPercentages)-1]/float32(subtiles))

	if !task() {
		return false
	}

	p.levelProgressPercentages = p.levelProgressPercentages[1:]
	p.progressStrParts = p.progressStrParts[1:]

	p.levelProgressesLevel -= 1
	if p.levelProgressesLevel == 0 {
		p.levelProgresses = []interface{}{}
	}
	return true
}

func (p *TaskProgress) Running() bool {
	return true
}

func (p *TaskProgress) Progress() float32 {
	return p.progress
}

func (p *TaskProgress) AlreadyProcessed() bool {
	return p.canSkip(p.oldLevelProgresses, p.levelProgresses)
}

func (p *TaskProgress) CurrentProgressIdentifier() interface{} {
	if p.AlreadyProcessed() || p.levelProgresses == nil {
		return p.oldLevelProgresses
	}
	return p.levelProgresses[:]
}

func (p *TaskProgress) canSkip(old_progress, current_progress []interface{}) bool {
	if current_progress == nil {
		return false
	}
	if old_progress == nil {
		return false
	}
	if len(old_progress) == 0 {
		return true
	}

	zips := izip_longest(nil, old_progress, current_progress)
	for i := range zips {
		old := zips[i][0]
		current := zips[i][1]
		if old == nil {
			return false
		}
		if current == nil {
			return false
		}
		cold := old.([2]int)
		ccurrent := current.([2]int)
		if cold[0] < ccurrent[0] {
			return false
		} else if cold[0] == ccurrent[0] {
			if cold[1] < ccurrent[1] {
				return false
			}
		}
		if cold[0] > ccurrent[0] {
			return true
		} else if cold[0] == ccurrent[0] {
			if cold[1] > ccurrent[1] {
				return true
			}
		}
	}
	return false
}

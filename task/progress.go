package task

type TaskProgress struct {
	currertTiles         int
	totalTiles           int
	levelProgresses      [][2]int
	levelProgressesLevel int
	oldLevelProgresses   [][2]int
}

func NewTaskProgress(oldLevelProgresses [][2]int) *TaskProgress {
	return &TaskProgress{
		currertTiles:         0,
		totalTiles:           0,
		levelProgressesLevel: 0,
		oldLevelProgresses:   oldLevelProgresses,
	}
}

func (p *TaskProgress) StepDown(i, subtiles int, task func() bool) bool {
	if p.levelProgresses == nil {
		p.levelProgresses = [][2]int{}
	}
	p.levelProgresses = p.levelProgresses[:p.levelProgressesLevel]
	p.levelProgresses = append(p.levelProgresses, [2]int{i, subtiles})
	p.levelProgressesLevel += 1

	if !task() {
		return false
	}

	p.levelProgressesLevel -= 1
	if p.levelProgressesLevel == 0 {
		p.levelProgresses = [][2]int{}
	}
	return true
}

func (p *TaskProgress) Update(tiles int) {
	p.currertTiles += tiles
}

func (p *TaskProgress) TotalTiles() int {
	return p.totalTiles
}

func (p *TaskProgress) CurrertTiles() int {
	return p.currertTiles
}

func (p *TaskProgress) Progress() float32 {
	return float32(p.currertTiles) / float32(p.totalTiles)
}

func (p *TaskProgress) AlreadyProcessed() bool {
	return p.canSkip(p.oldLevelProgresses, p.levelProgresses)
}

func (p *TaskProgress) CurrentProgressIdentifier() [][2]int {
	if p.AlreadyProcessed() || p.levelProgresses == nil {
		return p.oldLevelProgresses
	}
	return p.levelProgresses[:]
}

func (p *TaskProgress) canSkip(old_progress, current_progress [][2]int) bool {
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

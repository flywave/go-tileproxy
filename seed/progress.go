package seed

type SeedProgress struct {
	progress                 float32
	levelProgressPercentages []float32
	levelProgresses          [][2]int
	levelProgressesLevel     int
	progressStrParts         []string
	oldLevelProgresses       [][2]int
}

func (p *SeedProgress) StepForward(subtiles int) {

}

func (p *SeedProgress) ToString() string {
	return ""
}

func (p *SeedProgress) StepDown(i, subtiles int) {

}

func (p *SeedProgress) AlreadyProcessed() bool {
	return p.canSkip(p.oldLevelProgresses, p.levelProgresses)
}

func iziplongest(fillvalue int, iterables ...[][2]int) [][]int {
	if len(iterables) == 0 {
		return nil
	}

	size := len(iterables[0])
	for _, v := range iterables[1:] {
		if len(v) > size {
			size = len(v)
		}
	}

	results := [][]int{}

	for i := 0; i < size; i += 1 {
		newresult := make([]int, len(iterables))
		for j, v := range iterables {
			if i < len(v) {
				newresult[j] = v[i][0]
			} else {
				newresult[j] = fillvalue
			}

		}

		results = append(results, newresult)

	}

	return results
}

func (p *SeedProgress) canSkip(old_progress, current_progress [][2]int) bool {
	if current_progress == nil {
		return false
	}
	if old_progress == nil {
		return false
	}
	if len(old_progress) == 0 {
		return true
	}
	zips := iziplongest(-1, old_progress, current_progress)
	for i := range zips {
		old := zips[i][0]
		current := zips[i][1]
		if old == -1 {
			return false
		}
		if current == -1 {
			return false
		}
		if old < current {
			return false
		}
		if old > current {
			return true
		}
	}
	return false
}

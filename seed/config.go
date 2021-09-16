package seed

import "github.com/flywave/go-geo"

type LevelsList []int

func (ll LevelsList) ForGrid(grid geo.TileGrid) []int {
	ret := make([]int, 0)
	for _, l := range ll {
		if l >= 0 && l <= (int(grid.Levels)-1) {
			ret = append(ret, l)
		}
	}
	return ret
}

type LevelsRange [2]int

func (ll LevelsRange) ForGrid(grid geo.TileGrid) []int {
	start, stop := ll[0], ll[1]
	if start == -1 {
		start = 0
	}
	if stop == -1 {
		stop = 999
	}

	stop = geo.MinInt(stop, int(grid.Levels)-1)
	ret := make([]int, 0)
	for i := start; i < stop+1; i++ {
		ret = append(ret, i)
	}

	return ret
}

type LevelsResolutionRange [2]int

func (ls LevelsResolutionRange) ForGrid(grid geo.TileGrid) []int {
	start, stop := ls[0], ls[1]
	if start == -1 {
		start = 0
	} else {
		start = grid.ClosestLevel(float64(start))
	}

	if stop == -1 {
		stop = int(grid.Levels) - 1
	} else {
		stop = grid.ClosestLevel(float64(stop))
	}
	ret := make([]int, 0)
	for i := start; i < stop+1; i++ {
		ret = append(ret, i)
	}

	return ret
}

type LevelsResolutionList []int

func (ls LevelsResolutionList) ForGrid(grid geo.TileGrid) []int {
	ret := make([]int, 0)
	for _, res := range ls {
		i := grid.ClosestLevel(float64(res))
		ret = append(ret, i)
	}

	return ret
}

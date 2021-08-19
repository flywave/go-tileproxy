package seed

import (
	"testing"

	"github.com/flywave/go-tileproxy/geo"
)

func assertLevelInLevels(aa int, b []int, t *testing.T) {
	flag := false
	for i := range b {
		bb := b[i]

		if aa == bb {
			flag = true
		}
	}

	if !flag {
		t.FailNow()
	}
}

func TestLevelsList(t *testing.T) {
	levels := LevelsList{-10, 3, 1, 3, 5, 7, 50}

	res := levels.ForGrid(*geo.TileGridForEpsg("EPSG:4326", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{1, 3, 5, 7}, t)
	}
}

func TestLevelsRange(t *testing.T) {
	levels := LevelsRange{1, 5}

	res := levels.ForGrid(*geo.TileGridForEpsg("EPSG:4326", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{1, 2, 3, 4, 5}, t)
	}

	levels = LevelsRange{-1, 5}

	res = levels.ForGrid(*geo.TileGridForEpsg("EPSG:4326", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{0, 1, 2, 3, 4, 5}, t)
	}

	levels = LevelsRange{13, -1}

	res = levels.ForGrid(*geo.TileGridForEpsg("EPSG:4326", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{13, 14, 15, 16, 17, 18, 19}, t)
	}
}

func TestLevelsResolutionRange(t *testing.T) {
	levels := LevelsResolutionRange{1000, 100}

	res := levels.ForGrid(*geo.TileGridForEpsg("EPSG:900913", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{8, 9, 10, 11}, t)
	}

	levels = LevelsResolutionRange{-1, 100}

	res = levels.ForGrid(*geo.TileGridForEpsg("EPSG:900913", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11}, t)
	}

	levels = LevelsResolutionRange{1000, -1}

	res = levels.ForGrid(*geo.TileGridForEpsg("EPSG:900913", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{0,
			8,
			9,
			10,
			11,
			12,
			13,
			14,
			15,
			16,
			17,
			18,
			19}, t)
	}
}

func TestLevelsResolutionList(t *testing.T) {
	levels := LevelsResolutionList{1000, 100, 500}
	res := levels.ForGrid(*geo.TileGridForEpsg("EPSG:900913", nil, []uint32{256, 256}, nil))

	for i := range res {
		assertLevelInLevels(res[i], []int{8, 9, 11}, t)
	}

}

package task

import (
	"os"
	"sort"
	"testing"
)

func TestProgressLogger(t *testing.T) {
	local := NewLocalProgressStore("./test.task", false)

	logger := NewDefaultProgressLogger(nil, false, true, local)

	if logger == nil {
		t.FailNow()
	}

	logger.SetCurrentTaskId("test")
	logger.LogMessage("hello")
}

func TestLocalProgressStore(t *testing.T) {
	store := NewLocalProgressStore("./test.task", false)

	store.Store("1", []int{0, 1})
	store.Store("2", []int{0, 1})

	store.flush()

	store = NewLocalProgressStore("./test.task", true)

	pr := store.Get("1").([]interface{})

	if len(pr) != 2 {
		t.FailNow()
	}

	os.Remove("./test.task")
}

func TestIziplongest(t *testing.T) {
	list1 := [][2]int{{0, 4}, {0, 4}, {2, 4}}
	list2 := [][2]int{{0, 4}, {0, 4}, {1, 4}, {2, 4}}

	arrs := izip_longest(nil, list1, list2)

	// izip_longest应该返回一个二维数组，而不是nil
	if arrs == nil {
		t.Error("izip_longest should return a non-nil slice")
		return
	}

	// 验证返回的数组长度
	expectedLength := 4 // 最长数组的长度
	if len(arrs) != expectedLength {
		t.Errorf("Expected length %d, got %d", expectedLength, len(arrs))
	}

	// 验证每个元素的结构
	for i, arr := range arrs {
		if len(arr) != 2 {
			t.Errorf("Expected 2 elements in sub-array at index %d, got %d", i, len(arr))
		}
	}

	// 验证前3个元素匹配
	for i := 0; i < 3; i++ {
		if arrs[i][0] != list1[i] {
			t.Errorf("Mismatch at [%d][0]: expected %v, got %v", i, list1[i], arrs[i][0])
		}
		if arrs[i][1] != list2[i] {
			t.Errorf("Mismatch at [%d][1]: expected %v, got %v", i, list2[i], arrs[i][1])
		}
	}

	// 验证最后一个元素，list1应该用nil填充
	if arrs[3][0] != nil {
		t.Errorf("Expected nil for missing element, got %v", arrs[3][0])
	}
	if arrs[3][1] != list2[3] {
		t.Errorf("Expected %v, got %v", list2[3], arrs[3][1])
	}
}

func TestLevels(t *testing.T) {
	levels := []int{4, 2, 3, 1, 0, 5}
	sort.Ints(levels)

	if levels[0] != 0 {
		t.FailNow()
	}

	levels = levels[1:]

	if levels[0] != 1 {
		t.FailNow()
	}

}

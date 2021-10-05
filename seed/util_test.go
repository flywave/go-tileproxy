package seed

import (
	"os"
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

	store.Save()

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

	for i := range arrs {
		arr1 := arrs[i]
		if arr1 != nil {
			t.FailNow()
		}
	}
}

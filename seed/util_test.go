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

	logger.SetCurrentTaskID("test")
	logger.LogMessage("hello")
}

func TestLocalProgressStore(t *testing.T) {
	store := NewLocalProgressStore("./test.task", false)

	store.Store("1", [][2]int{{0, 1}, {1, 1}})
	store.Store("2", [][2]int{{0, 1}, {1, 1}})

	store.Save()

	store = NewLocalProgressStore("./test.task", true)

	pr := store.Get("1")

	if len(pr) != 2 {
		t.FailNow()
	}

	os.Remove("./test.task")
}

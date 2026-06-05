package task

import (
	"sync"
	"testing"

	"github.com/flywave/go-tileproxy/utils/flock"
)

func TestDummyCacheLocker(t *testing.T) {
	locker := &DummyCacheLocker{}
	called := false
	err := locker.Lock("test", func() {
		called = true
	})
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if !called {
		t.Error("run function was not called")
	}
}

func TestLocalCacheLocker_LockUnlock(t *testing.T) {
	locker := &LocalCacheLocker{
		locks: make(map[string]*flock.Flock),
	}
	called := false
	err := locker.Lock("test-key", func() {
		called = true
	})
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if !called {
		t.Error("run function was not called")
	}
}

func TestLocalCacheLocker_ConcurrentAccess(t *testing.T) {
	locker := &LocalCacheLocker{
		locks: make(map[string]*flock.Flock),
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	runCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := locker.Lock("concurrent-key", func() {
				mu.Lock()
				runCount++
				mu.Unlock()
			})
			if err != nil {
				t.Errorf("Lock failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if runCount != 10 {
		t.Errorf("Expected 10 runs, got %d", runCount)
	}
}

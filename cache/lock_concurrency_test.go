package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// concurrent Lock calls on the same tile must be serialized: only one
// holder may run the critical section at a time
func TestInMemoryTileLocker_SerializesSameTile(t *testing.T) {
	locker := NewInMemoryTileLocker()

	var inside int64
	var violations int64
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := locker.Lock(context.Background(), NewTile([3]int{7, 8, 9}), func() error {
				if atomic.AddInt64(&inside, 1) > 1 {
					atomic.AddInt64(&violations, 1)
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&inside, -1)
				return nil
			})
			if err != nil {
				t.Errorf("lock failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if violations > 0 {
		t.Fatalf("%d concurrent entries detected on the same tile lock", violations)
	}
}

// different tiles must not block each other
func TestInMemoryTileLocker_ParallelDifferentTiles(t *testing.T) {
	locker := NewInMemoryTileLocker()

	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				locker.Lock(context.Background(), NewTile([3]int{i, 0, 0}), func() error {
					time.Sleep(20 * time.Millisecond)
					return nil
				})
			}(i)
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("different tiles appear to serialize; lock scope is too coarse")
	}
}

// the file locker must drop its flock entries once the last user releases
func TestFileTileLocker_ReleasesEntries(t *testing.T) {
	locker := NewFileTileLocker(t.TempDir()+"/lock", 10*time.Millisecond).(*FileTileLocker)

	for i := 0; i < 50; i++ {
		err := locker.Lock(context.Background(), NewTile([3]int{i, 0, 0}), func() error { return nil })
		if err != nil {
			t.Fatalf("lock failed: %v", err)
		}
	}

	if len(locker.lockMap) != 0 {
		t.Fatalf("lockMap leaked %d entries", len(locker.lockMap))
	}
}

package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryTileLocker_LockUnlock(t *testing.T) {
	locker := NewInMemoryTileLocker()
	tile := NewTile([3]int{1, 2, 3})

	ctx := context.Background()
	called := false
	err := locker.Lock(ctx, tile, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if !called {
		t.Error("run function was not called")
	}
}

func TestInMemoryTileLocker_ConcurrentAccess(t *testing.T) {
	locker := NewInMemoryTileLocker()
	tile := NewTile([3]int{1, 2, 3})

	var wg sync.WaitGroup
	concurrent := 10
	runCount := 0
	var mu sync.Mutex

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			err := locker.Lock(ctx, tile, func() error {
				mu.Lock()
				runCount++
				mu.Unlock()
				return nil
			})
			if err != nil {
				t.Errorf("Lock failed: %v", err)
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	if runCount != concurrent {
		t.Errorf("Expected %d runs, got %d", concurrent, runCount)
	}
	mu.Unlock()
}

func TestInMemoryTileLocker_ContextCancellation(t *testing.T) {
	locker := NewInMemoryTileLocker()
	tile := NewTile([3]int{1, 2, 3})

	// First lock to hold the lock
	holdCtx, holdCancel := context.WithCancel(context.Background())
	defer holdCancel()
	holdDone := make(chan struct{})
	go func() {
		locker.Lock(holdCtx, tile, func() error {
			<-holdCtx.Done()
			return nil
		})
		close(holdDone)
	}()

	// Ensure the first lock is acquired
	time.Sleep(10 * time.Millisecond)

	// Second lock attempt with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := locker.Lock(cancelCtx, tile, func() error {
		t.Error("run should not be called on cancelled context")
		return nil
	})
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}

	// Cleanup
	holdCancel()
	<-holdDone
}

func TestInMemoryTileLocker_ContextCancellationWhileWaiting(t *testing.T) {
	locker := NewInMemoryTileLocker()
	tile := NewTile([3]int{1, 2, 3})

	// Hold the lock
	holdCtx, holdCancel := context.WithCancel(context.Background())
	defer holdCancel()
	holdDone := make(chan struct{})
	go func() {
		locker.Lock(holdCtx, tile, func() error {
			<-holdCtx.Done()
			return nil
		})
		close(holdDone)
	}()

	time.Sleep(10 * time.Millisecond)

	// Try to acquire lock and cancel while waiting
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := locker.Lock(ctx, tile, func() error {
		t.Error("run should not be called")
		return nil
	})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	holdCancel()
	<-holdDone
}

func TestFileTileLocker_LockUnlock(t *testing.T) {
	retryDelay := 10 * time.Millisecond
	locker := NewFileTileLocker(t.TempDir()+"/tilelock", retryDelay).(*FileTileLocker)
	tile := NewTile([3]int{1, 2, 3})

	ctx := context.Background()
	called := false
	err := locker.Lock(ctx, tile, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if !called {
		t.Error("run function was not called")
	}
}

func TestFileTileLocker_ReentrantLock(t *testing.T) {
	retryDelay := 10 * time.Millisecond
	locker := NewFileTileLocker(t.TempDir()+"/tilelock", retryDelay).(*FileTileLocker)
	tile := NewTile([3]int{1, 2, 3})

	ctx := context.Background()
	// First lock
	err := locker.Lock(ctx, tile, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("First lock failed: %v", err)
	}

	// Second lock on same tile should work (previous lock was released)
	err = locker.Lock(ctx, tile, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("Second lock failed: %v", err)
	}
}

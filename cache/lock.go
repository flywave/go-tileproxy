package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/utils/flock"
)

type TileLocker interface {
	Lock(ctx context.Context, tile *Tile, run func() error) error
}

type FileTileLocker struct {
	TileLocker
	basePath   string
	retryDelay time.Duration
	lockMap    map[string]*flock.Flock
	mu         sync.Mutex
}

func NewFileTileLocker(basePath string, retryDelay time.Duration) TileLocker {
	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond
	}
	return &FileTileLocker{
		basePath:   basePath,
		retryDelay: retryDelay,
		lockMap:    make(map[string]*flock.Flock),
	}
}

func (l *FileTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	key := fmt.Sprintf("%d_%d_%d", tile.Coord[0], tile.Coord[1], tile.Coord[2])
	lockPath := l.basePath + "-" + key + ".lock"

	l.mu.Lock()
	f, exists := l.lockMap[key]
	if !exists {
		f = flock.NewFlock(lockPath)
		l.lockMap[key] = f
	}
	l.mu.Unlock()

	locked, err := f.TryLockContext(ctx, l.retryDelay)
	if err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}
	if !locked {
		return ctx.Err()
	}

	return run()
}

type InMemoryTileLocker struct {
	TileLocker
	mu      sync.Mutex
	locks   map[string]struct{}
	waitChs map[string][]chan struct{}
}

func NewInMemoryTileLocker() TileLocker {
	return &InMemoryTileLocker{
		locks:   make(map[string]struct{}),
		waitChs: make(map[string][]chan struct{}),
	}
}

func (l *InMemoryTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	key := fmt.Sprintf("%d_%d_%d", tile.Coord[0], tile.Coord[1], tile.Coord[2])

	l.mu.Lock()
	_, locked := l.locks[key]
	if locked {
		ch := make(chan struct{})
		l.waitChs[key] = append(l.waitChs[key], ch)
		l.mu.Unlock()

		select {
		case <-ch:
		case <-ctx.Done():
			l.mu.Lock()
			channels := l.waitChs[key]
			for i, c := range channels {
				if c == ch {
					l.waitChs[key] = append(channels[:i], channels[i+1:]...)
					break
				}
			}
			l.mu.Unlock()
			return ctx.Err()
		}

		l.mu.Lock()
		if _, locked := l.locks[key]; locked {
			l.mu.Unlock()
			return fmt.Errorf("inconsistent lock state for key %s", key)
		}
	}
	l.locks[key] = struct{}{}
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		delete(l.locks, key)
		for _, ch := range l.waitChs[key] {
			close(ch)
		}
		delete(l.waitChs, key)
		l.mu.Unlock()
	}()

	return run()
}

type DummyTileLocker struct {
}

func (l *DummyTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	return run()
}

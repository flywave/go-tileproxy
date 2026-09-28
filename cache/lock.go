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

type flockEntry struct {
	f    *flock.Flock
	refs int
}

type FileTileLocker struct {
	TileLocker
	basePath   string
	retryDelay time.Duration
	lockMap    map[string]*flockEntry
	mu         sync.Mutex
}

func NewFileTileLocker(basePath string, retryDelay time.Duration) TileLocker {
	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond
	}
	return &FileTileLocker{
		basePath:   basePath,
		retryDelay: retryDelay,
		lockMap:    make(map[string]*flockEntry),
	}
}

func (l *FileTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	key := fmt.Sprintf("%d_%d_%d", tile.Coord[0], tile.Coord[1], tile.Coord[2])
	lockPath := l.basePath + "-" + key + ".lock"

	l.mu.Lock()
	entry, exists := l.lockMap[key]
	if !exists {
		entry = &flockEntry{f: flock.NewFlock(lockPath)}
		l.lockMap[key] = entry
	}
	entry.refs++
	l.mu.Unlock()

	release := func() {
		l.mu.Lock()
		entry.refs--
		if entry.refs <= 0 {
			// drop the entry so the map does not grow with every tile
			delete(l.lockMap, key)
		}
		l.mu.Unlock()
	}

	locked, err := entry.f.TryLockContext(ctx, l.retryDelay)
	if err != nil {
		release()
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}
	if !locked {
		release()
		return ctx.Err()
	}

	defer func() {
		entry.f.Unlock()
		release()
	}()

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
	if _, locked := l.locks[key]; locked {
		ch := make(chan struct{})
		l.waitChs[key] = append(l.waitChs[key], ch)
		l.mu.Unlock()

		select {
		case <-ch:
			l.mu.Lock()
		case <-ctx.Done():
			l.mu.Lock()
			channels := l.waitChs[key]
			for i, c := range channels {
				if c == ch {
					l.waitChs[key] = append(channels[:i], channels[i+1:]...)
					break
				}
			}
			select {
			case <-ch:
			default:
				l.mu.Unlock()
				return ctx.Err()
			}
		}
	}
	l.locks[key] = struct{}{}
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		delete(l.locks, key)
		if chans := l.waitChs[key]; len(chans) > 0 {
			// wake exactly the first waiter; waking all of them would let
			// every one of them take the lock at once
			close(chans[0])
			l.waitChs[key] = chans[1:]
		} else {
			delete(l.waitChs, key)
		}
		l.mu.Unlock()
	}()

	return run()
}

type DummyTileLocker struct {
}

func (l *DummyTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	return run()
}

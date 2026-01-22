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
	f          *flock.Flock
	retryDelay time.Duration
}

func NewFileTileLocker(path string, retryDelay time.Duration) TileLocker {
	return &FileTileLocker{f: flock.NewFlock(path), retryDelay: retryDelay}
}

func (l *FileTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for {
		if err := run(); err != nil {
			return err
		}
		return nil
	}
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
		case <-ctx.Done():
			return ctx.Err()
		}

		l.mu.Lock()
		defer l.mu.Unlock()
	} else {
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
	}

	return run()
}

type DummyTileLocker struct {
}

func (l *DummyTileLocker) Lock(ctx context.Context, tile *Tile, run func() error) error {
	return run()
}

package cache

import (
	"context"
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.retryDelay):
			// try again
		}
	}
}

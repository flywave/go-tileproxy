package flock

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"
)

type Flock struct {
	path string
	m    sync.RWMutex
	fh   *os.File
	l    bool
	r    bool
}

func New(path string) *Flock {
	return &Flock{path: path}
}

func NewFlock(path string) *Flock {
	return New(path)
}

func (f *Flock) Close() error {
	return f.Unlock()
}

func (f *Flock) Path() string {
	return f.path
}

func (f *Flock) Locked() bool {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.l
}

func (f *Flock) RLocked() bool {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.r
}

func (f *Flock) String() string {
	return f.path
}

func (f *Flock) TryLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, f.TryLock, retryDelay)
}

func (f *Flock) TryRLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, f.TryRLock, retryDelay)
}

func tryCtx(ctx context.Context, fn func() (bool, error), retryDelay time.Duration) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	for {
		if ok, err := fn(); ok || err != nil {
			return ok, err
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryDelay):
			// try again
		}
	}
}

func (f *Flock) setFh() error {
	flags := os.O_CREATE
	if runtime.GOOS == "aix" {
		flags |= os.O_RDWR
	} else {
		flags |= os.O_RDONLY
	}
	fh, err := os.OpenFile(f.path, flags, os.FileMode(0600))
	if err != nil {
		return err
	}

	f.fh = fh
	return nil
}

func (f *Flock) ensureFhState() {
	if !f.l && !f.r && f.fh != nil {
		f.fh.Close()
		f.fh = nil
	}
}

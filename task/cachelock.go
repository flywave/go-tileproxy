package task

import (
	"fmt"
	"os"
	"sync"

	"github.com/flywave/go-tileproxy/utils/flock"
)

type CacheLocker interface {
	Lock(name string, runf func()) error
}

type DummyCacheLocker struct {
	CacheLocker
}

func (l *DummyCacheLocker) Lock(name string, runf func()) error {
	runf()
	return nil
}

type LocalCacheLocker struct {
	CacheLocker
	locks map[string]*flock.Flock
	m     sync.RWMutex
}

func (l *LocalCacheLocker) getTempFileName(name string) string {
	tmpFileFh, _ := os.CreateTemp(os.TempDir(), fmt.Sprintf("cache-flock-%s-", name))
	tmpFileFh.Close()
	tmpFile := tmpFileFh.Name()
	os.Remove(tmpFile)
	return tmpFile
}

func (l *LocalCacheLocker) Lock(name string, runf func()) error {
	l.m.RLock()
	var lock *flock.Flock
	if lock_, ok := l.locks[name]; !ok {
		filaname := l.getTempFileName(name)
		lock = flock.NewFlock(filaname)
		l.locks[name] = lock
	} else {
		lock = lock_
	}
	l.m.RUnlock()

	err := lock.Lock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
	runf()
	return nil
}

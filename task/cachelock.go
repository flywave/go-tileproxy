package task

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/flywave/go-tileproxy/utils/flock"
)

type CacheLocker interface {
	Lock(name string, run func() error) error
}

type DummyCacheLocker struct {
	CacheLocker
}

func (l *DummyCacheLocker) Lock(name string, run func() error) error {
	return run()
}

type LocalCacheLocker struct {
	CacheLocker
	locks map[string]*flock.Flock
	m     sync.RWMutex
}

func (l *LocalCacheLocker) getTempFileName(name string) string {
	tmpFileFh, _ := ioutil.TempFile(os.TempDir(), fmt.Sprintf("cache-flock-%s-", name))
	tmpFileFh.Close()
	tmpFile := tmpFileFh.Name()
	os.Remove(tmpFile)
	return tmpFile
}

func (l *LocalCacheLocker) Lock(name string, run func() error) error {
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

	lock.Lock()
	defer lock.Unlock()
	return run()
}

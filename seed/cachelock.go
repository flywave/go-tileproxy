package seed

type CacheLocker interface {
	Lock(name string, run func() error) error
}

type DummyCacheLocker struct {
	CacheLocker
}

func (l *DummyCacheLocker) Lock(name string, run func() error) error {
	return run()
}

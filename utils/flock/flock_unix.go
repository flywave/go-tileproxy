//go:build !aix && !windows
// +build !aix,!windows

package flock

import (
	"os"
	"syscall"
)

func (f *Flock) Lock() error {
	return f.lock(&f.l, syscall.LOCK_EX)
}

func (f *Flock) RLock() error {
	return f.lock(&f.r, syscall.LOCK_SH)
}

func (f *Flock) lock(locked *bool, flag int) error {
	f.m.Lock()
	defer f.m.Unlock()

	if *locked {
		return nil
	}

	if f.fh == nil {
		if err := f.setFh(); err != nil {
			return err
		}
		defer f.ensureFhState()
	}

	if err := syscall.Flock(int(f.fh.Fd()), flag); err != nil {
		shouldRetry, reopenErr := f.reopenFDOnError(err)
		if reopenErr != nil {
			return reopenErr
		}

		if !shouldRetry {
			return err
		}

		if err = syscall.Flock(int(f.fh.Fd()), flag); err != nil {
			return err
		}
	}

	*locked = true
	return nil
}

func (f *Flock) Unlock() error {
	f.m.Lock()
	defer f.m.Unlock()

	if (!f.l && !f.r) || f.fh == nil {
		return nil
	}

	if err := syscall.Flock(int(f.fh.Fd()), syscall.LOCK_UN); err != nil {
		return err
	}

	f.fh.Close()

	f.l = false
	f.r = false
	f.fh = nil

	return nil
}

func (f *Flock) TryLock() (bool, error) {
	return f.try(&f.l, syscall.LOCK_EX)
}

func (f *Flock) TryRLock() (bool, error) {
	return f.try(&f.r, syscall.LOCK_SH)
}

func (f *Flock) try(locked *bool, flag int) (bool, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if *locked {
		return true, nil
	}

	if f.fh == nil {
		if err := f.setFh(); err != nil {
			return false, err
		}
		defer f.ensureFhState()
	}

	var retried bool
retry:
	err := syscall.Flock(int(f.fh.Fd()), flag|syscall.LOCK_NB)

	switch err {
	case syscall.EWOULDBLOCK:
		return false, nil
	case nil:
		*locked = true
		return true, nil
	}
	if !retried {
		if shouldRetry, reopenErr := f.reopenFDOnError(err); reopenErr != nil {
			return false, reopenErr
		} else if shouldRetry {
			retried = true
			goto retry
		}
	}

	return false, err
}

func (f *Flock) reopenFDOnError(err error) (bool, error) {
	if err != syscall.EIO && err != syscall.EBADF {
		return false, nil
	}
	if st, err := f.fh.Stat(); err == nil {
		if st.Mode()&0600 == 0600 {
			f.fh.Close()
			f.fh = nil

			fh, err := os.OpenFile(f.path, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
			if err != nil {
				return false, err
			}
			f.fh = fh
			return true, nil
		}
	}

	return false, nil
}

package flock

import (
	"syscall"
)

const ErrorLockViolation syscall.Errno = 0x21 // 33

func (f *Flock) Lock() error {
	return f.lock(&f.l, winLockfileExclusiveLock)
}

func (f *Flock) RLock() error {
	return f.lock(&f.r, winLockfileSharedLock)
}

func (f *Flock) lock(locked *bool, flag uint32) error {
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

	if _, errNo := lockFileEx(syscall.Handle(f.fh.Fd()), flag, 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		return errNo
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

	if _, errNo := unlockFileEx(syscall.Handle(f.fh.Fd()), 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		return errNo
	}

	f.fh.Close()

	f.l = false
	f.r = false
	f.fh = nil

	return nil
}

func (f *Flock) TryLock() (bool, error) {
	return f.try(&f.l, winLockfileExclusiveLock)
}

func (f *Flock) TryRLock() (bool, error) {
	return f.try(&f.r, winLockfileSharedLock)
}

func (f *Flock) try(locked *bool, flag uint32) (bool, error) {
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

	_, errNo := lockFileEx(syscall.Handle(f.fh.Fd()), flag|winLockfileFailImmediately, 0, 1, 0, &syscall.Overlapped{})

	if errNo > 0 {
		if errNo == ErrorLockViolation || errNo == syscall.ERROR_IO_PENDING {
			return false, nil
		}

		return false, errNo
	}

	*locked = true

	return true, nil
}

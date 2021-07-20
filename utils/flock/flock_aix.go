//+build aix

package flock

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type lockType int16

const (
	readLock  lockType = unix.F_RDLCK
	writeLock lockType = unix.F_WRLCK
)

type cmdType int

const (
	tryLock  cmdType = unix.F_SETLK
	waitLock cmdType = unix.F_SETLKW
)

type inode = uint64

type inodeLock struct {
	owner *Flock
	queue []<-chan *Flock
}

var (
	mu     sync.Mutex
	inodes = map[*Flock]inode{}
	locks  = map[inode]inodeLock{}
)

func (f *Flock) Lock() error {
	return f.lock(&f.l, writeLock)
}

func (f *Flock) RLock() error {
	return f.lock(&f.r, readLock)
}

func (f *Flock) lock(locked *bool, flag lockType) error {
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

	if _, err := f.doLock(waitLock, flag, true); err != nil {
		return err
	}

	*locked = true
	return nil
}

func (f *Flock) doLock(cmd cmdType, lt lockType, blocking bool) (bool, error) {
	fi, err := f.fh.Stat()
	if err != nil {
		return false, err
	}
	ino := inode(fi.Sys().(*syscall.Stat_t).Ino)

	mu.Lock()
	if i, dup := inodes[f]; dup && i != ino {
		mu.Unlock()
		return false, &os.PathError{
			Path: f.Path(),
			Err:  errors.New("inode for file changed since last Lock or RLock"),
		}
	}

	inodes[f] = ino

	var wait chan *Flock
	l := locks[ino]
	if l.owner == f {
		//
	} else if l.owner == nil {
		l.owner = f
	} else if !blocking {
		mu.Unlock()
		return false, nil
	} else {
		wait = make(chan *Flock)
		l.queue = append(l.queue, wait)
	}
	locks[ino] = l
	mu.Unlock()

	if wait != nil {
		wait <- f
	}

	err = setlkw(f.fh.Fd(), cmd, lt)

	if err != nil {
		f.doUnlock()
		if cmd == tryLock && err == unix.EACCES {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (f *Flock) Unlock() error {
	f.m.Lock()
	defer f.m.Unlock()

	if (!f.l && !f.r) || f.fh == nil {
		return nil
	}

	if err := f.doUnlock(); err != nil {
		return err
	}

	f.fh.Close()

	f.l = false
	f.r = false
	f.fh = nil

	return nil
}

func (f *Flock) doUnlock() (err error) {
	var owner *Flock
	mu.Lock()
	ino, ok := inodes[f]
	if ok {
		owner = locks[ino].owner
	}
	mu.Unlock()

	if owner == f {
		err = setlkw(f.fh.Fd(), waitLock, unix.F_UNLCK)
	}

	mu.Lock()
	l := locks[ino]
	if len(l.queue) == 0 {
		delete(locks, ino)
	} else {
		l.owner = <-l.queue[0]
		l.queue = l.queue[1:]
		locks[ino] = l
	}
	delete(inodes, f)
	mu.Unlock()

	return err
}

func (f *Flock) TryLock() (bool, error) {
	return f.try(&f.l, writeLock)
}

func (f *Flock) TryRLock() (bool, error) {
	return f.try(&f.r, readLock)
}

func (f *Flock) try(locked *bool, flag lockType) (bool, error) {
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

	haslock, err := f.doLock(tryLock, flag, false)
	if err != nil {
		return false, err
	}

	*locked = haslock
	return haslock, nil
}

func setlkw(fd uintptr, cmd cmdType, lt lockType) error {
	for {
		err := unix.FcntlFlock(fd, int(cmd), &unix.Flock_t{
			Type:   int16(lt),
			Whence: io.SeekStart,
			Start:  0,
			Len:    0,
		})
		if err != unix.EINTR {
			return err
		}
	}
}

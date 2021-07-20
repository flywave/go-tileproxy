package flock_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flywave/go-tileproxy/utils/flock"
)

func ExampleFlock_Locked() {
	f := flock.New(os.TempDir() + "/go-lock.lock")
	f.TryLock()

	fmt.Printf("locked: %v\n", f.Locked())

	f.Unlock()

	fmt.Printf("locked: %v\n", f.Locked())
}

func ExampleFlock_TryLock() {
	fileLock := flock.New(os.TempDir() + "/go-lock.lock")

	locked, err := fileLock.TryLock()

	if err != nil {
		//
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())

		if err := fileLock.Unlock(); err != nil {
			//
		}
	}

	fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())
}

func ExampleFlock_TryLockContext() {
	fileLock := flock.New(os.TempDir() + "/go-lock.lock")

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, 678*time.Millisecond)

	if err != nil {
		//
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())

		if err := fileLock.Unlock(); err != nil {
			//
		}
	}

	fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())
}

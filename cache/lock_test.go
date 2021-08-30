package cache

import (
	"context"
	"testing"
	"time"
)

func TestFileTileLocker(t *testing.T) {
	retryDelay := time.Duration(20 * time.Second)
	lock := NewFileTileLocker("./tilelock", retryDelay)

	ctx, cancel := context.WithCancel(context.Background())

	lock.Lock(ctx, nil, func() error {
		cancel()
		return nil
	})
}

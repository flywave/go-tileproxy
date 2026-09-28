package tileproxy

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/setting"
)

// ConfigWatcher watches a configuration file for changes and hot-reloads
// the proxy. It polls the file metadata and confirms a change by comparing
// content hashes, which keeps it reliable across editors that save via
// rename (a pattern inotify-style watchers easily miss) and needs no
// platform specific dependencies. Reload failures are logged and leave the
// running services untouched.
type ConfigWatcher struct {
	proxy    *TileProxy
	path     string
	fac      setting.CacheFactory
	interval time.Duration
	debounce time.Duration

	initialHash string

	stop    chan struct{}
	stopped chan struct{}
	once    sync.Once
}

func NewConfigWatcher(proxy *TileProxy, path string, fac setting.CacheFactory, interval, debounce time.Duration) *ConfigWatcher {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if debounce <= 0 {
		debounce = 1 * time.Second
	}
	w := &ConfigWatcher{
		proxy:    proxy,
		path:     path,
		fac:      fac,
		interval: interval,
		debounce: debounce,
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
	// capture the baseline synchronously so that changes made right after
	// Start() can never be swallowed by a late initial read
	w.initialHash = w.contentHash()
	return w
}

// Start begins watching in a background goroutine.
func (w *ConfigWatcher) Start() {
	go w.loop()
}

// Stop terminates the watcher. It is safe to call multiple times.
func (w *ConfigWatcher) Stop() {
	w.once.Do(func() {
		close(w.stop)
		<-w.stopped
	})
}

func (w *ConfigWatcher) loop() {
	defer close(w.stopped)

	hash := w.initialHash
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
		}

		current := w.contentHash()
		if current == "" || current == hash {
			continue
		}

		// the file changed: wait for writes to settle before reloading so a
		// save split across several writes triggers a single reload
		stable := current
		deadline := time.Now().Add(w.debounce)
		for time.Now().Before(deadline) {
			select {
			case <-w.stop:
				return
			case <-time.After(w.debounce / 4):
			}
			next := w.contentHash()
			if next == "" {
				continue
			}
			if next == stable {
				break
			}
			stable = next
			deadline = time.Now().Add(w.debounce)
		}

		hash = stable
		log.Printf("config %s changed, reloading", w.path)
		if err := w.proxy.ReloadFromFile(w.path, w.fac); err != nil {
			log.Printf("hot reload failed, keeping current services: %v", err)
			continue
		}
		log.Printf("hot reload complete: %v", w.proxy.GetServiceIDs())
	}
}

// contentHash returns a hash of the file contents, or "" when the file
// cannot be read (e.g. mid-write).
func (w *ConfigWatcher) contentHash() string {
	data, err := os.ReadFile(w.path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

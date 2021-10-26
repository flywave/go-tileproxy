package task

import (
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/exports"
	"github.com/flywave/go-tileproxy/imports"
)

type Work interface {
	Run()
	Done() <-chan struct{}
}

type SeedWorker struct {
	Work
	task    Task
	manager cache.Manager
	tiles   [][3]int
	err     error
	done    chan struct{}
}

func (w *SeedWorker) Done() <-chan struct{} {
	return w.done
}

func (w *SeedWorker) Run() {
	_, err := w.manager.LoadTileCoords(w.tiles, nil, false)

	if err != nil {
		w.err = err
	}

	close(w.done)
}

type CleanupWorker struct {
	Work
	task    Task
	manager cache.Manager
	tiles   [][3]int
	err     error
	done    chan struct{}
}

func (w *CleanupWorker) Done() <-chan struct{} {
	return w.done
}

func (w *CleanupWorker) Run() {
	err := w.manager.RemoveTileCoords(w.tiles)

	if err != nil {
		w.err = err
	}

	close(w.done)
}

type ExportWorker struct {
	Work
	task    Task
	io      exports.Export
	manager cache.Manager
	tiles   [][3]int
	err     error
	done    chan struct{}
}

func (w *ExportWorker) Done() <-chan struct{} {
	return w.done
}

func (w *ExportWorker) Run() {
	tc, err := w.manager.LoadTileCoords(w.tiles, nil, false)

	if err != nil {
		w.err = err
		return
	}

	err = w.io.StoreTileCollection(tc, w.manager.GetGrid())

	if err != nil {
		w.err = err
		return
	}

	close(w.done)
}

type ImportWorker struct {
	Work
	task            Task
	io              imports.Import
	manager         cache.Manager
	tiles           [][3]int
	err             error
	done            chan struct{}
	force_overwrite bool
}

func (w *ImportWorker) Done() <-chan struct{} {
	return w.done
}

func (w *ImportWorker) Run() {
	tc, err := w.io.LoadTileCoords(w.tiles, w.manager.GetGrid())

	if err != nil {
		w.err = err
		return
	}

	if tc.Empty() {
		return
	}

	if w.force_overwrite {
		err = w.manager.StoreTiles(tc)

		if err != nil {
			w.err = err
		}
	} else {
		for _, t := range tc.GetSlice() {
			if !w.manager.IsCached(t.Coord, nil) {
				err = w.manager.StoreTile(t)
				if err != nil {
					w.err = err
					break
				}
			}
		}
	}

	close(w.done)
}

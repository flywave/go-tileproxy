package task

import (
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/exports"
	"github.com/flywave/go-tileproxy/imports"
)

type Work interface {
	Run()
}

type SeedWorker struct {
	Work
	task    Task
	manager cache.Manager
	tiles   [][3]int
	err     error
}

func (w *SeedWorker) Run() {
	_, err := w.manager.LoadTileCoords(w.tiles, nil, false)

	if err != nil {
		w.err = err
	}
}

type CleanupWorker struct {
	Work
	task    Task
	manager cache.Manager
	tiles   [][3]int
	err     error
}

func (w *CleanupWorker) Run() {
	err := w.manager.RemoveTileCoords(w.tiles)

	if err != nil {
		w.err = err
	}
}

type ExportWorker struct {
	Work
	task    Task
	io      exports.Export
	manager cache.Manager
	tiles   [][3]int
	err     error
}

func (w *ExportWorker) Run() {
	tc, err := w.manager.LoadTileCoords(w.tiles, nil, false)

	if err != nil {
		w.err = err
		return
	}

	err = w.io.StoreTileCollection(tc)

	if err != nil {
		w.err = err
		return
	}
}

type ImportWorker struct {
	Work
	task    Task
	io      imports.Import
	manager cache.Manager
	tiles   [][3]int
	err     error
}

func (w *ImportWorker) Run() {
	tc, err := w.io.LoadTileCoords(w.tiles)

	if err != nil {
		w.err = err
		return
	}

	if tc.Empty() {
		return
	}

	err = w.manager.StoreTiles(tc)

	if err != nil {
		w.err = err
	}
}

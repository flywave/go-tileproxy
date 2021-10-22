package task

import (
	"errors"
	"sync"

	"github.com/flywave/go-tileproxy/utils"
)

type workerQueue struct {
	Threads int
	wake    chan struct{}
	mut     sync.Mutex
	running bool
	storage *utils.Deque
}

func newWorkerQueue(threads int) *workerQueue {
	return &workerQueue{
		Threads: threads,
		running: true,
		storage: utils.NewDeque(20),
	}
}

func (q *workerQueue) IsEmpty() bool {
	return q.Size() == 0
}

func (q *workerQueue) AddRequest(tiles Work) error {
	q.mut.Lock()
	waken := q.wake != nil
	q.mut.Unlock()
	if !waken {
		return q.storeRequest(tiles)
	}
	err := q.storeRequest(tiles)
	if err != nil {
		return err
	}
	q.wake <- struct{}{}
	return nil
}

func (q *workerQueue) storeRequest(tiles Work) error {
	q.storage.PushBack(tiles)
	return nil
}

func (q *workerQueue) Size() int {
	return q.storage.Len()
}

func (q *workerQueue) IsRuning() bool {
	q.mut.Lock()
	defer q.mut.Unlock()
	return q.running
}

func (q *workerQueue) Run() error {
	q.mut.Lock()
	if q.wake != nil && q.running {
		q.mut.Unlock()
		panic("cannot call duplicate Queue.Run")
	}
	q.wake = make(chan struct{})
	q.running = true
	q.mut.Unlock()

	requestc := make(chan Work)
	complete, errc := make(chan struct{}), make(chan error, 1)
	for i := 0; i < q.Threads; i++ {
		go independentRunner(requestc, complete)
	}
	go q.loop(requestc, complete, errc)
	defer close(requestc)
	return <-errc
}

func (q *workerQueue) Stop() {
	q.mut.Lock()
	q.running = false
	q.mut.Unlock()
}

func (q *workerQueue) loop(requestc chan<- Work, complete <-chan struct{}, errc chan<- error) {
	var active int
	for {
		size := q.storage.Len()
		if size == 0 && active == 0 || !q.running {
			errc <- nil
			break
		}
		sent := requestc
		var req Work
		if size > 0 {
			var err error
			req, err = q.loadRequest()
			if err != nil {
				continue
			}
		} else {
			sent = nil
		}
	Sent:
		for {
			select {
			case sent <- req:
				active++
				break Sent
			case <-q.wake:
				if sent == nil {
					break Sent
				}
			case <-complete:
				active--
				if sent == nil && active == 0 {
					break Sent
				}
			}
		}
	}
}

func independentRunner(requestc <-chan Work, complete chan<- struct{}) {
	for req := range requestc {
		req.Run()
		complete <- struct{}{}
	}
}

func (q *workerQueue) loadRequest() (Work, error) {
	tiles, ok := q.storage.PopFront().(Work)
	if !ok {
		return nil, errors.New("storage error")
	}
	return tiles, nil
}

type WorkerPool interface {
	Process(tiles Work, progress *TaskProgress)
}

type TileWorkerPool struct {
	WorkerPool
	Queue  *workerQueue
	Logger ProgressLogger
	Task   Task
}

func NewTileWorkerPool(threads int, task Task, logger ProgressLogger) *TileWorkerPool {
	queue := newWorkerQueue(threads)
	go queue.Run()
	return &TileWorkerPool{Queue: queue, Logger: logger, Task: task}
}

func (p *TileWorkerPool) Process(tiles Work, progress *TaskProgress) {
	p.Queue.AddRequest(tiles)

	if p.Logger != nil {
		p.Logger.LogStep(progress)
	}
}

package client

import (
	"sync"

	"github.com/flywave/go-tileproxy/crawler"
)

type Queue struct {
	Threads int
	wake    chan struct{}
	mut     sync.Mutex
	running bool
	storage Storage
	q       sync.Mutex
}

func NewQueue(threads int, maxSize int) (*Queue, error) {
	s := &InMemoryQueueStorage{MaxSize: maxSize}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return &Queue{
		Threads: threads,
		running: true,
		storage: s,
	}, nil
}

func (q *Queue) IsEmpty() bool {
	s, _ := q.Size()
	return s == 0
}

func (q *Queue) AddRequest(r *crawler.Request) (*Future, error) {
	q.mut.Lock()
	waken := q.wake != nil
	q.mut.Unlock()
	if !waken {
		return q.storeRequest(r)
	}
	f, err := q.storeRequest(r)
	if err != nil {
		return nil, err
	}
	q.wake <- struct{}{}
	return f, nil
}

func (q *Queue) storeRequest(r *crawler.Request) (*Future, error) {
	d, err := r.Marshal()
	if err != nil {
		return nil, err
	}
	f := newFuture(d)
	q.q.Lock()
	q.storage.AddFuture(f)
	q.q.Unlock()
	return f, nil
}

func (q *Queue) Size() (int, error) {
	return q.storage.QueueSize()
}

func (q *Queue) Run(c *crawler.Collector) error {
	q.mut.Lock()
	if q.wake != nil && q.running {
		q.mut.Unlock()
		panic("cannot call duplicate Queue.Run")
	}
	q.wake = make(chan struct{})
	q.running = true
	q.mut.Unlock()

	requestc := make(chan *Future)
	complete, errc := make(chan struct{}), make(chan error, 1)
	for i := 0; i < q.Threads; i++ {
		go independentRunner(requestc, complete, errc)
	}
	go q.loop(c, requestc, complete, errc)
	defer close(requestc)
	return nil
}

func (q *Queue) Stop() {
	q.mut.Lock()
	q.running = false
	q.mut.Unlock()
}

func (q *Queue) loop(c *crawler.Collector, requestc chan<- *Future, complete <-chan struct{}, errc chan<- error) {
	var active int
	for {
		size, err := q.Size()
		if err != nil {
			break
		}
		if size == 0 && active == 0 || !q.running {
			break
		}
		sent := requestc
		var fut *Future
		if size > 0 {
			fut, err = q.loadFuture(c)
			if err != nil {
				continue
			}
		} else {
			sent = nil
		}
	Sent:
		for {
			select {
			case sent <- fut:
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

func independentRunner(requestc <-chan *Future, complete chan<- struct{}, errorchan chan<- error) {
	for req := range requestc {
		err := req.Do()

		if err != nil {
			errorchan <- err
			return
		}
		complete <- struct{}{}
	}
}

func (q *Queue) loadFuture(c *crawler.Collector) (*Future, error) {
	fut, err := q.storage.GetFuture()
	if err != nil {
		return nil, err
	}
	if fut != nil {
		fut.setCollector(c)
	}
	return fut, nil
}

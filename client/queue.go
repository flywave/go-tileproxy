package client

import (
	"sync"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/utils"
)

type Queue struct {
	Threads int
	wake    chan struct{}
	mut     sync.Mutex
	running bool
	storage *utils.Deque
	q       sync.Mutex
}

func NewQueue(threads int, maxSize int) (*Queue, error) {
	return &Queue{
		Threads: threads,
		running: true,
		storage: utils.NewDeque(maxSize),
	}, nil
}

func (q *Queue) IsEmpty() bool {
	return q.Size() == 0
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
	q.storage.PushBack(f)
	q.q.Unlock()
	return f, nil
}

func (q *Queue) Size() int {
	return q.storage.Cap()
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
		go independentRunner(requestc, complete)
	}
	go q.loop(c, requestc, complete, errc)
	defer close(requestc)
	return <-errc
}

func (q *Queue) Stop() {
	q.mut.Lock()
	q.running = false
	q.mut.Unlock()
}

func (q *Queue) loop(c *crawler.Collector, requestc chan<- *Future, complete <-chan struct{}, errc chan<- error) {
	var active int
	for {
		size := q.Size()
		if size == 0 && active == 0 || !q.running {
			errc <- nil
			break
		}
		sent := requestc
		var fut *Future
		if size > 0 {
			fut = q.loadFuture(c.Clone())
			if fut == nil {
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

func independentRunner(requestc <-chan *Future, complete chan<- struct{}) {
	for req := range requestc {
		req.Do()
		complete <- struct{}{}
	}
}

func (q *Queue) loadFuture(c *crawler.Collector) *Future {
	if q.storage.Len() == 0 {
		return nil
	}
	q.q.Lock()
	fraw := q.storage.PopFront()
	q.q.Unlock()
	fut := fraw.(*Future)
	if fut != nil {
		fut.setCollector(c)
	}
	return fut
}

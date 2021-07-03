package queue

import (
	"net/url"
	"sync"

	"github.com/flywave/go-tileproxy/crawler"
)

const stop = true

type Storage interface {
	Init() error
	AddRequest([]byte) error
	GetRequest() ([]byte, error)
	QueueSize() (int, error)
}

type Queue struct {
	Threads int
	storage Storage
	wake    chan struct{}
	mut     sync.Mutex
	running bool
}

type InMemoryQueueStorage struct {
	MaxSize int
	lock    *sync.RWMutex
	size    int
	first   *inMemoryQueueItem
	last    *inMemoryQueueItem
}

type inMemoryQueueItem struct {
	Request []byte
	Next    *inMemoryQueueItem
}

func New(threads int, s Storage) (*Queue, error) {
	if s == nil {
		s = &InMemoryQueueStorage{MaxSize: 100000}
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return &Queue{
		Threads: threads,
		storage: s,
		running: true,
	}, nil
}

func (q *Queue) IsEmpty() bool {
	s, _ := q.Size()
	return s == 0
}

func (q *Queue) AddURL(URL string) error {
	u, err := url.Parse(URL)
	if err != nil {
		return err
	}
	r := &crawler.Request{
		URL:    u,
		Method: "GET",
	}
	d, err := r.Marshal()
	if err != nil {
		return err
	}
	return q.storage.AddRequest(d)
}

func (q *Queue) AddRequest(r *crawler.Request) error {
	q.mut.Lock()
	waken := q.wake != nil
	q.mut.Unlock()
	if !waken {
		return q.storeRequest(r)
	}
	err := q.storeRequest(r)
	if err != nil {
		return err
	}
	q.wake <- struct{}{}
	return nil
}

func (q *Queue) storeRequest(r *crawler.Request) error {
	d, err := r.Marshal()
	if err != nil {
		return err
	}
	return q.storage.AddRequest(d)
}

func (q *Queue) Size() (int, error) {
	return q.storage.QueueSize()
}

func (q *Queue) Run(c *crawler.Collector) error {
	q.mut.Lock()
	if q.wake != nil && q.running == true {
		q.mut.Unlock()
		panic("cannot call duplicate Queue.Run")
	}
	q.wake = make(chan struct{})
	q.running = true
	q.mut.Unlock()

	requestc := make(chan *crawler.Request)
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

func (q *Queue) loop(c *crawler.Collector, requestc chan<- *crawler.Request, complete <-chan struct{}, errc chan<- error) {
	var active int
	for {
		size, err := q.storage.QueueSize()
		if err != nil {
			errc <- err
			break
		}
		if size == 0 && active == 0 || !q.running {
			errc <- nil
			break
		}
		sent := requestc
		var req *crawler.Request
		if size > 0 {
			req, err = q.loadRequest(c)
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

func independentRunner(requestc <-chan *crawler.Request, complete chan<- struct{}) {
	for req := range requestc {
		req.Do()
		complete <- struct{}{}
	}
}

func (q *Queue) loadRequest(c *crawler.Collector) (*crawler.Request, error) {
	buf, err := q.storage.GetRequest()
	if err != nil {
		return nil, err
	}
	copied := make([]byte, len(buf))
	copy(copied, buf)
	return c.UnmarshalRequest(copied)
}

func (q *InMemoryQueueStorage) Init() error {
	q.lock = &sync.RWMutex{}
	return nil
}

func (q *InMemoryQueueStorage) AddRequest(r []byte) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.MaxSize > 0 && q.size >= q.MaxSize {
		return crawler.ErrQueueFull
	}
	i := &inMemoryQueueItem{Request: r}
	if q.first == nil {
		q.first = i
	} else {
		q.last.Next = i
	}
	q.last = i
	q.size++
	return nil
}

func (q *InMemoryQueueStorage) GetRequest() ([]byte, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.size == 0 {
		return nil, nil
	}
	r := q.first.Request
	q.first = q.first.Next
	q.size--
	return r, nil
}

func (q *InMemoryQueueStorage) QueueSize() (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.size, nil
}

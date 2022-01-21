package client

import (
	"errors"
	"sync"
)

type Storage interface {
	Init() error
	AddFuture(*Future) error
	GetFuture() (*Future, error)
	QueueSize() (int, error)
}

type InMemoryQueueStorage struct {
	MaxSize int
	lock    *sync.RWMutex
	size    int
	first   *inMemoryQueueItem
	last    *inMemoryQueueItem
}

type inMemoryQueueItem struct {
	Future *Future
	Next   *inMemoryQueueItem
}

func (q *InMemoryQueueStorage) Init() error {
	q.lock = &sync.RWMutex{}
	return nil
}

func (q *InMemoryQueueStorage) AddFuture(f *Future) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.MaxSize > 0 && q.size >= q.MaxSize {
		return errors.New("ErrQueueFull")
	}
	i := &inMemoryQueueItem{Future: f}
	if q.first == nil {
		q.first = i
	} else {
		q.last.Next = i
	}
	q.last = i
	q.size++
	return nil
}

func (q *InMemoryQueueStorage) GetFuture() (*Future, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.size == 0 {
		return nil, nil
	}
	r := q.first.Future
	q.first = q.first.Next
	q.size--
	return r, nil
}

func (q *InMemoryQueueStorage) QueueSize() (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.size, nil
}

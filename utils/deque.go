package utils

const minCapacity = 16

type Deque struct {
	buf    []interface{}
	head   int
	tail   int
	count  int
	minCap int
}

func NewDeque(size ...int) *Deque {
	var capacity, minimum int
	if len(size) >= 1 {
		capacity = size[0]
		if len(size) >= 2 {
			minimum = size[1]
		}
	}

	minCap := minCapacity
	for minCap < minimum {
		minCap <<= 1
	}

	var buf []interface{}
	if capacity != 0 {
		bufSize := minCap
		for bufSize < capacity {
			bufSize <<= 1
		}
		buf = make([]interface{}, bufSize)
	}

	return &Deque{
		buf:    buf,
		minCap: minCap,
	}
}

func (q *Deque) Cap() int {
	return len(q.buf)
}

func (q *Deque) Len() int {
	return q.count
}

func (q *Deque) PushBack(elem interface{}) {
	q.growIfFull()

	q.buf[q.tail] = elem
	q.tail = q.next(q.tail)
	q.count++
}

func (q *Deque) PushFront(elem interface{}) {
	q.growIfFull()

	q.head = q.prev(q.head)
	q.buf[q.head] = elem
	q.count++
}

func (q *Deque) PopFront() interface{} {
	if q.count <= 0 {
		panic("deque: PopFront() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = q.next(q.head)
	q.count--

	q.shrinkIfExcess()
	return ret
}

func (q *Deque) PopBack() interface{} {
	if q.count <= 0 {
		panic("deque: PopBack() called on empty queue")
	}

	q.tail = q.prev(q.tail)

	ret := q.buf[q.tail]
	q.buf[q.tail] = nil
	q.count--

	q.shrinkIfExcess()
	return ret
}

func (q *Deque) Front() interface{} {
	if q.count <= 0 {
		panic("deque: Front() called when empty")
	}
	return q.buf[q.head]
}

func (q *Deque) Back() interface{} {
	if q.count <= 0 {
		panic("deque: Back() called when empty")
	}
	return q.buf[q.prev(q.tail)]
}

func (q *Deque) At(i int) interface{} {
	if i < 0 || i >= q.count {
		panic("deque: At() called with index out of range")
	}
	return q.buf[(q.head+i)&(len(q.buf)-1)]
}

func (q *Deque) Set(i int, elem interface{}) {
	if i < 0 || i >= q.count {
		panic("deque: Set() called with index out of range")
	}
	q.buf[(q.head+i)&(len(q.buf)-1)] = elem
}

func (q *Deque) Clear() {
	modBits := len(q.buf) - 1
	for h := q.head; h != q.tail; h = (h + 1) & modBits {
		q.buf[h] = nil
	}
	q.head = 0
	q.tail = 0
	q.count = 0
}

func (q *Deque) Rotate(n int) {
	if q.count <= 1 {
		return
	}
	n %= q.count
	if n == 0 {
		return
	}

	modBits := len(q.buf) - 1
	if q.head == q.tail {
		q.head = (q.head + n) & modBits
		q.tail = (q.tail + n) & modBits
		return
	}

	if n < 0 {
		for ; n < 0; n++ {
			q.head = (q.head - 1) & modBits
			q.tail = (q.tail - 1) & modBits
			q.buf[q.head] = q.buf[q.tail]
			q.buf[q.tail] = nil
		}
		return
	}

	for ; n > 0; n-- {
		q.buf[q.tail] = q.buf[q.head]
		q.buf[q.head] = nil
		q.head = (q.head + 1) & modBits
		q.tail = (q.tail + 1) & modBits
	}
}

func (q *Deque) SetMinCapacity(minCapacityExp uint) {
	if 1<<minCapacityExp > minCapacity {
		q.minCap = 1 << minCapacityExp
	} else {
		q.minCap = minCapacity
	}
}

func (q *Deque) prev(i int) int {
	return (i - 1) & (len(q.buf) - 1)
}

func (q *Deque) next(i int) int {
	return (i + 1) & (len(q.buf) - 1)
}

func (q *Deque) growIfFull() {
	if q.count != len(q.buf) {
		return
	}
	if len(q.buf) == 0 {
		if q.minCap == 0 {
			q.minCap = minCapacity
		}
		q.buf = make([]interface{}, q.minCap)
		return
	}
	q.resize()
}

func (q *Deque) shrinkIfExcess() {
	if len(q.buf) > q.minCap && (q.count<<2) == len(q.buf) {
		q.resize()
	}
}

func (q *Deque) resize() {
	newBuf := make([]interface{}, q.count<<1)
	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

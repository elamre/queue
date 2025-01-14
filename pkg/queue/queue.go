package queue

import (
	"math/rand"
	"sync"
)

const minQueueLen = 32

type Queue[T comparable] struct {
	items             map[int64]T
	ids               map[T]int64
	buf               []int64
	head, tail, count int
	mutex             *sync.Mutex
	notEmpty          *sync.Cond
	// You can subscribe to this channel to know whether queue is not empty
	NotEmpty chan struct{}
}

func New[T comparable]() *Queue[T] {
	q := &Queue[T]{
		items:    make(map[int64]T),
		ids:      make(map[T]int64),
		buf:      make([]int64, minQueueLen),
		mutex:    &sync.Mutex{},
		NotEmpty: make(chan struct{}, 1),
	}

	q.notEmpty = sync.NewCond(q.mutex)

	return q
}

// Removes all elements from queue
func (q *Queue[T]) Clean() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items = make(map[int64]T)
	q.ids = make(map[T]int64)
	q.buf = make([]int64, minQueueLen)
	q.tail = 0
	q.head = 0
	q.count = 0
}

// Returns the number of elements in queue
func (q *Queue[T]) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.items)
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[T]) resize() {
	newCount := q.count << 1

	if q.count < 2<<18 {
		newCount = newCount << 2
	}

	newBuf := make([]int64, newCount)

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

func (q *Queue[T]) notify() {
	if len(q.items) > 0 {
		select {
		case q.NotEmpty <- struct{}{}:
		default:
		}
	}
}

// Adds one element at the back of the queue
func (q *Queue[T]) Append(elem T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.count == len(q.buf) {
		q.resize()
	}

	id := q.newId()
	q.items[id] = elem
	q.ids[elem] = id
	q.buf[q.tail] = id
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++

	q.notify()

	if q.count == 1 {
		q.notEmpty.Broadcast()
	}
}

func (q *Queue[T]) newId() int64 {
	for {
		id := rand.Int63()
		_, ok := q.items[id]
		if id != 0 && !ok {
			return id
		}
	}
}

// Adds one element at the front of queue
func (q *Queue[T]) Prepend(elem T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.count == len(q.buf) {
		q.resize()
	}

	q.head = (q.head - 1) & (len(q.buf) - 1)
	id := q.newId()
	q.items[id] = elem
	q.ids[elem] = id
	q.buf[q.head] = id
	// bitwise modulus
	q.count++

	q.notify()

	if q.count == 1 {
		q.notEmpty.Broadcast()
	}
}

// Previews element at the front of queue
func (q *Queue[T]) Front() T {
	var result T
	q.mutex.Lock()
	defer q.mutex.Unlock()

	id := q.buf[q.head]
	if id != 0 {
		result = q.items[id]
	}
	return result
}

// Previews element at the back of queue
func (q *Queue[T]) Back() T {
	var result T
	q.mutex.Lock()
	defer q.mutex.Unlock()
	id := q.buf[(q.tail-1)&(len(q.buf)-1)]
	if id != 0 {
		result = q.items[id]
	}
	return result
}

func (q *Queue[T]) pop() int64 {
	for {
		if q.count <= 0 {
			q.notEmpty.Wait()
		}

		// I have no idea why, but sometimes it's less than 0
		if q.count > 0 {
			break
		}
	}

	id := q.buf[q.head]
	q.buf[q.head] = 0

	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	if len(q.buf) > minQueueLen && (q.count<<1) == len(q.buf) {
		q.resize()
	}

	return id
}

// Pop removes and returns the element from the front of the queue.
// If the queue is empty, it will block
func (q *Queue[T]) Pop() T {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for {
		id := q.pop()

		item, ok := q.items[id]

		if ok {
			delete(q.ids, item)
			delete(q.items, id)
			q.notify()
			return item
		}
	}
}

// Removes one element from the queue
func (q *Queue[T]) Remove(elem T) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	id, ok := q.ids[elem]
	if !ok {
		return false
	}
	delete(q.ids, elem)
	delete(q.items, id)
	return true
}

func (q *Queue[T]) swapElem(idx1, idx2 int64) {
	t := q.buf[idx1]
	q.buf[idx1] = q.buf[idx2]
	q.buf[idx2] = t
}

func (q *Queue[T]) partition(s func(elem1 T, elem2 T) int, low, high int64) int64 {
	pivot := q.items[q.buf[high]]
	i := low - 1
	for j := low; j < high; j++ {
		id := q.buf[j]
		elem := q.items[id]
		if s(elem, pivot) <= 0 {
			i++
			q.swapElem(i, j)
		}
	}
	q.swapElem(i+1, high)
	return i + 1
}

// Sorts the queue
func (q *Queue[T]) quickSort(s func(elem1 T, elem2 T) int, low, high int64) {
	if low < high {
		pi := q.partition(s, low, high)
		q.quickSort(s, low, pi-1)
		q.quickSort(s, pi+1, high)
	}
}

func (q *Queue[T]) QuickSort(s func(elem1 T, elem2 T) int) {
	q.quickSort(s, 0, int64(q.Length())-1)
}

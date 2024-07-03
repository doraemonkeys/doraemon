package doraemon

import "sync"

// SimpleMQ is a simple message queue with a minimum buffer capacity.
type SimpleMQ[T any] struct {
	bufMinCap int
	buffer    []T
	// popAll之前必须要获取到这个信号
	popallSemaChan chan struct{}
	bufferLock     *sync.Mutex
}

// NewSimpleMQ creates a new SimpleMQ with the specified minimum buffer capacity.
func NewSimpleMQ[T any](bufMinCap int) *SimpleMQ[T] {
	return &SimpleMQ[T]{
		bufMinCap:      bufMinCap,
		buffer:         make([]T, 0, bufMinCap),
		popallSemaChan: make(chan struct{}, 1),
		bufferLock:     &sync.Mutex{},
	}
}

// Push adds one or more elements to the queue.
func (b *SimpleMQ[T]) Push(v ...T) {
	b.PushSlice(v)
}

// PushSlice adds a slice of elements to the queue.
func (b *SimpleMQ[T]) PushSlice(values []T) {
	b.bufferLock.Lock()
	b.buffer = append(b.buffer, values...)
	if len(b.buffer) == len(values) {
		if len(b.popallSemaChan) != 0 {
			panic("unexpected situation: len(b.semaChan) != 0")
		}
		b.popallSemaChan <- struct{}{}
	}
	b.bufferLock.Unlock()
}

// popAll removes and returns all elements from the queue.
func (b *SimpleMQ[T]) popAll() []T {
	var newBuffer = make([]T, 0, b.bufMinCap)
	var ret []T
	b.bufferLock.Lock()
	ret = b.buffer
	b.buffer = newBuffer
	b.bufferLock.Unlock()
	return ret
}

// WaitPopAll waits for elements to be available and then removes and returns all elements from the queue.
func (b *SimpleMQ[T]) WaitPopAll() []T {
	<-b.popallSemaChan
	return b.popAll()
}

// TryPopAll tries to remove and return all elements from the queue without blocking. Returns the elements and a boolean indicating success.
func (b *SimpleMQ[T]) TryPopAll() ([]T, bool) {
	select {
	case <-b.popallSemaChan:
		return b.popAll(), true
	default:
		return nil, false
	}
}

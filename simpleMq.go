package doraemon

import "sync"

// SimpleMQ is a simple message queue with a minimum buffer capacity.
type SimpleMQ[T any] struct {
	bufMinCap  int
	buffer     *[]T
	bufferPool *sync.Pool
	// popAll之前必须要获取到这个信号
	popallSemaChan chan struct{}
	bufferLock     *sync.Mutex
}

// NewSimpleMQ creates a new SimpleMQ with the specified minimum buffer capacity.
func NewSimpleMQ[T any](bufMinCap int) *SimpleMQ[T] {
	var buffer = make([]T, 0, bufMinCap)
	return &SimpleMQ[T]{
		bufMinCap:      bufMinCap,
		buffer:         &buffer,
		bufferLock:     &sync.Mutex{},
		popallSemaChan: make(chan struct{}, 1),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				var buffer = make([]T, 0, bufMinCap)
				return &buffer
			},
		},
	}
}

// Push adds one or more elements to the queue.
func (b *SimpleMQ[T]) Push(v ...T) {
	b.PushSlice(v)
}

// PushSlice adds a slice of elements to the queue.
func (b *SimpleMQ[T]) PushSlice(values []T) {
	if len(values) == 0 {
		return
	}
	b.bufferLock.Lock()
	*b.buffer = append(*b.buffer, values...)
	if len(*b.buffer) == len(values) {
		if len(b.popallSemaChan) != 0 {
			panic("unexpected situation: semaChan != 0")
		}
		b.popallSemaChan <- struct{}{}
	}
	b.bufferLock.Unlock()
}

// popAll removes and returns all elements from the queue.
// The caller must get the semaphore before calling this function.
func (b *SimpleMQ[T]) popAll() *[]T {
	var newBuffer = b.bufferPool.Get().(*[]T)
	var ret *[]T
	b.bufferLock.Lock()
	ret = b.buffer
	b.buffer = newBuffer
	b.bufferLock.Unlock()
	return ret
}

func (b *SimpleMQ[T]) clearSeamChan() {
	select {
	case <-b.popallSemaChan:
	default:
	}
}

func (b *SimpleMQ[T]) enableSeamChan() {
	select {
	case b.popallSemaChan <- struct{}{}:
	default:
	}
}

// SwapBuffer swaps the current buffer with a new buffer and returns the old buffer.
func (b *SimpleMQ[T]) SwapBuffer(newBuffer *[]T) *[]T {
	var ret *[]T
	b.bufferLock.Lock()
	ret = b.buffer
	b.buffer = newBuffer
	if len(*newBuffer) == 0 {
		b.clearSeamChan()
	} else {
		b.enableSeamChan()
	}
	b.bufferLock.Unlock()
	return ret
}

// RecycleBuffer returns the given buffer to the buffer pool for reuse.
// This helps to reduce memory allocations by reusing previously allocated buffers.
func (b *SimpleMQ[T]) RecycleBuffer(buffer *[]T) {
	b.bufferPool.Put(buffer)
}

// WaitPopAll waits for elements to be available and then removes and returns all elements from the queue.
func (b *SimpleMQ[T]) WaitPopAll() *[]T {
	<-b.popallSemaChan
	return b.popAll()
}

// TryPopAll tries to remove and return all elements from the queue without blocking. Returns the elements and a boolean indicating success.
func (b *SimpleMQ[T]) TryPopAll() (*[]T, bool) {
	select {
	case <-b.popallSemaChan:
		return b.popAll(), true
	default:
		return nil, false
	}
}

func (b *SimpleMQ[T]) Len() int {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	return len(*b.buffer)
}

func (b *SimpleMQ[T]) Cap() int {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	return cap(*b.buffer)
}

func (b *SimpleMQ[T]) IsEmpty() bool {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	return len(*b.buffer) == 0
}

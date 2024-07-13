package doraemon

import (
	"context"
	"sync"
)

// SimpleMQ is a simple message queue with a minimum buffer capacity.
type SimpleMQ[T any] struct {
	bufMinCap  int
	buffer     *[]T
	bufferPool *sync.Pool
	// chan waiting for notification before emptying the buffer
	popallCondChan chan struct{}
	bufferLock     *sync.Mutex
}

// NewSimpleMQ creates a new SimpleMQ with the specified minimum buffer capacity.
func NewSimpleMQ[T any](bufMinCap int) *SimpleMQ[T] {
	var buffer = make([]T, 0, bufMinCap)
	return &SimpleMQ[T]{
		bufMinCap:      bufMinCap,
		buffer:         &buffer,
		bufferLock:     &sync.Mutex{},
		popallCondChan: make(chan struct{}, 1),
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
	// notify the waiting goroutine that the buffer is not empty
	if len(*b.buffer) == len(values) {
		b.enableSignal()
	}
	b.bufferLock.Unlock()
}

// popAll removes and returns all elements from the queue.
// If the queue is empty, it returns nil, this happens only with a low probability when SwapBuffer is called.
//
// The caller should wait the signal before calling this function.
func (b *SimpleMQ[T]) popAll() *[]T {
	var newBuffer = b.bufferPool.Get().(*[]T)
	var ret *[]T

	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()

	// low probability
	if len(*b.buffer) == 0 {
		b.bufferPool.Put(newBuffer)
		return nil
	}
	ret = b.buffer
	b.buffer = newBuffer
	return ret
}

func (b *SimpleMQ[T]) enableSignal() {
	select {
	case b.popallCondChan <- struct{}{}:
	default:
	}
}

// SwapBuffer swaps the current buffer with a new buffer and returns the old buffer.
//
// Note: The function will directly replace the old buffer with the new buffer without clearing the new buffer's elements.
func (b *SimpleMQ[T]) SwapBuffer(newBuffer *[]T) *[]T {
	var ret *[]T

	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()

	ret = b.buffer
	b.buffer = newBuffer

	if len(*ret) == 0 && len(*newBuffer) != 0 {
		// case 1: empty buffer -> non-empty buffer
		b.enableSignal()
	} else if len(*ret) != 0 && len(*newBuffer) == 0 {
		// case 2: non-empty buffer -> empty buffer
		select {
		case <-b.popallCondChan:
		default:
			// Uh-oh, there's a poor kid who can't read the buffer.
		}
	}

	// case 3: empty buffer -> empty buffer, do nothing
	// case 4: non-empty buffer -> non-empty buffer, do nothing

	return ret
}

// RecycleBuffer returns the given buffer to the buffer pool for reuse.
// This helps to reduce memory allocations by reusing previously allocated buffers.
func (b *SimpleMQ[T]) RecycleBuffer(buffer *[]T) {
	*buffer = (*buffer)[:0]
	b.bufferPool.Put(buffer)
}

// WaitPopAll waits for elements to be available and then removes and returns all elements from the queue.
// After calling this function, you can call RecycleBuffer to recycle the buffer.
func (b *SimpleMQ[T]) WaitPopAll() *[]T {
	for range b.popallCondChan {
		if elements := b.popAll(); elements != nil {
			return elements
		}
	}
	panic("unreachable")
}

// WaitPopAllWithContext like WaitPopAll but with a context.
func (b *SimpleMQ[T]) WaitPopAllWithContext(ctx context.Context) (*[]T, bool) {
	for {
		select {
		case <-b.popallCondChan:
			if elements := b.popAll(); elements != nil {
				return elements, true
			}
		case <-ctx.Done():
			return nil, false
		}
	}
}

// TryPopAll tries to remove and return all elements from the queue without blocking. Returns the elements and a boolean indicating success.
func (b *SimpleMQ[T]) TryPopAll() (*[]T, bool) {
	select {
	case <-b.popallCondChan:
		if elements := b.popAll(); elements != nil {
			return elements, true
		}
		return nil, false
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

func (b *SimpleMQ[T]) LenNoLock() int {
	return len(*b.buffer)
}

func (b *SimpleMQ[T]) CapNoLock() int {
	return cap(*b.buffer)
}

func (b *SimpleMQ[T]) IsEmptyNoLock() bool {
	return len(*b.buffer) == 0
}

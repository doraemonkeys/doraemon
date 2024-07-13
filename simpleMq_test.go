package doraemon

import (
	"sync"
	"testing"
	"time"
)

func TestNewSimpleMQ(t *testing.T) {
	mq := NewSimpleMQ[int](10)
	if mq == nil {
		t.Fatal("Expected non-nil SimpleMQ")
	}
	if mq.bufMinCap != 10 {
		t.Fatalf("Expected bufMinCap to be 10, got %d", mq.bufMinCap)
	}
	if mq.buffer == nil {
		t.Fatal("Expected non-nil buffer")
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0, got %d", len(*mq.buffer))
	}
}

func TestPushAndPopAll(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Test Push
	mq.Push(1, 2, 3)
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Test popAll
	buffer := mq.popAll()
	if len(*buffer) != 3 {
		t.Fatalf("Expected popped buffer length to be 3, got %d", len(*buffer))
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0 after popAll, got %d", len(*mq.buffer))
	}
}

func TestPushSlice(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Test PushSlice
	mq.PushSlice([]int{4, 5, 6})
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Test popAll
	buffer := mq.popAll()
	if len(*buffer) != 3 {
		t.Fatalf("Expected popped buffer length to be 3, got %d", len(*buffer))
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0 after popAll, got %d", len(*mq.buffer))
	}
}

func TestSwapBuffer(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Push some elements
	mq.Push(7, 8, 9)
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Create a new buffer to swap
	newBuffer := make([]int, 0, 10)
	newBuffer = append(newBuffer, 10, 11, 12)

	// Swap buffer
	oldBuffer := mq.SwapBuffer(&newBuffer)
	if len(*oldBuffer) != 3 {
		t.Fatalf("Expected old buffer length to be 3, got %d", len(*oldBuffer))
	}
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected new buffer length to be 3, got %d", len(*mq.buffer))
	}
}

func TestEmptyPopAll(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Test popAll on empty buffer
	buffer := mq.popAll()
	if buffer != nil {
		t.Fatalf("Expected nil buffer, got %v", buffer)
	}
}

func TestEnableSignal(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Test signal enabling
	mq.Push(13, 14, 15)
	select {
	case <-mq.popallCondChan:
		// Expected to receive a signal
	default:
		t.Fatal("Expected to receive a signal, but did not")
	}

	// Test signal disabling
	mq.popAll()
	select {
	case <-mq.popallCondChan:
		t.Fatal("Did not expect to receive a signal, but did")
	default:
		// Expected to not receive a signal
	}
}

func TestSimpleMQ(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	var wg sync.WaitGroup
	numProducers := 5
	numConsumers := 5
	numMessages := 100

	// Producer goroutines
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				mq.Push(producerID*numMessages + j)
			}
		}(i)
	}

	// Consumer goroutines
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			for {
				select {
				case <-mq.popallCondChan:
					buffer := mq.popAll()
					if buffer == nil {
						continue
					}
					for _, value := range *buffer {
						t.Logf("Consumer %d received: %d", consumerID, value)
					}
					mq.bufferPool.Put(buffer)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestSwapBuffer2(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Initial buffer with some elements
	initialBuffer := []int{1, 2, 3, 4, 5}
	mq.PushSlice(initialBuffer)

	// New buffer to swap in
	newBuffer := make([]int, 0, 10)
	swappedBuffer := mq.SwapBuffer(&newBuffer)

	// Check if the swapped buffer is the initial buffer
	if len(*swappedBuffer) != len(initialBuffer) {
		t.Errorf("Expected swapped buffer length to be %d, got %d", len(initialBuffer), len(*swappedBuffer))
	}

	// Check if the new buffer is now the current buffer
	if len(*mq.buffer) != 0 {
		t.Errorf("Expected current buffer length to be 0, got %d", len(*mq.buffer))
	}
}

func TestPushSlice2(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Push some elements
	elements := []int{1, 2, 3, 4, 5}
	mq.PushSlice(elements)

	// Check if the buffer contains the elements
	mq.bufferLock.Lock()
	defer mq.bufferLock.Unlock()
	if len(*mq.buffer) != len(elements) {
		t.Errorf("Expected buffer length to be %d, got %d", len(elements), len(*mq.buffer))
	}

	for i, v := range *mq.buffer {
		if v != elements[i] {
			t.Errorf("Expected buffer element at index %d to be %d, got %d", i, elements[i], v)
		}
	}
}

func TestPush(t *testing.T) {
	mq := NewSimpleMQ[int](10)

	// Push some elements
	mq.Push(1, 2, 3, 4, 5)

	// Check if the buffer contains the elements
	mq.bufferLock.Lock()
	defer mq.bufferLock.Unlock()
	expected := []int{1, 2, 3, 4, 5}
	if len(*mq.buffer) != len(expected) {
		t.Errorf("Expected buffer length to be %d, got %d", len(expected), len(*mq.buffer))
	}

	for i, v := range *mq.buffer {
		if v != expected[i] {
			t.Errorf("Expected buffer element at index %d to be %d, got %d", i, expected[i], v)
		}
	}
}

package timewheel

import (
	"math/rand/v2"
	"sync"
)

type Node[T any] struct {
	next, prev *Node[T]

	// The value stored with this element.
	Value T
}

type ConcurrentList[T any] struct {
	head       *Node[T]
	tail       *Node[T]
	pushBackMu sync.Mutex
}

func NewConcurrentList[T any]() *ConcurrentList[T] {
	return &ConcurrentList[T]{}
}

// RangeInSingleThread is a method that ranges over the elements of the list in a single thread.
// It is not thread-safe and should only be used in a single thread.
func (l *ConcurrentList[T]) RangeInSingleThread(f func(e *Node[T], remove func())) {
	l.pushBackMu.Lock()
	if l.tail == nil {
		l.pushBackMu.Unlock()
		return
	}
	current := l.tail
	for current == l.tail {
		prev := current.prev
		f(current, func() {
			l.removeInSingleThreadRange(current)
		})
		current = prev
		if current == nil {
			l.pushBackMu.Unlock()
			return
		}
	}
	l.pushBackMu.Unlock()

	for current != nil {
		prev := current.prev
		f(current, func() {
			l.removeInSingleThreadRange(current)
		})
		current = prev
	}
}

// PushBack is a method that pushes a value to the back of the list.
// It is thread-safe.
func (l *ConcurrentList[T]) PushBack(v T) *Node[T] {
	newNode := &Node[T]{Value: v}
	l.pushBackMu.Lock()
	defer l.pushBackMu.Unlock()
	if l.tail == nil {
		l.head = newNode
		l.tail = newNode
		return newNode
	}
	newNode.prev = l.tail
	l.tail.next = newNode
	l.tail = newNode
	return newNode
}

func (l *ConcurrentList[T]) removeInSingleThreadRange(e *Node[T]) {
	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		return
	}
	if e == l.head {
		l.head = l.head.next
		l.head.prev = nil
		return
	}
	if e == l.tail {
		l.tail = l.tail.prev
		l.tail.next = nil
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
}

type SharedConcurrentSet[T any] struct {
	sets []*ConcurrentList[T]
}

func NewSharedConcurrentSet[T any](size int) SharedConcurrentSet[T] {
	ret := SharedConcurrentSet[T]{
		sets: make([]*ConcurrentList[T], size),
	}
	for i := range size {
		ret.sets[i] = NewConcurrentList[T]()
	}
	return ret
}

func (s SharedConcurrentSet[T]) PushBack(v T) *Node[T] {
	shardIndex := rand.IntN(len(s.sets))
	return s.sets[shardIndex].PushBack(v)
}

func (s SharedConcurrentSet[T]) RangeInSingleThread(f func(e *Node[T], remove func())) {
	for _, set := range s.sets {
		set.RangeInSingleThread(f)
	}
}

type DoubleLinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
	len  int
}

func NewDoubleLinkedList[T any]() *DoubleLinkedList[T] {
	return &DoubleLinkedList[T]{}
}

func (l *DoubleLinkedList[T]) PushBack(v T) *Node[T] {
	newNode := &Node[T]{Value: v}
	l.len++
	if l.tail == nil {
		l.head = newNode
		l.tail = newNode
		return newNode
	}
	newNode.prev = l.tail
	l.tail.next = newNode
	l.tail = newNode
	return newNode
}

// Remove removes an element from the list, the element must be in the list.
func (l *DoubleLinkedList[T]) Remove(e *Node[T]) T {
	l.len--
	defer func() {
		e.next = nil
		e.prev = nil
	}()
	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		return e.Value
	}
	if e == l.head {
		l.head = l.head.next
		l.head.prev = nil
		return e.Value
	}
	if e == l.tail {
		l.tail = l.tail.prev
		l.tail.next = nil
		return e.Value
	}
	e.prev.next = e.next
	e.next.prev = e.prev
	return e.Value
}

func (l *DoubleLinkedList[T]) Len() int {
	return l.len
}

func (l *DoubleLinkedList[T]) Front() *Node[T] {
	return l.head
}

func (l *DoubleLinkedList[T]) Back() *Node[T] {
	return l.tail
}

func (l *DoubleLinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.len = 0
}

func (l *DoubleLinkedList[T]) Merge(other *DoubleLinkedList[T]) {
	if l == other {
		panic("cannot merge list to itself")
	}
	if other.head == nil {
		return
	}
	if l.head == nil {
		l.head = other.head
		l.tail = other.tail
		l.len = other.len
		other.Clear()
		return
	}
	l.tail.next = other.head
	other.head.prev = l.tail
	l.tail = other.tail
	l.len += other.len
	other.Clear()
}

func (l *DoubleLinkedList[T]) Range(yield func(e *Node[T]) bool) {
	current := l.head
	for current != nil {
		if !yield(current) {
			break
		}
		current = current.next
	}
}

func (l *DoubleLinkedList[T]) RangeBack(yield func(e *Node[T]) bool) {
	current := l.tail
	for current != nil {
		if !yield(current) {
			break
		}
		current = current.prev
	}
}

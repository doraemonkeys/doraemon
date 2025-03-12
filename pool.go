package doraemon

import (
	"context"
	"errors"
	"sync"
)

// Pool contains logic of goroutine reuse.
type Pool struct {
	sema      chan struct{}
	workCH    chan func()
	workLock  sync.RWMutex
	closeOnce sync.Once
	closed    bool
}

// NewPool creates new goroutine pool with given size. It also creates a work
// queue of given size. Finally, it spawns given amount of goroutines immediately.
func NewPool(size, queue, spawn int) *Pool {
	if spawn <= 0 && queue > 0 {
		panic("dead queue configuration detected")
	}
	if spawn > size {
		panic("spawn > workers")
	}
	if size <= 0 || queue < 0 {
		panic("invalid pool configuration")
	}
	p := &Pool{
		sema:   make(chan struct{}, size),
		workCH: make(chan func(), queue),
	}
	for range spawn {
		p.sema <- struct{}{}
		go p.worker(func() {})
	}
	return p
}

// Go schedules task to be executed over pool's workers.
func (p *Pool) Go(task func()) {
	if err := p.schedule(context.Background(), task); err != nil {
		panic(err)
	}
}

func (p *Pool) GoContext(ctx context.Context, task func()) error {
	return p.schedule(ctx, task)
}

func (p *Pool) schedule(ctx context.Context, task func()) error {
	p.workLock.RLock()
	defer p.workLock.RUnlock()

	if p.closed {
		return errors.New("schedule on closed pool")
	}

	select {
	case p.workCH <- task:
		return nil
	default:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.workCH <- task:
		return nil
	case p.sema <- struct{}{}:
		go p.worker(task)
		return nil
	}
}

func (p *Pool) worker(task func()) {
	defer func() { <-p.sema }()

	task()

	for task := range p.workCH {
		task()
	}
}

// Close closes the pool.
//
// After closing, the pool will not accept new tasks and all workers will exit after
// all the tasks in the queue are finished.
func (p *Pool) Close() {
	p.closeOnce.Do(func() {
		p.workLock.Lock()
		close(p.workCH)
		p.closed = true
		p.workLock.Unlock()
	})
}

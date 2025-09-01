package doraemon

import (
	"context"
	"sync"
)

type GoroutinePool interface {
	Go(task func())
	TryShrink()
	Close()
}

var _ GoroutinePool = (*Pool2)(nil)

// Pool contains logic of goroutine reuse.
type Pool struct {
	sema   chan struct{}
	workCH chan func()
}

// NewPool creates new goroutine pool with given size. It also creates a work
// queue of given size. Finally, it spawns given amount of goroutines immediately.
func NewPool(size, queue, spawn int) *Pool {
	if size <= 0 {
		panic("pool size must be positive")
	}
	if queue < 0 {
		panic("queue size cannot be negative")
	}
	if spawn < 0 || spawn > size {
		panic("spawn count must be between 0 and pool size")
	}
	if spawn == 0 && queue > 0 {
		panic("spawn must be positive when queue is non-zero")
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

// Close closes the pool. Only call this method when you are sure that no more tasks will be scheduled.
//
// After closing, the pool will not accept new tasks and all workers will exit after
// all the tasks in the queue are finished.
func (p *Pool) Close() {
	close(p.workCH)
}

// Pool contains logic of goroutine reuse.
type Pool2 struct {
	sema       chan struct{}
	workCH     chan func()
	idleExit   chan struct{}
	idleExitMu sync.RWMutex
}

// NewPool creates new goroutine pool with given size. It also creates a work
// queue of given size. Finally, it spawns given amount of goroutines immediately.
func NewPool2(size, queue, spawn int) *Pool2 {
	if size <= 0 {
		panic("pool size must be positive")
	}
	if queue < 0 {
		panic("queue size cannot be negative")
	}
	if spawn < 0 || spawn > size {
		panic("spawn count must be between 0 and pool size")
	}
	if spawn == 0 && queue > 0 {
		panic("spawn must be positive when queue is non-zero")
	}
	p := &Pool2{
		sema:     make(chan struct{}, size),
		workCH:   make(chan func(), queue),
		idleExit: make(chan struct{}),
	}

	if queue > 0 {
		// At least one goroutine is running
		p.sema <- struct{}{}
		go p.worker0(func() {})
		spawn--
		spawn = max(spawn, 0)
	}

	for range spawn {
		p.sema <- struct{}{}
		go p.worker(func() {})
	}
	return p
}

// Go schedules task to be executed over pool's workers.
func (p *Pool2) Go(task func()) {
	if err := p.schedule(context.Background(), task); err != nil {
		panic(err)
	}
}

func (p *Pool2) GoContext(ctx context.Context, task func()) error {
	return p.schedule(ctx, task)
}

func (p *Pool2) schedule(ctx context.Context, task func()) error {
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

func (p *Pool2) worker0(task func()) {
	defer func() { <-p.sema }()

	task()

	for task := range p.workCH {
		task()
	}
}

func (p *Pool2) worker(task func()) {
	defer func() { <-p.sema }()

	p.idleExitMu.RLock()
	idleExit := p.idleExit
	p.idleExitMu.RUnlock()

	task()

	for {
		// work first
		select {
		case task, ok := <-p.workCH:
			if !ok {
				return
			}
			task()
			continue
		default:
		}

		select {
		case <-idleExit:
			return
		case task, ok := <-p.workCH:
			if !ok {
				return
			}
			task()
		}
	}
}

// TryShrink attempts to terminate idle workers and reduce the number of active
// goroutines.
//
// It signals all currently idle workers to exit. Busy workers will not be
// affected until they finish their current task and become idle. This method
// is useful for releasing resources when the pool's workload has decreased.
func (p *Pool2) TryShrink() {

	p.idleExitMu.Lock()
	oldIdleExit := p.idleExit
	p.idleExit = make(chan struct{})
	p.idleExitMu.Unlock()

	close(oldIdleExit)
}

// Close closes the pool. Only call this method when you are sure that no more tasks will be scheduled.
//
// After closing, the pool will not accept new tasks and all workers will exit after
// all the tasks in the queue are finished.
func (p *Pool2) Close() {
	close(p.workCH)
}

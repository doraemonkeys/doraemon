package doraemon

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"sync/atomic"
)

type WaitGroup struct {
	wg sync.WaitGroup
}

func (w *WaitGroup) Start(fn func()) {
	w.wg.Add(1)
	go func() {
		fn()
		w.wg.Done()
	}()
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// PanicHandlers is the global default panic handler.
// It will print the panic message and stack trace to the console and write it to a file.
var PanicHandlers = []func(any){
	func(recoveredErr any) {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		msg := fmt.Sprintf("%v\n%s\n", recoveredErr, buf[:n])
		_ = os.WriteFile(fmt.Sprintf("panic-%s.log",
			time.Now().Format("2006.01.02_15.04.05")),
			[]byte(msg), 0644)
		_, _ = os.Stderr.WriteString("\x1b[31mpanic\x1b[0m: " + msg + "\n")
	},
}

func handleCrash(additionalHandlers ...func(any)) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
	}
}

func GoWithRecover(fn func(), additionalHandlers ...func(recoveredErr any)) {
	go func() {
		defer handleCrash(additionalHandlers...)
		fn()
	}()
}

type SlidingWindowRateLimiter struct {
	limit          int
	windowSize     time.Duration
	subWindowNum   int
	buckets        []int
	bucketsMu      sync.Mutex
	currentBucket  int
	lastUpdateTime time.Time
}

func NewSlidingWindowRateLimiter(limit int, windowSize time.Duration, subWindowNum int) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		limit:          limit,
		windowSize:     windowSize,
		subWindowNum:   subWindowNum,
		buckets:        make([]int, subWindowNum),
		lastUpdateTime: time.Now(),
	}
}

func (rl *SlidingWindowRateLimiter) Allow() bool {
	now := time.Now()
	timePassed := now.Sub(rl.lastUpdateTime)
	bucketsToAdvance := int(timePassed / (rl.windowSize / time.Duration(rl.subWindowNum)))

	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()

	if bucketsToAdvance > 0 {
		if bucketsToAdvance > rl.subWindowNum {
			// Maximum to clear all sub-windows
			bucketsToAdvance = rl.subWindowNum
		}
		// Move to the next sub-window and clear the current sub-window count
		for range bucketsToAdvance {
			rl.currentBucket = (rl.currentBucket + 1) % rl.subWindowNum
			rl.buckets[rl.currentBucket] = 0
		}
		rl.lastUpdateTime = now
	}

	totalRequests := 0
	for _, count := range rl.buckets {
		totalRequests += count
	}

	if totalRequests >= rl.limit {
		return false
	}

	rl.buckets[rl.currentBucket]++
	return true
}

type RateLimiter struct {
	limiters      map[string]*SlidingWindowRateLimiter
	limitersMu    sync.RWMutex
	limit         int
	windowSize    time.Duration
	subWindowNum  int
	cleanupCancel context.CancelFunc
}

// NewRateLimiter creates a new RateLimiter.
//
// Example:
//
//	// 100 requests per minute, divided into 6 sub-windows
//	NewRateLimiter(100, time.Minute, 6)
func NewRateLimiter(limit int, windowSize time.Duration, subWindowNum int) *RateLimiter {
	if limit <= 0 {
		panic("limit must be greater than 0")
	}
	if windowSize <= 0 {
		panic("windowSize must be greater than 0")
	}
	if subWindowNum <= 0 {
		panic("subWindowNum must be greater than 0")
	}
	rl := &RateLimiter{
		limiters:     make(map[string]*SlidingWindowRateLimiter),
		limit:        limit,
		windowSize:   windowSize,
		subWindowNum: subWindowNum,
	}
	ctx, cancel := context.WithCancel(context.Background())
	rl.cleanupCancel = cancel
	go rl.cleanupInactiveLimiters(ctx)
	return rl
}

func (rl *RateLimiter) Allow(userID string) bool {
	rl.limitersMu.RLock()
	if limiter, exists := rl.limiters[userID]; exists {
		rl.limitersMu.RUnlock()
		return limiter.Allow()
	}
	rl.limitersMu.RUnlock()

	limiter := NewSlidingWindowRateLimiter(rl.limit, rl.windowSize, rl.subWindowNum)

	rl.limitersMu.Lock()
	if limiter, exists := rl.limiters[userID]; exists {
		rl.limitersMu.Unlock()
		return limiter.Allow()
	}
	rl.limiters[userID] = limiter
	rl.limitersMu.Unlock()

	return limiter.Allow()
}

func (rl *RateLimiter) cleanupInactiveLimiters(ctx context.Context) {
	const maxIterations = 5000

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(rl.windowSize * 5):
		}
		var zombies []string
		rl.limitersMu.RLock()
		iterations := 0
		for userID, limiter := range rl.limiters {
			if time.Since(limiter.lastUpdateTime) > rl.windowSize {
				zombies = append(zombies, userID)
			}
			iterations++
			if iterations >= maxIterations {
				break
			}
		}
		rl.limitersMu.RUnlock()

		if len(zombies) == 0 {
			continue
		}

		rl.limitersMu.Lock()
		for _, userID := range zombies {
			delete(rl.limiters, userID)
		}
		rl.limitersMu.Unlock()
	}
}

func (rl *RateLimiter) CancelCleanup() {
	rl.cleanupCancel()
}

type Cancel = func() <-chan struct{}

type SyncOptional[T any] struct {
	mu      sync.RWMutex
	hasItem bool
	item    T
}

func (so *SyncOptional[T]) Set(item T) {
	so.mu.Lock()
	so.item = item
	so.hasItem = true
	so.mu.Unlock()
}

func (so *SyncOptional[T]) Swap(item T) (T, bool) {
	so.mu.Lock()
	oldItem := so.item
	so.item = item
	hasItem := so.hasItem
	so.hasItem = true
	so.mu.Unlock()
	return oldItem, hasItem
}

func (so *SyncOptional[T]) Get() (T, bool) {
	so.mu.RLock()
	defer so.mu.RUnlock()
	return so.item, so.hasItem
}

func (so *SyncOptional[T]) Take() T {
	so.mu.Lock()
	defer so.mu.Unlock()
	if !so.hasItem {
		panic("no item")
	}
	so.hasItem = false
	return so.item
}

func (so *SyncOptional[T]) HasItem() bool {
	so.mu.RLock()
	defer so.mu.RUnlock()
	return so.hasItem
}

type CyclicStartGate2 struct {
	ch              chan struct{}
	subscriberCount uint
}

func NewCyclicStartGate2(count uint) CyclicStartGate2 {
	return CyclicStartGate2{
		ch:              make(chan struct{}),
		subscriberCount: count,
	}
}

func (c CyclicStartGate2) ReadyAtGate() {
	c.ch <- struct{}{}
}

func (c CyclicStartGate2) OpenGate() {
	for range c.subscriberCount {
		<-c.ch
	}
}

type CyclicStartGate struct {
	wg              sync.WaitGroup
	waiterGroup     sync.WaitGroup
	fastWaiterCond  sync.Cond
	startCount      atomic.Uint32
	subscriberCount uint32
	wgAddCount      int
}

func NewCyclicStartGate(count uint32) (*CyclicStartGate, []*StartGateRunner) {
	c := &CyclicStartGate{
		subscriberCount: count,
		fastWaiterCond:  sync.Cond{L: &sync.Mutex{}},
	}
	c.resetState()
	waiters := make([]*StartGateRunner, count)
	for i := range waiters {
		waiters[i] = &StartGateRunner{
			currentCount: 0,
			barrier:      c,
		}
	}
	return c, waiters
}

func (c *CyclicStartGate) resetState() {
	// The lock ensures memory visibility
	c.fastWaiterCond.L.Lock()
	c.wgAddCount++
	c.wg.Add(1)
	c.fastWaiterCond.Broadcast()
	c.fastWaiterCond.L.Unlock()
	c.startCount.Store(0)
}

func (c *CyclicStartGate) arriveAtGate(expect int) {
	c.fastWaiterCond.L.Lock()
	for c.wgAddCount != expect {
		c.fastWaiterCond.Wait()
	}
	c.fastWaiterCond.L.Unlock()

	c.wg.Wait()
	if c.startCount.Add(1) == c.subscriberCount {
		// all subscribers were awakened
		c.resetState()
	}
}

func (c *CyclicStartGate) OpenGate() {
	if c.startCount.Load() != 0 {
		panic("should wait all workers before open")
	}
	c.waiterGroup.Add(int(c.subscriberCount))
	c.wg.Done()
}

func (c *CyclicStartGate) WaitAllRunnerFinished() {
	c.waiterGroup.Wait()
}

type StartGateRunner struct {
	currentCount int
	barrier      *CyclicStartGate
}

func (c *StartGateRunner) ReadyAtGate() {
	c.currentCount++
	c.barrier.arriveAtGate(c.currentCount)
}

func (c *StartGateRunner) FinishCycle() {
	c.barrier.waiterGroup.Done()
}

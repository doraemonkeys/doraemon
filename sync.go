package doraemon

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
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

// 全局默认的Panic处理
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
	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()

	now := time.Now()
	timePassed := now.Sub(rl.lastUpdateTime)
	bucketsToAdvance := int(timePassed / (rl.windowSize / time.Duration(rl.subWindowNum)))

	if bucketsToAdvance > 0 {
		for i := 0; i < bucketsToAdvance; i++ {
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

	rl.limitersMu.Lock()
	defer rl.limitersMu.Unlock()
	if limiter, exists := rl.limiters[userID]; exists {
		return limiter.Allow()
	}
	limiter := NewSlidingWindowRateLimiter(rl.limit, rl.windowSize, rl.subWindowNum)
	rl.limiters[userID] = limiter
	return limiter.Allow()
}

func (rl *RateLimiter) cleanupInactiveLimiters(ctx context.Context) {
	ticker := time.NewTicker(rl.windowSize * 5)
	defer ticker.Stop()
	const maxIterations = 5000

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		zombies := []string{}
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

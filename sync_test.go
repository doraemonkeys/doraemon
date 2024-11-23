package doraemon

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 测试SlidingWindowRateLimiter的构造函数
func TestNewSlidingWindowRateLimiter(t *testing.T) {
	limit := 100
	windowSize := time.Minute
	subWindowNum := 6

	rl := NewSlidingWindowRateLimiter(limit, windowSize, subWindowNum)

	if rl.limit != limit {
		t.Errorf("Expected limit %d, got %d", limit, rl.limit)
	}
	if rl.windowSize != windowSize {
		t.Errorf("Expected windowSize %v, got %v", windowSize, rl.windowSize)
	}
	if rl.subWindowNum != subWindowNum {
		t.Errorf("Expected subWindowNum %d, got %d", subWindowNum, rl.subWindowNum)
	}
	if len(rl.buckets) != subWindowNum {
		t.Errorf("Expected %d buckets, got %d", subWindowNum, len(rl.buckets))
	}
}

// 测试SlidingWindowRateLimiter的Allow方法
func TestSlidingWindowRateLimiterAllow(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		t.Parallel()
		rl := NewSlidingWindowRateLimiter(5, time.Second*6, 6)

		// 测试允许的请求
		for i := 0; i < 5; i++ {
			if !rl.Allow() {
				t.Errorf("Expected request %d to be allowed", i+1)
			}
		}

		// 测试超出限制的请求
		if rl.Allow() {
			t.Error("Expected request to be denied")
		}

		time.Sleep(6 * time.Second)
		if !rl.Allow() {
			t.Error("Expected request to be allowed after window slide")
		}
	})
}

// 测试RateLimiter的构造函数
func TestNewRateLimiter(t *testing.T) {
	limit := 100
	windowSize := time.Minute
	subWindowNum := 6

	rl := NewRateLimiter(limit, windowSize, subWindowNum)

	if rl.limit != limit {
		t.Errorf("Expected limit %d, got %d", limit, rl.limit)
	}
	if rl.windowSize != windowSize {
		t.Errorf("Expected windowSize %v, got %v", windowSize, rl.windowSize)
	}
	if rl.subWindowNum != subWindowNum {
		t.Errorf("Expected subWindowNum %d, got %d", subWindowNum, rl.subWindowNum)
	}
}

// 测试RateLimiter的Allow方法
func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute, 6)

	// 测试单个用户
	userID := "user1"
	for i := 0; i < 5; i++ {
		if !rl.Allow(userID) {
			t.Errorf("Expected request %d for user1 to be allowed", i+1)
		}
	}
	if rl.Allow(userID) {
		t.Error("Expected request for user1 to be denied")
	}

	// 测试多个用户
	if !rl.Allow("user2") {
		t.Error("Expected request for user2 to be allowed")
	}
}

// 测试清理不活跃的限速器
func TestCleanupInactiveLimiters(t *testing.T) {
	rl := NewRateLimiter(5, 10*time.Millisecond, 2)

	rl.Allow("user1")
	rl.Allow("user2")

	if len(rl.limiters) != 2 {
		t.Errorf("Expected 2 limiters, got %d", len(rl.limiters))
	}

	// 等待清理周期
	time.Sleep(60 * time.Millisecond)

	rl.limitersMu.RLock()
	count := len(rl.limiters)
	rl.limitersMu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", count)
	}
}

// 测试并发访问
func TestRateLimiterConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(1000, time.Second, 10)
	var wg sync.WaitGroup
	iterations := 1000
	userCount := 10

	for i := 0; i < userCount; i++ {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				rl.Allow(userID)
			}
		}(fmt.Sprintf("user%d", i))
	}

	wg.Wait()

	rl.limitersMu.RLock()
	defer rl.limitersMu.RUnlock()
	if len(rl.limiters) != userCount {
		t.Errorf("Expected %d limiters, got %d", userCount, len(rl.limiters))
	}
}

// 测试取消清理
func TestCancelCleanup(t *testing.T) {
	rl := NewRateLimiter(5, 10*time.Millisecond, 2)

	rl.Allow("user1")

	rl.CancelCleanup()

	// 等待一段时间，确保清理已停止
	time.Sleep(50 * time.Millisecond)

	rl.limitersMu.RLock()
	count := len(rl.limiters)
	rl.limitersMu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 limiter after canceling cleanup, got %d", count)
	}
}

// 测试无效参数
func TestInvalidParameters(t *testing.T) {
	testCases := []struct {
		limit        int
		windowSize   time.Duration
		subWindowNum int
	}{
		{0, time.Second, 1},
		{1, 0, 1},
		{1, time.Second, 0},
	}

	for _, tc := range testCases {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for limit=%d, windowSize=%v, subWindowNum=%d", tc.limit, tc.windowSize, tc.subWindowNum)
				}
			}()
			NewRateLimiter(tc.limit, tc.windowSize, tc.subWindowNum)
		}()
	}
}

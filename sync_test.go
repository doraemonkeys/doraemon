package doraemon

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
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

func TestSlidingWindowRateLimiter(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		limiter := NewSlidingWindowRateLimiter(5, time.Second, 10)
		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d, but it was denied", i+1)
			}
		}
		if limiter.Allow() {
			t.Error("Expected to deny 6th request, but it was allowed")
		}
	})

	t.Run("Full window reset", func(t *testing.T) {
		limiter := NewSlidingWindowRateLimiter(5, 100*time.Millisecond, 10)
		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d, but it was denied", i+1)
			}
		}
		if limiter.Allow() {
			t.Error("Expected to deny 6th request, but it was allowed")
		}

		time.Sleep(110 * time.Millisecond)
		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d after full window reset, but it was denied", i+1)
			}
		}
	})

	t.Run("High concurrency", func(t *testing.T) {
		limiter := NewSlidingWindowRateLimiter(1000, time.Second, 10)
		done := make(chan bool)
		for i := 0; i < 1000; i++ {
			go func() {
				if !limiter.Allow() {
					t.Error("Expected to allow request in high concurrency scenario, but it was denied")
				}
				done <- true
			}()
		}
		for i := 0; i < 1000; i++ {
			<-done
		}
		if limiter.Allow() {
			t.Error("Expected to deny request after limit reached in high concurrency scenario, but it was allowed")
		}
	})

	t.Run("Long inactivity period", func(t *testing.T) {
		limiter := NewSlidingWindowRateLimiter(5, time.Second, 10)
		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d, but it was denied", i+1)
			}
		}
		if limiter.Allow() {
			t.Error("Expected to deny 6th request, but it was allowed")
		}

		time.Sleep(2 * time.Second)
		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d after long inactivity, but it was denied", i+1)
			}
		}
	})

	t.Run("Very small window size", func(t *testing.T) {
		limiter := NewSlidingWindowRateLimiter(3, 10*time.Millisecond, 5)
		for i := 0; i < 3; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected to allow request %d with small window, but it was denied", i+1)
			}
		}
		if limiter.Allow() {
			t.Error("Expected to deny 4th request with small window, but it was allowed")
		}

		time.Sleep(12 * time.Millisecond)
		if !limiter.Allow() {
			t.Error("Expected to allow request after small window passed, but it was denied")
		}
	})
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

	rl.limitersMu.RLock()
	if len(rl.limiters) != 2 {
		rl.limitersMu.RUnlock()
		t.Errorf("Expected 2 limiters, got %d", len(rl.limiters))
		return
	}
	rl.limitersMu.RUnlock()

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

// TestCyclicStartGate_Basic 是核心功能测试，模拟多个 Runner 在多个周期内的工作流程。
func TestCyclicStartGate_Basic(t *testing.T) {
	const (
		numRunners = 10 // 参与同步的协程数量
		numCycles  = 5  // 测试的周期数
	)

	// 创建门和对应的 Runners
	gate, runners := NewCyclicStartGate(uint32(numRunners))
	if len(runners) != numRunners {
		t.Fatalf("NewCyclicStartGate returned %d runners, want %d", len(runners), numRunners)
	}

	// workCounter 用于验证在每个周期中，所有 runner 都确实执行了门后的“工作”
	var workCounter atomic.Uint64

	// 启动所有 runner 协程
	for i := 0; i < numRunners; i++ {
		go func(runner *StartGateRunner) {
			// 每个 runner 将会执行 numCycles 轮
			for j := 0; j < numCycles; j++ {
				// 1. 到达门前，并准备就绪
				runner.ReadyAtGate()

				// 2. 门被打开后，执行工作
				// 为了模拟工作并增加并发调度的机会，这里加入短暂延时
				time.Sleep(time.Millisecond)
				workCounter.Add(1) // 记录工作完成

				// 3. 完成本周期的工作
				runner.FinishCycle()
			}
		}(runners[i])
	}

	// 主协程作为“协调者”，控制每个周期的开始
	for i := 0; i < numCycles; i++ {
		// 短暂等待，以确保所有 runner 协程大概率已经运行到 ReadyAtGate() 并阻塞
		// 在实际逻辑中，这个等待不是必须的，因为 WaitAllRunnerFinished 会确保同步
		time.Sleep(50 * time.Millisecond)

		// 验证：在开门前，工作计数器不应该增加
		expectedBeforeOpen := uint64(i * numRunners)
		if current := workCounter.Load(); current != expectedBeforeOpen {
			t.Errorf("Cycle %d: before OpenGate, workCounter is %d, want %d", i+1, current, expectedBeforeOpen)
		}

		// 4. 打开门，释放所有等待的 runner
		gate.OpenGate()

		// 5. 等待所有 runner 完成本周期的工作
		gate.WaitAllRunnerFinished()

		// 验证：在所有 runner 完成工作后，工作计数器应该等于 (当前周期+1) * runner数量
		expectedAfterCycle := uint64((i + 1) * numRunners)
		if current := workCounter.Load(); current != expectedAfterCycle {
			t.Errorf("Cycle %d: after WaitAllRunnerFinished, workCounter is %d, want %d", i+1, current, expectedAfterCycle)
		}
		t.Logf("Cycle %d finished successfully.", i+1)
	}

	// 最终验证
	finalCount := workCounter.Load()
	if finalCount != numRunners*numCycles {
		t.Fatalf("Final work count is %d, want %d", finalCount, numRunners*numCycles)
	}
}

// TestCyclicStartGate_SingleRunner 测试只有一个 runner 的边界情况。
func TestCyclicStartGate_SingleRunner(t *testing.T) {
	const numCycles = 10
	gate, runners := NewCyclicStartGate(1)
	runner := runners[0]

	var completedCycles atomic.Int32

	go func() {
		for i := 0; i < numCycles; i++ {
			runner.ReadyAtGate()
			// 模拟工作
			completedCycles.Add(1)
			runner.FinishCycle()
		}
	}()

	for i := 0; i < numCycles; i++ {
		// 等待 runner 到达门
		time.Sleep(10 * time.Millisecond)

		// 验证开门前
		if current := completedCycles.Load(); int(current) != i {
			t.Fatalf("Cycle %d: before open, completed cycles is %d, want %d", i+1, current, i)
		}

		gate.OpenGate()
		gate.WaitAllRunnerFinished()

		// 验证开门后
		if current := completedCycles.Load(); int(current) != i+1 {
			t.Fatalf("Cycle %d: after finish, completed cycles is %d, want %d", i+1, current, i+1)
		}
	}

	if finalCount := completedCycles.Load(); int(finalCount) != numCycles {
		t.Errorf("Final completed cycles is %d, want %d", finalCount, numCycles)
	}
}

// TestBasicSingleCycle 验证门的基本功能：
// 1. 所有 goroutine 都在门处等待。
// 2. OpenGate 释放所有 goroutine。
// 3. WaitAllRunnerFinished 等待所有 goroutine 完成其周期。
func TestBasicSingleCycle(t *testing.T) {
	const numRunners = 10

	gate, runners := NewCyclicStartGate(numRunners)
	var workCounter atomic.Int32

	for _, runner := range runners {
		go func(r *StartGateRunner) {
			// 在门前准备就绪
			r.ReadyAtGate()

			// 门打开后执行的工作
			workCounter.Add(1)

			// 通知主 goroutine 该 runner 已完成其周期
			r.FinishCycle()
		}(runner)
	}

	// 给 goroutine 一点时间到达门处
	// 在实际应用中，这通常不是必需的，但在测试中可以使行为更可预测
	time.Sleep(50 * time.Millisecond)

	// 此时，workCounter 应该为 0，因为所有 goroutine 都被阻塞
	if count := workCounter.Load(); count != 0 {
		t.Fatalf("Expected workCounter to be 0 before opening gate, but got %d", count)
	}

	// 打开门
	gate.OpenGate()

	// 等待所有 runner 完成
	gate.WaitAllRunnerFinished()

	// 验证所有 runner 是否都已执行其工作
	if count := workCounter.Load(); count != numRunners {
		t.Errorf("Expected final workCounter to be %d, but got %d", numRunners, count)
	}
}

// TestMultipleCycles 验证门的可重用性。
// 这个测试运行多个周期，并检查每个周期是否都正确完成。
// 这对于验证 resetState 逻辑至关重要。
func TestMultipleCycles(t *testing.T) {
	const numRunners = 20
	const numCycles = 5

	gate, runners := NewCyclicStartGate(numRunners)
	var totalWorkCounter atomic.Int32

	// 启动所有 runner，每个 runner 将循环 numCycles 次
	for _, runner := range runners {
		go func(r *StartGateRunner) {
			for i := 0; i < numCycles; i++ {
				r.ReadyAtGate()
				totalWorkCounter.Add(1)
				r.FinishCycle()
			}
		}(runner)
	}

	// 主 goroutine 驱动周期
	for i := 0; i < numCycles; i++ {
		// 在打开门之前，检查计数器是否与上一个周期结束时的值匹配
		expectedBeforeOpen := int32(i * numRunners)
		if count := totalWorkCounter.Load(); count != expectedBeforeOpen {
			t.Fatalf("Cycle %d: Expected work counter to be %d before opening gate, but got %d", i, expectedBeforeOpen, count)
		}

		gate.OpenGate()
		gate.WaitAllRunnerFinished()

		// 在周期结束后，检查计数器是否已正确递增
		expectedAfterCycle := int32((i + 1) * numRunners)
		if count := totalWorkCounter.Load(); count != expectedAfterCycle {
			t.Fatalf("Cycle %d: Expected work counter to be %d after cycle, but got %d", i, expectedAfterCycle, count)
		}
		t.Logf("Cycle %d completed successfully.", i)
	}

	// 最终验证
	if count := totalWorkCounter.Load(); count != numRunners*numCycles {
		t.Errorf("Expected final workCounter to be %d, but got %d", numRunners*numCycles, count)
	}
}

// TestRaceConditionAndFastRunners 在高并发下进行压力测试。
// 这个测试使用大量的 goroutine 和周期，并且 runner 的工作负载非常小。
// 这会产生“快速 runner”——它们完成一个周期并迅速返回到下一个 ReadyAtGate() 调用。
// 这种情况正是 fastWaiterCond 被设计用来处理的。
// **强烈建议使用 `go test -race` 来运行此测试。**
func TestRaceConditionAndFastRunners(t *testing.T) {
	const numRunners = 100
	const numCycles = 100

	gate, runners := NewCyclicStartGate(numRunners)

	var wg sync.WaitGroup
	wg.Add(numRunners)

	for _, runner := range runners {
		go func(r *StartGateRunner) {
			defer wg.Done()
			for i := 0; i < numCycles; i++ {
				r.ReadyAtGate()
				// 没有额外的工作，以最大化竞争的可能性
				r.FinishCycle()
			}
		}(runner)
	}

	// 驱动周期
	for i := 0; i < numCycles; i++ {
		gate.OpenGate()
		gate.WaitAllRunnerFinished()
	}

	// 等待所有 goroutine 退出
	wg.Wait()
	// 如果测试在没有死锁或竞争条件（用 -race 标志）的情况下完成，则视为通过。
	t.Log("High concurrency test completed without deadlock.")
}

// TestWithSlowAndFastRunners 模拟一个更真实的环境，
// 其中一些 runner 的工作负载比其他 runner 大，导致它们到达门的时间不同。
func TestWithSlowAndFastRunners(t *testing.T) {
	t.Parallel() // 这个测试可以并行运行

	const numRunners = 50
	const numCycles = 10

	gate, runners := NewCyclicStartGate(numRunners)
	var totalWorkCounter atomic.Int32

	var wg sync.WaitGroup
	wg.Add(numRunners)

	// 启动 runner，每个 runner 的工作时间随机
	for i, runner := range runners {
		go func(r *StartGateRunner, id int) {
			defer wg.Done()
			for c := 0; c < numCycles; c++ {
				r.ReadyAtGate()

				// 模拟随机的工作负载
				// 一些 runner 会很快，一些会很慢
				sleepTime := time.Duration(rand.IntN(10)) * time.Millisecond
				time.Sleep(sleepTime)
				totalWorkCounter.Add(1)

				r.FinishCycle()
			}
		}(runner, i)
	}

	// 驱动周期
	for i := 0; i < numCycles; i++ {
		// 等待足够长的时间以确保即使最慢的 runner 也可能到达了门
		// 这有助于验证门是否正确地等待了所有 runner
		time.Sleep(15 * time.Millisecond)

		gate.OpenGate()
		gate.WaitAllRunnerFinished()

		expected := int32((i + 1) * numRunners)
		if count := totalWorkCounter.Load(); count != expected {
			t.Errorf("Cycle %d: Expected count %d, got %d", i, expected, count)
		}
	}

	wg.Wait()
	finalExpected := int32(numRunners * numCycles)
	if count := totalWorkCounter.Load(); count != finalExpected {
		t.Errorf("Final expected count %d, got %d", finalExpected, count)
	}
}

// runBenchmark 是一个辅助函数，它包含了设置、运行和清理的完整逻辑，
// 以便在不同的协程数量下复用。
func runBenchmarkCyclicStartGate(b *testing.B, numRunners uint32) {
	// 1. 设置
	// 创建 CyclicStartGate 和对应的 Runners
	gate, runners := NewCyclicStartGate(numRunners)

	for i := 0; i < int(numRunners); i++ {
		runner := runners[i]
		go func() {

			for {

				// a. 到达闸门并等待
				runner.ReadyAtGate()

				// b. 被释放后，模拟做一些工作。
				// runtime.Gosched() 会让出 CPU，鼓励调度器运行其他协程，
				// 这能更好地模拟真实世界中协程间的交错执行。
				runtime.Gosched()

				// c. 通知主协程，本周期的工作已完成
				runner.FinishCycle()

			}
		}()
	}

	// 3. 运行基准测试核心逻辑
	// 重置计时器，不把上面的设置时间计算在内
	b.ResetTimer()

	// b.N 是由 Go 测试框架决定的迭代次数
	for n := 0; n < b.N; n++ {
		// 这是 "主" 协程在每个周期中做的事情
		gate.OpenGate()              // 打开闸门，释放所有等待的协程
		gate.WaitAllRunnerFinished() // 等待所有协程完成它们本周期的工作
	}

	// 停止计时器，不把下面的清理时间计算在内
	b.StopTimer()

}

// BenchmarkCyclicStartGate 是主基准测试函数
func BenchmarkCyclicStartGate(b *testing.B) {
	// 定义一组协程数量，用于测试不同并发级别下的性能
	runnerCounts := []uint32{1, 8, 64, 256, 1024}

	for _, count := range runnerCounts {
		// 使用 b.Run 创建一个命名的子基准测试。
		// 这使得测试结果更清晰，易于比较。
		b.Run(fmt.Sprintf("%d-Runners", count), func(b *testing.B) {
			runBenchmarkCyclicStartGate(b, count)
		})
	}
}

// runBenchmarkGate 是一个总的基准测试函数
func BenchmarkGates(b *testing.B) {
	// 测试不同数量的并发 goroutine
	counts := []uint32{10, 100, 1000}

	for _, count := range counts {
		// 基准测试：复杂的 WaitGroup + Cond 实现
		b.Run(fmt.Sprintf("Complex-WaitGroup-%d", count), func(b *testing.B) {
			gate, runners := NewCyclicStartGate(count)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for j := 0; j < int(count); j++ {
					go func(runner *StartGateRunner) {
						runner.ReadyAtGate()
						// --- Gate is open, runner is released ---
						runner.FinishCycle()
					}(runners[j])
				}
				gate.OpenGate()
				gate.WaitAllRunnerFinished()
			}
		})

		// 基准测试：简单的 Channel 实现
		b.Run(fmt.Sprintf("Simple-Channel-%d", count), func(b *testing.B) {
			// 由于 Simple-Channel 版本不是真正可循环的，
			// 我们必须在每次迭代中重新创建它。
			// 这也是这个设计的一个重要特性（或缺点），应该被包含在测量中。
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				gate := NewCyclicStartGate2(uint(count))
				// 这个 WaitGroup 是测试必需的，用来等待所有 goroutine 结束，
				// 以确保一次迭代的清理工作在下一次迭代开始前完成。
				var cycleWg sync.WaitGroup
				cycleWg.Add(int(count))

				for j := uint32(0); j < count; j++ {
					go func() {
						defer cycleWg.Done()
						gate.ReadyAtGate()
						// --- Gate is open, runner is released ---
					}()
				}

				gate.OpenGate()
				cycleWg.Wait()
			}
		})
	}
}

// runBenchmarkGate 是一个总的基准测试函数
func BenchmarkGates2(b *testing.B) {
	// 测试不同数量的并发 goroutine
	counts := []uint32{10, 100, 1000}

	for _, count := range counts {
		// 基准测试：复杂的 WaitGroup + Cond 实现
		b.Run(fmt.Sprintf("Complex-WaitGroup-%d", count), func(b *testing.B) {
			gate, runners := NewCyclicStartGate(count)

			b.ReportAllocs()
			b.ResetTimer()

			for j := 0; j < int(count); j++ {
				go func(runner *StartGateRunner) {
					for {
						runner.ReadyAtGate()
						randMs := rand.IntN(10)
						time.Sleep(time.Duration(randMs) * time.Millisecond)
						runner.FinishCycle()
					}
				}(runners[j])
			}

			for i := 0; i < b.N; i++ {
				gate.OpenGate()
				gate.WaitAllRunnerFinished()
			}
		})

		// 基准测试：简单的 Channel 实现
		b.Run(fmt.Sprintf("Simple-Channel-%d", count), func(b *testing.B) {
			// 由于 Simple-Channel 版本不是真正可循环的，
			// 我们必须在每次迭代中重新创建它。
			// 这也是这个设计的一个重要特性（或缺点），应该被包含在测量中。
			b.ReportAllocs()
			// 这个 WaitGroup 是测试必需的，用来等待所有 goroutine 结束，
			// 以确保一次迭代的清理工作在下一次迭代开始前完成。
			var cycleWg sync.WaitGroup

			gate := NewCyclicStartGate2(uint(count))

			for j := uint32(0); j < count; j++ {
				go func() {
					for {
						gate.ReadyAtGate()
						randMs := rand.IntN(10)
						time.Sleep(time.Duration(randMs) * time.Millisecond)
						cycleWg.Done()
					}
				}()
			}

			for i := 0; i < b.N; i++ {
				cycleWg.Add(int(count))
				gate.OpenGate()
				cycleWg.Wait()
			}
		})
	}
}

// runBenchmarkGate 是一个总的基准测试函数
func BenchmarkGates3(b *testing.B) {
	// 测试不同数量的并发 goroutine
	counts := []uint32{10, 100, 1000}

	for _, count := range counts {
		// 基准测试：复杂的 WaitGroup + Cond 实现
		b.Run(fmt.Sprintf("Complex-WaitGroup-%d", count), func(b *testing.B) {
			gate, runners := NewCyclicStartGate(count)

			b.ReportAllocs()
			b.ResetTimer()

			for j := 0; j < int(count); j++ {
				go func(runner *StartGateRunner) {
					for {
						runner.ReadyAtGate()
						runtime.Gosched()
						runner.FinishCycle()
					}
				}(runners[j])
			}

			for i := 0; i < b.N; i++ {
				gate.OpenGate()
				gate.WaitAllRunnerFinished()
			}
		})

		// 基准测试：简单的 Channel 实现
		b.Run(fmt.Sprintf("Simple-Channel-%d", count), func(b *testing.B) {
			// 由于 Simple-Channel 版本不是真正可循环的，
			// 我们必须在每次迭代中重新创建它。
			// 这也是这个设计的一个重要特性（或缺点），应该被包含在测量中。
			b.ReportAllocs()
			// 这个 WaitGroup 是测试必需的，用来等待所有 goroutine 结束，
			// 以确保一次迭代的清理工作在下一次迭代开始前完成。
			var cycleWg sync.WaitGroup

			gate := NewCyclicStartGate2(uint(count))

			for j := uint32(0); j < count; j++ {
				go func() {
					for {
						gate.ReadyAtGate()
						runtime.Gosched()
						cycleWg.Done()
					}
				}()
			}

			for i := 0; i < b.N; i++ {
				cycleWg.Add(int(count))
				gate.OpenGate()
				cycleWg.Wait()
			}
		})
	}
}

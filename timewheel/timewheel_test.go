package timewheel

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doraemonkeys/doraemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// idleTimeout simulates a 10ms idle timeout for a connection.
	// This is short enough to run the benchmark quickly but long enough
	// to be a realistic timeout for high-performance services.
	idleTimeout = 10 * time.Millisecond
)

// BenchmarkSleep simulates managing timeouts by creating a new goroutine for each "connection"
// and using time.Sleep. This is the traditional, straightforward approach.
func BenchmarkSleep(b *testing.B) {
	// A WaitGroup is used to ensure all simulated timeouts have actually fired
	// before the benchmark iteration is considered complete. This is crucial for
	// getting an accurate measurement of the total work done.
	var wg sync.WaitGroup

	b.ResetTimer()
	b.ReportAllocs()

	count := atomic.Uint64{}

	// b.RunParallel runs the provided function in parallel from multiple goroutines.
	// This is the ideal way to simulate many concurrent client connections.
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			count.Add(1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(idleTimeout)
				// In a real app, you would close the connection here.
			}()
		}
	})

	// Wait for all goroutines from the last parallel batch to complete.
	wg.Wait()

	// fmt.Println("goroutine count", count.Load())
}

// BenchmarkTimeWheel uses the timewheel to manage the same timeouts.
func BenchmarkTimeWheel(b *testing.B) {

	tw := newForUnitTesting(time.Millisecond, 200)
	tw.Start()
	defer tw.Stop()

	var wg sync.WaitGroup

	// The TaskFn will simply signal the WaitGroup that the task is complete.
	taskFn := func() {
		wg.Done()
	}

	count := atomic.Uint64{}
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			count.Add(1)
			wg.Add(1)
			tw.AddTask(idleTimeout, taskFn)
		}
	})

	wg.Wait()
	// fmt.Println("goroutine count", count.Load())
}

// noopTaskFn 是一个空操作的任务函数，用于在 benchmark 中避免测量任务执行本身的开销。
// 它立即返回0，表示任务只执行一次。
var noopTaskFn = func() {

}

// BenchmarkTimeWheel_Go 测试单协程下任务投递的性能。
// 这个测试主要衡量 Go() 方法内部的逻辑开销，包括ID生成、时间计算和将任务放入桶中的操作。
func BenchmarkTimeWheel_Go(b *testing.B) {
	// 初始化时间轮
	// tw := New(
	// 	// 使用一个较小的协程池，因为我们主要测试投递性能，而非执行
	// 	WithWorkerPool(nil), // 在这个 benchmark 中我们不关心执行，可以设为 nil 来简化
	// 	WithTimeLevel(
	// 		Level(60, time.Second),
	// 		Level(100, 10*time.Millisecond), // 增加一个毫秒级别，使任务分布更广
	// 	),
	// )

	tw := newForUnitTesting(time.Millisecond*10, 100)
	tw.Start()

	// 测试结束后停止时间轮，防止 goroutine 泄露
	b.Cleanup(func() {
		tw.Stop()
	})

	// 预设一个未来的执行时间，避免任务立即执行
	delay := time.Minute

	// 重置计时器，排除初始化开销
	b.ResetTimer()

	// b.N 是由 testing 框架提供的迭代次数
	for i := 0; i < b.N; i++ {
		// 每次执行时间略有不同，模拟真实场景，避免所有任务落入完全相同的槽位
		tw.AddTask(time.Duration(i)*time.Millisecond+delay, noopTaskFn)
	}
}

// BenchmarkTimeWheel_Go_Parallel 测试多协程并发投递任务的性能。
// 这个测试是评估时间轮扩展性的关键，它会暴露在 Bucket 上的锁竞争问题。
func BenchmarkTimeWheel_Go_Parallel(b *testing.B) {
	// tw := New(
	// 	WithWorkerPool(nil),
	// 	WithTimeLevel(
	// 		Level(60, time.Second),
	// 		Level(100, 10*time.Millisecond),
	// 	),
	// )
	tw := newForUnitTesting(time.Millisecond*10, 100)
	tw.Start()

	b.Cleanup(func() {
		tw.Stop()
	})

	// baseExecTime := time.Now().Add(time.Minute)
	baseDelay := time.Minute
	var counter int64 // 使用原子计数器为不同协程的任务生成不同的执行时间

	b.ResetTimer()

	// RunParallel 会创建 GOMAXPROCS 个 goroutine 并发执行闭包中的代码
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 原子增加计数器，确保每个任务的执行时间都独一无二
			offset := atomic.AddInt64(&counter, 1)
			tw.AddTask(time.Duration(offset)*time.Millisecond+baseDelay, noopTaskFn)
		}
	})
}

// Helper function to check for panics
func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected a panic, but did not get one", name)
		}
	}()
	f()
}

// TestNew tests the constructor of TimeWheel.
func TestNew(t *testing.T) {
	// Test case 1: Valid parameters
	t.Run("ValidParameters", func(t *testing.T) {
		interval := 100 * time.Millisecond
		slotNum := 10
		tw := newForUnitTesting(interval, slotNum)

		if tw == nil {
			t.Fatal("New() returned nil")
		}
		if tw.interval != interval {
			t.Errorf("expected interval to be %v, got %v", interval, tw.interval)
		}
		if len(tw.slots) != slotNum {
			t.Errorf("expected slot number to be %d, got %d", slotNum, len(tw.slots))
		}
		if tw.stopCh == nil {
			t.Error("stopCh should not be nil")
		}
		if tw.workerPool == nil {
			t.Error("workerPool should not be nil")
		}
		for i, slot := range tw.slots {
			if slot == nil {
				t.Errorf("slot %d is nil", i)
			}
		}
	})

	// Test case 2: Invalid interval
	t.Run("InvalidInterval", func(t *testing.T) {
		mustPanic(t, "zero interval", func() { newForUnitTesting(0, 10) })
		mustPanic(t, "negative interval", func() { newForUnitTesting(-1*time.Second, 10) })
	})

	// Test case 3: Invalid slot number
	t.Run("InvalidSlotNum", func(t *testing.T) {
		mustPanic(t, "zero slotNum", func() { newForUnitTesting(1*time.Second, 0) })
		mustPanic(t, "negative slotNum", func() { newForUnitTesting(1*time.Second, -10) })
	})
}

// TestTimeWheel_StartStop tests the lifecycle of the TimeWheel.
func TestTimeWheel_StartStop(t *testing.T) {
	tw := newForUnitTesting(10*time.Millisecond, 10)
	tw.Start()

	// Check if the ticker is running (indirectly)
	if tw.ticker == nil {
		t.Fatal("ticker should not be nil after Start()")
	}

	// Give the run goroutine some time to start
	time.Sleep(20 * time.Millisecond)

	// Stop the timewheel
	tw.Stop()

	// Check if the stop channel is closed
	select {
	case _, ok := <-tw.stopCh:
		if ok {
			t.Error("stopCh should be closed after Stop()")
		}
	default:
		t.Error("stopCh was not closed after Stop()")
	}

	// Wait a bit to ensure the run goroutine has exited
	time.Sleep(20 * time.Millisecond)
	// You might want to use a more sophisticated way to check for goroutine leaks in a real project,
	// but for this test, a simple sleep and observing no panics is sufficient.
}

// TestTimeWheel_AddTask_Simple tests adding and executing a single task with a short delay.
func TestTimeWheel_AddTask_Simple(t *testing.T) {
	interval := 20 * time.Millisecond
	tw := newForUnitTesting(interval, 10)
	tw.Start()
	defer tw.Stop()

	delay := 50 * time.Millisecond // Should execute on the 3rd tick (60ms)
	executed := make(chan time.Time, 1)

	start := time.Now()
	tw.AddTask(delay, func() {
		executed <- time.Now()
	})

	select {
	case execTime := <-executed:
		elapsed := execTime.Sub(start)
		// The task should execute after the specified delay.
		if elapsed < delay {
			t.Errorf("task executed too early: expected after %v, got %v", delay, elapsed)
		}
		// It should not be excessively late. The max delay is delay + interval.
		maxExpectedDelay := delay + interval + (10 * time.Millisecond) // Add a small buffer for scheduling
		if elapsed > maxExpectedDelay {
			t.Errorf("task executed too late: expected before %v, got %v", maxExpectedDelay, elapsed)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("task was not executed within the expected time")
	}
}

// TestTimeWheel_AddTask_LongDelay tests tasks with delays longer than one wheel rotation.
func TestTimeWheel_AddTask_LongDelay(t *testing.T) {
	interval := 10 * time.Millisecond
	slotNum := 10
	wheelDuration := interval * time.Duration(slotNum) // 100ms
	tw := newForUnitTesting(interval, slotNum)
	tw.Start()
	defer tw.Stop()

	// Delay is more than 2 full rotations
	delay := wheelDuration*2 + 55*time.Millisecond // 255ms
	executed := make(chan time.Time, 1)

	// Add a canary task to ensure the wheel is ticking and not executing things early.
	canaryExecuted := make(chan bool, 1)
	tw.AddTask(20*time.Millisecond, func() {
		canaryExecuted <- true
	})

	start := time.Now()
	tw.AddTask(delay, func() {
		executed <- time.Now()
	})

	// Check canary first
	select {
	case <-canaryExecuted:
		// good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("canary task did not execute")
	}

	// Check the long-delay task
	select {
	case execTime := <-executed:
		elapsed := execTime.Sub(start)
		if elapsed < delay {
			t.Errorf("long delay task executed too early: expected after %v, got %v", delay, elapsed)
		}
		maxExpectedDelay := delay + interval + (10 * time.Millisecond) // buffer
		if elapsed > maxExpectedDelay {
			t.Errorf("long delay task executed too late: expected before %v, got %v", maxExpectedDelay, elapsed)
		}
	case <-time.After(400 * time.Millisecond):
		t.Fatal("long delay task was not executed within the expected time")
	}
}

// TestTimeWheel_AddTask_MultipleTasks tests adding and executing multiple tasks concurrently.
func TestTimeWheel_AddTask_MultipleTasks(t *testing.T) {
	interval := 10 * time.Millisecond
	tw := newForUnitTesting(interval, 20)
	tw.Start()
	defer tw.Stop()

	numTasks := 100
	var wg sync.WaitGroup
	wg.Add(numTasks)

	var executionCount int32

	for i := 0; i < numTasks; i++ {
		// Stagger delays to distribute tasks across slots
		delay := time.Duration(10+i%5) * time.Millisecond
		tw.AddTask(delay, func() {
			atomic.AddInt32(&executionCount, 1)
			wg.Done()
		})
	}

	// Wait for all tasks to complete, with a timeout.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All tasks completed successfully.
		if atomic.LoadInt32(&executionCount) != int32(numTasks) {
			t.Errorf("expected %d tasks to be executed, but got %d", numTasks, atomic.LoadInt32(&executionCount))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for all tasks to complete. Executed: %d", atomic.LoadInt32(&executionCount))
	}
}

// TestTimeWheel_AddTask_Concurrent verifies thread-safety of AddTask.
func TestTimeWheel_AddTask_Concurrent(t *testing.T) {
	// Enable race detector for this test
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	interval := 5 * time.Millisecond
	tw := newForUnitTesting(interval, 50)
	tw.Start()
	defer tw.Stop()

	numGoroutines := 50
	tasksPerGoroutine := 20
	totalTasks := numGoroutines * tasksPerGoroutine

	var wg sync.WaitGroup
	wg.Add(totalTasks)
	var executionCount int32

	startWg := sync.WaitGroup{}
	startWg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			startWg.Done()
			startWg.Wait() // Wait for all goroutines to be ready
			for j := 0; j < tasksPerGoroutine; j++ {
				delay := time.Duration(10+j) * time.Millisecond
				tw.AddTask(delay, func() {
					atomic.AddInt32(&executionCount, 1)
					wg.Done()
				})
			}
		}()
	}

	// Wait for all tasks to complete, with a timeout.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		finalCount := atomic.LoadInt32(&executionCount)
		if finalCount != int32(totalTasks) {
			t.Errorf("expected %d tasks to be executed, but got %d", totalTasks, finalCount)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for all tasks to complete. Executed: %d", atomic.LoadInt32(&executionCount))
	}
}

// TestTimeWheel_GetPositionAndCircle tests the internal logic for slot calculation.
func TestTimeWheel_GetPositionAndCircle(t *testing.T) {
	interval := 100 * time.Millisecond
	slotNum := 10
	tw := newForUnitTesting(interval, slotNum)
	// No need to start the wheel for this test

	// Let's manually set currentPos for predictable results
	tw.currentPos = 5

	testCases := []struct {
		name           string
		delay          time.Duration
		expectedPos    int
		expectedCircle int
	}{
		{"ZeroDelay", 0, (5 + 1 + 0) % slotNum, 0},
		{"LessThanInterval", 50 * time.Millisecond, (5 + 1 + 0) % slotNum, 0},       // ticks = 0
		{"ExactlyOneInterval", 100 * time.Millisecond, (5 + 1 + 1) % slotNum, 0},    // ticks = 1
		{"OneAndHalfIntervals", 150 * time.Millisecond, (5 + 1 + 1) % slotNum, 0},   // ticks = 1
		{"AlmostOneRotation", 950 * time.Millisecond, (5 + 1 + 9) % slotNum, 0},     // ticks = 9, pos = (6+9)%10 = 5
		{"ExactlyOneRotation", 1000 * time.Millisecond, (5 + 1 + 10) % slotNum, 1},  // ticks = 10, circle = 1, pos = (6+10)%10 = 6
		{"MoreThanOneRotation", 1250 * time.Millisecond, (5 + 1 + 12) % slotNum, 1}, // ticks = 12, circle = 1, pos = (6+12)%10 = 8
		{"MultipleRotations", 3550 * time.Millisecond, (5 + 1 + 35) % slotNum, 3},   // ticks = 35, circle = 3, pos = (6+35)%10 = 1
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos, circle := tw.getPositionAndCircle(tc.delay)
			if pos != tc.expectedPos {
				t.Errorf("for delay %v, expected pos %d, got %d", tc.delay, tc.expectedPos, pos)
			}
			if circle != tc.expectedCircle {
				t.Errorf("for delay %v, expected circle %d, got %d", tc.delay, tc.expectedCircle, circle)
			}
		})
	}
}

// TestTimeWheel_AddTask_NegativeDelay ensures negative delays are ignored.
func TestTimeWheel_AddTask_NegativeDelay(t *testing.T) {
	interval := 10 * time.Millisecond
	tw := newForUnitTesting(interval, 10)
	tw.Start()
	defer tw.Stop()

	executed := false
	tw.AddTask(-100*time.Millisecond, func() {
		executed = true
	})

	// Wait for a few ticks to ensure it wasn't executed
	time.Sleep(50 * time.Millisecond)

	if executed {
		t.Error("task with negative delay should not be executed")
	}
}

// newForUnitTestingWithMockPool is a helper to inject our mock pool for tests.
func newForUnitTestingWithMockPool(interval time.Duration, slotNum int) (*TimeWheel, *doraemon.GoroutinePool) {
	tw := New(
		WithInterval(interval),
		WithSlotNum(slotNum),
	)
	return tw, nil
}

func TestNewTimeWheel(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		tw := New()
		assert.Equal(t, defaultInterval, tw.interval)
		assert.Equal(t, defaultSlotNum, tw.slotNum)
		assert.NotNil(t, tw.workerPool)
	})

	t.Run("Custom values", func(t *testing.T) {
		interval := 50 * time.Millisecond
		slotNum := 60
		tw := New(WithInterval(interval), WithSlotNum(slotNum))
		assert.Equal(t, interval, tw.interval)
		assert.Equal(t, slotNum, tw.slotNum)
	})

	t.Run("Panic on invalid config", func(t *testing.T) {
		assert.Panics(t, func() { New(WithInterval(0)) })
		assert.Panics(t, func() { New(WithSlotNum(0)) })
		assert.Panics(t, func() { New(WithMaxGoroutineNum(0)) })
	})
}

func TestTimeWheel_BasicTaskExecution(t *testing.T) {
	tw, _ := newForUnitTestingWithMockPool(10*time.Millisecond, 10)
	defer tw.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	executed := new(atomic.Bool)
	job := func() {
		executed.Store(true)
		wg.Done()
	}

	tw.Start()
	// Add a task that should execute after ~3 ticks
	tw.AddTask(35*time.Millisecond, job)

	// Wait for the task to execute, with a timeout
	waitWithTimeout(t, &wg, 100*time.Millisecond)

	assert.True(t, executed.Load(), "Task was not executed")
}

func TestTimeWheel_LongDelayTask_MultipleCircles(t *testing.T) {
	interval := 10 * time.Millisecond
	slotNum := 10 // Total wheel duration = 10 * 10ms = 100ms
	tw, _ := newForUnitTestingWithMockPool(interval, slotNum)
	defer tw.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	var executionTime time.Time
	var mu sync.Mutex
	job := func() {
		mu.Lock()
		executionTime = time.Now()
		mu.Unlock()
		wg.Done()
	}

	tw.Start()
	startTime := time.Now()
	// Delay is 155ms, which requires one full circle (100ms) plus 5.5 more ticks.
	delay := 155 * time.Millisecond
	tw.AddTask(delay, job)

	waitWithTimeout(t, &wg, 300*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.False(t, executionTime.IsZero(), "Task did not execute")

	elapsed := executionTime.Sub(startTime)
	t.Logf("Task executed after %v", elapsed)

	// Check if the execution happened around the expected time.
	// Allow for some scheduling leeway.
	assert.GreaterOrEqual(t, elapsed, delay, "Task executed too early")
	assert.Less(t, elapsed, delay+3*interval, "Task executed too late")
}

func TestTimeWheel_Stop(t *testing.T) {
	tw, _ := newForUnitTestingWithMockPool(10*time.Millisecond, 10)

	executed := new(atomic.Bool)
	job := func() {
		executed.Store(true)
	}

	tw.Start()
	tw.AddTask(50*time.Millisecond, job)

	// Stop the wheel before the task is due
	tw.Stop()

	// Wait for longer than the task's delay to ensure it would have run
	time.Sleep(100 * time.Millisecond)

	assert.False(t, executed.Load(), "Task executed after Stop() was called")
}

func TestTimeWheel_AddTask_NegativeDelay2(t *testing.T) {
	tw, _ := newForUnitTestingWithMockPool(10*time.Millisecond, 10)
	defer tw.Stop()

	executed := new(atomic.Bool)
	job := func() { executed.Store(true) }

	tw.Start()
	// Negative delays should be ignored
	tw.AddTask(-1*time.Second, job)

	time.Sleep(50 * time.Millisecond)
	assert.False(t, executed.Load(), "Task with negative delay was executed")
}

func TestTimeWheel_MultipleTasksInSameSlot(t *testing.T) {
	tw, _ := newForUnitTestingWithMockPool(20*time.Millisecond, 10)
	defer tw.Stop()

	var wg sync.WaitGroup
	wg.Add(3)

	var executionCount atomic.Int32

	job := func() {
		executionCount.Add(1)
		wg.Done()
	}

	tw.Start()
	// These tasks should all land in the same slot (or very close slots) and execute on the same tick
	tw.AddTask(50*time.Millisecond, job)
	tw.AddTask(51*time.Millisecond, job)
	tw.AddTask(52*time.Millisecond, job)

	waitWithTimeout(t, &wg, 200*time.Millisecond)

	assert.Equal(t, int32(3), executionCount.Load(), "Not all tasks were executed")
}

// TestCyclicTask tests a recurring task that re-adds itself.
func TestTimeWheel_CyclicTask(t *testing.T) {
	interval := 20 * time.Millisecond
	tw, _ := newForUnitTestingWithMockPool(interval, 10)
	defer tw.Stop()

	var wg sync.WaitGroup
	wg.Add(1) // We only signal done on the last execution

	var executionCount atomic.Int32
	maxExecutions := int32(3)

	// Define the recurring job using a variable so it can refer to itself.
	var recurringJob func()

	recurringJob = func() {
		count := executionCount.Add(1)
		t.Logf("Task executed, count: %d", count)

		if count < maxExecutions {
			// Reschedule itself for the next interval
			tw.AddTask(interval*2, recurringJob)
		} else {
			// Last execution, signal completion
			wg.Done()
		}
	}

	tw.Start()
	// Add the first task
	tw.AddTask(interval*2, recurringJob)

	// Wait for the 3rd execution to complete
	waitWithTimeout(t, &wg, 500*time.Millisecond)

	assert.Equal(t, maxExecutions, executionCount.Load(), "Cyclic task did not execute the correct number of times")
}

func TestTimeWheel_AddTaskConcurrency(t *testing.T) {
	tw, _ := newForUnitTestingWithMockPool(5*time.Millisecond, 20)
	defer tw.Stop()

	numTasks := 100
	var wg sync.WaitGroup
	wg.Add(numTasks)

	var executionCount atomic.Int32
	job := func() {
		executionCount.Add(1)
		wg.Done()
	}

	tw.Start()

	// Start a bunch of goroutines to add tasks concurrently
	for i := 0; i < numTasks; i++ {
		go func() {
			// Add tasks with a small, slightly varied delay
			tw.AddTask(20*time.Millisecond, job)
		}()
	}

	waitWithTimeout(t, &wg, 500*time.Millisecond)

	assert.Equal(t, int32(numTasks), executionCount.Load(), "Not all concurrently added tasks were executed")
}

// waitWithTimeout is a test helper to wait on a WaitGroup with a timeout.
func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
		// Succeeded
	case <-time.After(timeout):
		t.Fatalf("timed out after %v waiting for WaitGroup", timeout)
	}
}

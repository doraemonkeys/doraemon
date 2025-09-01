package timewheel

import (
	"sync"
	"time"

	"github.com/doraemonkeys/doraemon"
)

// task represents the metadata of a task to be executed.
type task struct {
	// The delay before the task is executed.
	delay time.Duration
	// The number of full wheel rotations before the task is due.
	// This is used for delays longer than one full rotation of the wheel.
	circle int
	// The function to be executed.
	job func()
}

// TimeWheel is a data structure for scheduling tasks with delays.
type TimeWheel struct {
	// The duration between ticks, i.e., how often the wheel's pointer moves forward.
	interval time.Duration
	// The ticker that drives the wheel's rotation.
	ticker *time.Ticker

	// The slots of the wheel, where each slot holds a list of tasks.
	slots   []*ConcurrentList[*task]
	slotNum int // The number of slots in the wheel.

	// The current position of the wheel's pointer.
	currentPos     int
	currentPosLock sync.RWMutex
	workerPool     doraemon.GoroutinePool

	// Configuration for the default worker pool.
	maxGoroutineNum int

	stopCh chan struct{}
}

// Option is a function that configures a TimeWheel.
type Option func(*TimeWheel)

const (
	defaultInterval        = 1 * time.Second
	defaultSlotNum         = 120
	defaultMaxGoroutineNum = 10_000
	defaultQueueSize       = 2048
	defaultPreAllocNum     = 1
)

// WithInterval sets the tick interval for the TimeWheel.
// The default is 1 second.
func WithInterval(interval time.Duration) Option {
	return func(tw *TimeWheel) {
		tw.interval = interval
	}
}

// WithSlotNum sets the number of slots in the TimeWheel.
// The default is 120.
func WithSlotNum(slotNum int) Option {
	return func(tw *TimeWheel) {
		tw.slotNum = slotNum
	}
}

// WithMaxGoroutineNum sets the maximum number of goroutines for the default worker pool.
// This option is ignored if a custom pool is provided via WithWorkerPool.
// The default is 10,000.
func WithMaxGoroutineNum(num int) Option {
	return func(tw *TimeWheel) {
		tw.maxGoroutineNum = num
	}
}

// WithWorkerPool provides a custom worker pool.
// If this option is used, WithMaxGoroutineNum is ignored.
func WithWorkerPool(pool doraemon.GoroutinePool) Option {
	return func(tw *TimeWheel) {
		tw.workerPool = pool
	}
}

// New creates and returns a new TimeWheel with the given options.
func New(opts ...Option) *TimeWheel {
	// Initialize with default values
	tw := &TimeWheel{
		interval:        defaultInterval,
		slotNum:         defaultSlotNum,
		currentPos:      0,
		maxGoroutineNum: defaultMaxGoroutineNum,
		stopCh:          make(chan struct{}),
	}

	// Apply all provided options
	for _, opt := range opts {
		opt(tw)
	}

	// Validate configuration
	if tw.interval <= 0 {
		panic("timewheel: interval must be greater than 0")
	}
	if tw.slotNum <= 0 {
		panic("timewheel: slotNum must be greater than 0")
	}

	// Initialize slots
	tw.slots = make([]*ConcurrentList[*task], tw.slotNum)
	for i := range tw.slotNum {
		tw.slots[i] = NewConcurrentList[*task]()
	}

	// Initialize worker pool if not provided by an option
	if tw.workerPool == nil {
		if tw.maxGoroutineNum <= 0 {
			panic("timewheel: maxGoroutineNum must be greater than 0")
		}
		tw.workerPool = doraemon.NewPool2(tw.maxGoroutineNum, defaultQueueSize, defaultPreAllocNum)
	}

	return tw
}

func newForUnitTesting(interval time.Duration, slotNum int) *TimeWheel {
	return New(WithInterval(interval), WithSlotNum(slotNum))
}

// Start starts the TimeWheel's ticker.
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.run()
}

// Stop stops the TimeWheel.
// It does not wait for any currently running tasks to complete.
// Pending tasks that have not yet been executed will be discarded.
func (tw *TimeWheel) Stop() {
	close(tw.stopCh)
}

// AddTask adds a new task to the TimeWheel.
func (tw *TimeWheel) AddTask(delay time.Duration, job func()) {
	if delay < 0 {
		return
	}
	task := &task{delay: delay, job: job}

	tw.currentPosLock.RLock()
	defer tw.currentPosLock.RUnlock()

	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	tw.slots[pos].PushBack(task)
}

func (tw *TimeWheel) run() {
	for {
		select {
		case <-tw.ticker.C:
			// The tick logic is executed synchronously within this loop, not in a separate goroutine.
			// This is done for two main reasons:
			// 1. To prevent race conditions where a new tick arrives and begins processing
			//    a slot before the previous tick's iteration has completed.
			// 2. To ensure that the worker pool is not used after it has been closed, which
			//    could happen if a tick were running in a separate goroutine when the stop
			//    signal is received.
			tw.tick()
		case <-tw.stopCh:
			tw.ticker.Stop()
			tw.workerPool.Close()
			return
		}
	}
}

// tick moves the wheel's pointer forward by one slot and processes the tasks in that slot.
func (tw *TimeWheel) tick() {

	tw.currentPosLock.Lock()
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
	bucket := tw.slots[tw.currentPos]
	tw.currentPosLock.Unlock()

	// Iterate through all tasks in the current slot.
	for e, remove := range bucket.RangeInSingleThread {
		task := e.Value
		if task.circle > 0 {
			// This task is not yet due, decrement its circle count.
			task.circle--
			continue
		}
		// The task is due, execute it in a worker goroutine.
		tw.workerPool.Go(task.job)
		remove()
	}

	tw.workerPool.TryShrink()
}

// getPositionAndCircle calculates which slot a task should be placed in and how many rotations are required.
func (tw *TimeWheel) getPositionAndCircle(delay time.Duration) (pos int, circle int) {
	ticks := int(delay / tw.interval)

	circle = ticks / tw.slotNum
	// The position is calculated relative to the current pointer's next position (currentPos + 1).
	// This ensures that even a small delay places the task in a future slot.
	pos = (tw.currentPos + 1 + ticks) % tw.slotNum
	return
}

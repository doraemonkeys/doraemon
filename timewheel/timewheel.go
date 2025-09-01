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
	slots []*ConcurrentList[*task]
	// The current position of the wheel's pointer.
	currentPos     int
	currentPosLock sync.RWMutex
	workerPool     *doraemon.Pool2

	stopCh chan struct{}
}

// New creates and returns a new TimeWheel.
func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		panic("invalid interval or slotNum")
	}
	const maxGoroutineNum = 10_000
	tw := &TimeWheel{
		interval:   interval,
		slots:      make([]*ConcurrentList[*task], slotNum),
		currentPos: 0,
		stopCh:     make(chan struct{}),
		workerPool: doraemon.NewPool2(maxGoroutineNum, 2048, 1),
	}
	for i := range slotNum {
		tw.slots[i] = NewConcurrentList[*task]()
	}
	return tw
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
	tw.workerPool.Close()
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
			tw.tick()
		case <-tw.stopCh:
			tw.ticker.Stop()
			return
		}
	}
}

// tick moves the wheel's pointer forward by one slot and processes the tasks in that slot.
func (tw *TimeWheel) tick() {

	tw.currentPosLock.Lock()
	tw.currentPos = (tw.currentPos + 1) % len(tw.slots)
	bucket := tw.slots[tw.currentPos]
	tw.currentPosLock.Unlock()

	// Iterate through all tasks in the current slot.
	bucket.RangeInSingleThread(func(e *Node[*task], remove func()) {
		task := e.Value
		if task.circle > 0 {
			// This task is not yet due, decrement its circle count.
			task.circle--
			return
		}
		// The task is due, execute it in a worker goroutine.
		tw.workerPool.Go(task.job)
		remove()
	})

	tw.workerPool.TryShrink()
}

// getPositionAndCircle calculates which slot a task should be placed in and how many rotations are required.
func (tw *TimeWheel) getPositionAndCircle(delay time.Duration) (pos int, circle int) {
	ticks := int(delay / tw.interval)
	slotNum := len(tw.slots)

	circle = ticks / slotNum
	// The position is calculated relative to the current pointer's next position (currentPos + 1).
	// This ensures that even a small delay places the task in a future slot.
	pos = (tw.currentPos + 1 + ticks) % slotNum
	return
}

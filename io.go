package doraemon

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

func MultiReaderCloser(readers ...io.ReadCloser) io.ReadCloser {
	m := multiReadCloser{}
	rs := make([]io.Reader, len(readers))
	for i, r := range readers {
		rs[i] = r
	}
	m.readers = readers
	m.multiReader = io.MultiReader(rs...)
	return &m
}

type multiReadCloser struct {
	readers     []io.ReadCloser
	multiReader io.Reader
}

func (m *multiReadCloser) Read(p []byte) (n int, err error) {
	return m.multiReader.Read(p)
}

func (m *multiReadCloser) Close() error {
	var errs []error
	for _, r := range m.readers {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type WriteFlushCloser interface {
	io.WriteCloser
	Flush() error
}

// Close automatically Flush and then Close
type BufWriteFlushCloser struct {
	wc   io.WriteCloser
	bufW *bufio.Writer
}

func (b *BufWriteFlushCloser) Write(p []byte) (n int, err error) {
	return b.bufW.Write(p)
}

func (b *BufWriteFlushCloser) Flush() error {
	return b.bufW.Flush()
}

// Close closes the BufWriteFlushCloser, flushing the buffer and closing the underlying writer.
func (b *BufWriteFlushCloser) Close() error {
	err := b.bufW.Flush()
	if err != nil {
		return err
	}
	return b.wc.Close()
}

func NewBufWriteCloser(w io.WriteCloser) *BufWriteFlushCloser {
	return &BufWriteFlushCloser{
		wc:   w,
		bufW: bufio.NewWriter(w),
	}
}

func NewBufWriteCloserSize(w io.WriteCloser, size int) *BufWriteFlushCloser {
	return &BufWriteFlushCloser{
		wc:   w,
		bufW: bufio.NewWriterSize(w, size),
	}
}

type StdBaseLogger interface {
	Errorf(string, ...interface{})
	Errorln(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})
}

type StdLogger interface {
	StdBaseLogger
	Panicf(string, ...interface{})
	Panicln(...interface{})
	Tracef(string, ...interface{})
	Traceln(...interface{})
}

// ReadAllWithLimitBuffer reads from reader until EOF or an error occurs.
// If buf is full before EOF, ReadAllWithLimitBuffer returns an error.
func ReadAllWithLimitBuffer(reader io.Reader, buf []byte) (n int, err error) {
	if buf == nil {
		return 0, errors.New("buffer is nil")
	}
	var nn int
	for {
		nn, err = reader.Read(buf[n:])
		if n == len(buf) && nn == 0 && err != io.EOF {
			return n, errors.New("buffer full")
		}
		n += nn
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
	}
}

// Watchdog monitors a task to ensure it remains "alive".
// If its Pet() method is not called within a configured interval, a timeout is triggered.
type Watchdog struct {
	checkInterval   time.Duration
	onTimeout       func()
	timeoutAutoStop bool

	petCount       atomic.Uint32
	lastCheckCount uint32

	// startOnce ensures the monitoring goroutine is started only once.
	startOnce sync.Once
	// wg waits for the monitoring goroutine to exit completely.
	wg sync.WaitGroup
	// stopCh signals the monitoring goroutine to stop.
	stopCh chan struct{}
}

// Option defines a function that configures a Watchdog instance.
type Option func(*Watchdog)

// WithCheckInterval sets the interval for the watchdog's checks.
func WithCheckInterval(interval time.Duration) Option {
	return func(w *Watchdog) {
		if interval > 0 {
			w.checkInterval = interval
		}
	}
}

// WithOnTimeout sets the callback function to be executed on timeout.
func WithOnTimeout(onTimeout func()) Option {
	return func(w *Watchdog) {
		if onTimeout != nil {
			w.onTimeout = onTimeout
		}
	}
}

// WithAutoStopOnTimeout configures the watchdog to automatically stop
// after the first timeout event.
func WithAutoStopOnTimeout(autoStop bool) Option {
	return func(w *Watchdog) {
		w.timeoutAutoStop = autoStop
	}
}

// NewWatchdog creates a new Watchdog instance configured with the provided options.
func NewWatchdog(opts ...Option) *Watchdog {
	// Default configuration
	w := &Watchdog{
		checkInterval: 10 * time.Second,
		onTimeout: func() {
			fmt.Println("Watchdog timeout: No activity detected.")
		},
		timeoutAutoStop: false,
		stopCh:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Start begins the watchdog's monitoring process.
// It is safe to call Start multiple times.
func (w *Watchdog) Start() {
	w.startOnce.Do(func() {
		w.wg.Add(1)
		go w.monitor()
	})
}

// monitor is the internal loop that performs the checks.
func (w *Watchdog) monitor() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
		}
		if w.petCount.Load() == w.lastCheckCount {
			w.onTimeout()
			if w.timeoutAutoStop {
				// Directly close the stop channel to terminate instead of calling Stop()
				// to avoid a deadlock on the waitgroup.
				close(w.stopCh)
				return
			}
		}
		w.lastCheckCount = w.petCount.Load()
	}
}

// Pet signals to the watchdog that the monitored process is still alive.
func (w *Watchdog) Pet() {
	w.petCount.Add(1)
}

// Stop gracefully terminates the watchdog's monitoring process, blocking until it has shut down.
func (w *Watchdog) Stop() {
	// Use a non-blocking select to close the channel only if it's not already closed.
	select {
	case <-w.stopCh:
		// Already closed.
	default:
		close(w.stopCh)
	}
	// Wait for the monitor goroutine to finish.
	w.wg.Wait()
}

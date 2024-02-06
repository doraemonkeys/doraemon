package doraemon

import (
	"bufio"
	"io"
)

type WriteFlushCloser interface {
	io.WriteCloser
	Flush() error
}

// 实现Close时自动Flush再Close
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
	Tracef(string, ...interface{})
	Traceln(...interface{})
}

type StdLogger interface {
	StdBaseLogger
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

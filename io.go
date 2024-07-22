package doraemon

import (
	"bufio"
	"errors"
	"io"
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

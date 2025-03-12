package doraemon

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

type MonitoredConn struct {
	net.Conn
	bytesRead    uint64
	bytesWritten uint64
	id           string
	startTime    time.Time
	onRead       func(id string, data []byte)
	onWrite      func(id string, data []byte)
}

type MonitorOption func(*MonitoredConn)

func WithID(id string) MonitorOption {
	return func(mc *MonitoredConn) {
		mc.id = id
	}
}

func WithReadCallback(callback func(id string, data []byte)) MonitorOption {
	return func(mc *MonitoredConn) {
		mc.onRead = callback
	}
}

func WithWriteCallback(callback func(id string, data []byte)) MonitorOption {
	return func(mc *MonitoredConn) {
		mc.onWrite = callback
	}
}

func NewMonitoredConn(conn net.Conn, options ...MonitorOption) *MonitoredConn {
	mc := &MonitoredConn{
		Conn: conn,
		id:   fmt.Sprintf("%s-%s", conn.LocalAddr(), conn.RemoteAddr()),
		onRead: func(id string, data []byte) {
			fmt.Printf("[%s] Read %d bytes\n", id, len(data))
		},
		onWrite: func(id string, data []byte) {
			fmt.Printf("[%s] Write %d bytes\n", id, len(data))
		},
	}

	for _, option := range options {
		option(mc)
	}

	return mc
}

func (mc *MonitoredConn) Read(b []byte) (n int, err error) {
	n, err = mc.Conn.Read(b)
	if n > 0 {
		atomic.AddUint64(&mc.bytesRead, uint64(n))
		if mc.onRead != nil {
			mc.onRead(mc.id, b[:n])
		}
	}
	return
}

func (mc *MonitoredConn) Write(b []byte) (n int, err error) {
	n, err = mc.Conn.Write(b)
	if n > 0 {
		atomic.AddUint64(&mc.bytesWritten, uint64(n))
		if mc.onWrite != nil {
			mc.onWrite(mc.id, b[:n])
		}
	}
	return
}

func (mc *MonitoredConn) BytesRead() uint64 {
	return atomic.LoadUint64(&mc.bytesRead)
}

func (mc *MonitoredConn) BytesWritten() uint64 {
	return atomic.LoadUint64(&mc.bytesWritten)
}

func (mc *MonitoredConn) Stats() ConnectionStats {
	return ConnectionStats{
		ID:           mc.id,
		BytesRead:    mc.BytesRead(),
		BytesWritten: mc.BytesWritten(),
		LocalAddr:    mc.LocalAddr().String(),
		RemoteAddr:   mc.RemoteAddr().String(),
		Duration:     time.Since(mc.startTime),
	}
}

type ConnectionStats struct {
	ID           string
	BytesRead    uint64
	BytesWritten uint64
	LocalAddr    string
	RemoteAddr   string
	Duration     time.Duration
}

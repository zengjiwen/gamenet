package conn

import (
	"encoding/binary"
	"errors"
	"gamenet"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	_sendBacklog = 100
)

type TCPConn struct {
	conn     net.Conn
	sendChan chan []byte
	closing  int32
	dieChan  chan struct{}
}

func NewTCPSession(c net.Conn) *TCPConn {
	tc := &TCPConn{
		conn:     c,
		sendChan: make(chan []byte, _sendBacklog),
		dieChan:  make(chan struct{}),
	}
	return tc
}

func readFull(reader io.Reader, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := reader.Read(data)
		if n == left && err == nil {
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil {
			if !isTemporary(err) {
				return err
			} else {
				runtime.Gosched()
			}
		}
	}
	return nil
}

func writeFull(writer io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := writer.Write(data)
		if n == left && err == nil {
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil {
			if !isTemporary(err) {
				return err
			} else {
				runtime.Gosched()
			}
		}
	}
	return nil
}

func isTemporary(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if !errors.As(err, &netErr) {
		return false
	}

	return netErr.Temporary()
}

func (tc *TCPConn) Send(msg []byte) {
	if msg == nil {
		return
	}

	select {
	case tc.sendChan <- msg:
	case <-tc.dieChan:
	}
}

func (tc *TCPConn) WriteLoop(h gamenet.EventHandler) {
	for msg := range tc.sendChan {
		if msg == nil {
			break
		}
		if err := writeFull(tc.conn, msg); err != nil {
			break
		}
	}

	close(tc.dieChan)
	tc.conn.Close()
	h.OnConnClosed(tc)
}

func (tc *TCPConn) ReadLoop(h gamenet.EventHandler) {
	h.OnNewConn(tc)
	for {
		var head [4]byte
		if err := readFull(tc.conn, head[:]); err != nil {
			break
		}

		pLen := binary.LittleEndian.Uint32(head[:])
		p := make([]byte, pLen)
		if err := readFull(tc.conn, p); err != nil {
			break
		}

		h.OnRecv(tc, p)
	}

	tc.sendChan <- nil
}

func (tc *TCPConn) Close() {
	if !atomic.CompareAndSwapInt32(&tc.closing, 0, 1) {
		return
	}

	tcpConn := tc.conn.(*net.TCPConn)
	tcpConn.CloseRead()
	tcpConn.SetReadDeadline(time.Now())
}

package server

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_sendBacklog = 100
	_readTimeout = 5 * time.Minute

	_minDataCap = 128
	_dataShift  = 2
	_maxDataCap = 32 * 1024 * 1024
	_headLen    = 4
)

var (
	_bufrPool   = new(sync.Pool)
	_bufwPool   = new(sync.Pool)
	_packetPool = &sync.Pool{
		New: func() interface{} {
			p := new(packet)
			p.reset()
			return p
		},
	}
	_packetDataCaps  []int
	_packetDataPools = make(map[int]*sync.Pool)
)

func init() {
	dataCap := _minDataCap << _dataShift
	for dataCap < _maxDataCap {
		_packetDataCaps = append(_packetDataCaps, dataCap)
		dataCap <<= _dataShift
	}
	_packetDataCaps = append(_packetDataCaps, _maxDataCap)

	for _, dataCap := range _packetDataCaps {
		_packetDataPools[dataCap] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, dataCap)
			},
		}
	}
}

type tcpConn struct {
	server *tcpServer

	conn net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer

	sendChan chan []byte
	closing  int32
	dieChan  chan struct{}

	userData interface{}
}

func newTCPConn(ts *tcpServer, c net.Conn) *tcpConn {
	tc := &tcpConn{
		server:   ts,
		conn:     c,
		bufr:     newBufr(c),
		bufw:     newBufw(c),
		sendChan: make(chan []byte, _sendBacklog),
		dieChan:  make(chan struct{}),
	}
	return tc
}

func newBufr(r io.Reader) *bufio.Reader {
	bufr, ok := _bufrPool.Get().(*bufio.Reader)
	if !ok {
		return bufio.NewReader(r)
	}

	bufr.Reset(r)
	return bufr
}

func putBufr(bufr *bufio.Reader) {
	bufr.Reset(nil)
	_bufrPool.Put(bufr)
}

func newBufw(w io.Writer) *bufio.Writer {
	bufw, ok := _bufwPool.Get().(*bufio.Writer)
	if !ok {
		return bufio.NewWriter(w)
	}

	bufw.Reset(w)
	return bufw
}

func putBufw(bufw *bufio.Writer) {
	bufw.Reset(nil)
	_bufwPool.Put(bufw)
}

func readFull(r io.Reader, buf []byte) error {
	var n int
	bufLen := len(buf)
	for n < bufLen {
		nn, err := io.ReadFull(r, buf[n:])
		if err != nil {
			if !isTemporary(err) {
				return err
			} else {
				runtime.Gosched()
			}
		}
		n += nn
	}
	return nil
}

func writeFull(w io.Writer, buf []byte) error {
	left := len(buf)
	for left > 0 {
		n, err := w.Write(buf)
		if n == left && err == nil {
			return nil
		}

		if n > 0 {
			buf = buf[n:]
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

func (tc *tcpConn) Send(data []byte) {
	if data == nil {
		return
	}

	select {
	case tc.sendChan <- data:
	case <-tc.dieChan:
	}
}

func (tc *tcpConn) writeLoop() {
	defer func() {
		putBufw(tc.bufw)
		close(tc.dieChan)
		tc.conn.Close()
		if tc.server.opts.eventChan != nil {
			tc.server.opts.eventChan <- func() {
				tc.server.callback.OnConnClosed(tc)
			}
		} else {
			tc.server.callback.OnConnClosed(tc)
		}

		tc.server.tcpConnsMu.Lock()
		delete(tc.server.tcpConns, tc)
		tc.server.tcpConnsMu.Unlock()
	}()

	var headBuf [_headLen]byte
	var packetCount int
	tickTime := time.Now()
	flushTick := time.NewTicker(time.Millisecond)
	var firstRecvTime, lastRecvTime time.Time
	flushDelay := tc.server.opts.flushDelay
	maxFlushDelay := tc.server.opts.maxFlushDelay
	maxFlushPackets := tc.server.opts.maxFlushPackets
	for {
		select {
		case p := <-tc.sendChan:
			if p == nil {
				return
			}

			binary.LittleEndian.PutUint32(headBuf[:], uint32(len(p)))
			if err := writeFull(tc.bufw, headBuf[:]); err != nil {
				return
			}
			if err := writeFull(tc.bufw, p); err != nil {
				return
			}

			lastRecvTime = tickTime
			packetCount++
			if packetCount == 1 {
				firstRecvTime = tickTime
			}
		case tickTime = <-flushTick.C:
			if packetCount == 0 {
				continue
			}

			if packetCount >= maxFlushPackets {
				packetCount = 0
				if err := tc.bufw.Flush(); err != nil {
					return
				}
			} else if tickTime.Sub(lastRecvTime) >= flushDelay*time.Millisecond ||
				tickTime.Sub(firstRecvTime) >= maxFlushDelay*time.Millisecond {
				packetCount = 0
				if err := tc.bufw.Flush(); err != nil {
					return
				}
			}
		}
	}
}

func (tc *tcpConn) readLoop() {
	defer func() {
		putBufr(tc.bufr)
		tc.sendChan <- nil
	}()

	var head [_headLen]byte
	for {
		tc.conn.SetReadDeadline(time.Now().Add(_readTimeout))

		if err := readFull(tc.bufr, head[:]); err != nil {
			return
		}

		pLen := binary.LittleEndian.Uint32(head[:])
		p := newPacket(int(pLen))
		if err := readFull(tc.bufr, p.getData()); err != nil {
			return
		}

		if tc.server.opts.eventChan != nil {
			tc.server.opts.eventChan <- func() {
				tc.server.callback.OnRecvData(tc, p.getData())
				p.release()
			}
		} else {
			tc.server.callback.OnRecvData(tc, p.getData())
			p.release()
		}
	}
}

func (tc *tcpConn) Close() error {
	if !atomic.CompareAndSwapInt32(&tc.closing, 0, 1) {
		return nil
	}

	c := tc.conn.(*net.TCPConn)
	if err := c.CloseRead(); err != nil {
		return err
	}
	return nil
}

func (tc *tcpConn) serve() {
	go tc.writeLoop()
	tc.readLoop()
}

func (tc *tcpConn) SetUserData(userData interface{}) {
	tc.userData = userData
}

func (tc *tcpConn) UserData() interface{} {
	return tc.userData
}

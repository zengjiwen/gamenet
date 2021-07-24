package server

import (
	"github.com/zengjiwen/gamenet"
	"net"
	"sync"
	"time"
)

type tcpServer struct {
	addr string
	ln   net.Listener

	tcpConnsMu sync.Mutex
	tcpConns   map[*tcpConn]struct{}

	callback gamenet.EventCallback
	opts     options
}

func newTCPServer(addr string, callback gamenet.EventCallback, applies ...func(opts *options)) *tcpServer {
	ts := &tcpServer{
		addr:     addr,
		tcpConns: make(map[*tcpConn]struct{}),
		callback: callback,
	}
	for _, apply := range applies {
		apply(&ts.opts)
	}
	return ts
}

func (ts *tcpServer) ListenAndServe() error {
	ln, err := net.Listen("TCP", "127.0.0.1:0")
	if err != nil {
		return err
	}

	return ts.serve(ln)
}

func (ts *tcpServer) serve(ln net.Listener) error {
	ts.ln = ln

	var tempDelay time.Duration
	for {
		c, err := ln.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		tc := newTCPConn(ts, c)
		ts.tcpConnsMu.Lock()
		ts.tcpConns[tc] = struct{}{}
		ts.tcpConnsMu.Unlock()
		if ts.opts.eventChan != nil {
			ts.opts.eventChan <- func() {
				ts.callback.OnNewConn(tc)
			}
		} else {
			ts.callback.OnNewConn(tc)
		}
		go tc.serve()
	}
}

func (ts *tcpServer) Shutdown() error {
	if ts.ln != nil {
		if err := ts.ln.Close(); err != nil {
			return err
		}
	}

	ts.tcpConnsMu.Lock()
	for tc := range ts.tcpConns {
		tc.Close()
	}
	ts.tcpConns = nil
	ts.tcpConnsMu.Unlock()
	return nil
}

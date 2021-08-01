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

	mu       sync.Mutex
	tcpConns map[*tcpConn]struct{}

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
	ln, err := net.Listen("TCP", ts.addr)
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
		ts.mu.Lock()
		ts.tcpConns[tc] = struct{}{}
		ts.mu.Unlock()

		if eventChan := tc.server.opts.eventChan; eventChan != nil {
			eventChan <- func() {
				ts.callback.OnNewConn(tc)
			}
		} else {
			ts.callback.OnNewConn(tc)
		}

		go tc.serve()
	}
}

func (ts *tcpServer) Shutdown() error {
	var err error
	if ts.ln != nil {
		err = ts.ln.Close()
	}

	ts.mu.Lock()
	for tc := range ts.tcpConns {
		if err == nil {
			err = tc.Close()
		} else {
			tc.Close()
		}
	}
	ts.tcpConns = nil
	ts.mu.Unlock()
	return err
}

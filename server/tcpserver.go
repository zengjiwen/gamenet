package server

import (
	"errors"
	"gamenet"
	"gamenet/conn"
	"net"
	"sync"
)

type tcpServer struct {
	addr string

	ln   net.Listener
	lnWG sync.WaitGroup

	tcpConnsMU sync.Mutex
	tcpConns   map[*conn.TCPConn]struct{}

	handler gamenet.EventHandler
}

func NewTCPServer(addr string, handler gamenet.EventHandler) *tcpServer {
	ts := &tcpServer{
		addr: addr,
	}

	ts.handler = handler
	return ts
}

func (ts *tcpServer) Start() {
	ln, err := net.Listen("TCP", "127.0.0.1:0")
	if err != nil {
		return
	}

	ts.lnWG.Add(1)
	go ts.accepting(ln)
}

func (ts *tcpServer) accepting(ln net.Listener) {
	defer ts.lnWG.Done()

	for {
		c, err := ln.Accept()
		if err != nil {
			if isTimeoutError(err) {
				continue
			}
			break
		}

		go ts.serve(c)
	}
}

func (ts *tcpServer) serve(c net.Conn) {
	tc := conn.NewTCPSession(c)
	ts.tcpConnsMU.Lock()
	ts.tcpConns[tc] = struct{}{}
	ts.tcpConnsMU.Unlock()

	go func() {
		defer func() {
			ts.tcpConnsMU.Lock()
			delete(ts.tcpConns, tc)
			ts.tcpConnsMU.Unlock()
		}()
		tc.WriteLoop(ts.handler)
	}()

	tc.ReadLoop(ts.handler)
}

func (ts *tcpServer) Stop() {
	ts.ln.Close()
	ts.lnWG.Wait()

	ts.tcpConnsMU.Lock()
	for tc := range ts.tcpConns {
		tc.Close()
	}
	ts.tcpConns = nil
	ts.tcpConnsMU.Unlock()
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if !errors.As(err, &netErr) {
		return false
	}

	return netErr.Timeout()
}

package server

import (
	"gamenet"
	"net"
	"sync"
	"time"
)

const (
	_broadcastBacklog = 50
)

type tcpServer struct {
	addr string
	ln   net.Listener

	tcpConnsMu sync.Mutex
	tcpConns   map[*tcpConn]struct{}

	groupsMu sync.Mutex
	groups   map[string]map[*tcpConn]struct{}

	handler       gamenet.EventHandler
	broadcastChan chan []byte

	eventChan chan func()
}

func newTCPServer(addr string, handler gamenet.EventHandler, eventChan chan func()) *tcpServer {
	ts := &tcpServer{
		addr:          addr,
		tcpConns:      make(map[*tcpConn]struct{}),
		groups:        make(map[string]map[*tcpConn]struct{}),
		handler:       handler,
		broadcastChan: make(chan []byte, _broadcastBacklog),
		eventChan:     eventChan,
	}
	return ts
}

func (ts *tcpServer) ListenAndServe() error {
	ln, err := net.Listen("TCP", "127.0.0.1:0")
	if err != nil {
		return err
	}

	go ts.broadcastLoop()
	return ts.serve(ln)
}

func (ts *tcpServer) broadcastLoop() {
	for data := range ts.broadcastChan {
		ts.tcpConnsMu.Lock()
		for conn := range ts.tcpConns {
			conn.Send(data)
		}
		ts.tcpConnsMu.Unlock()
	}
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

		tc := newTCPConn(c)
		ts.tcpConnsMu.Lock()
		ts.tcpConns[tc] = struct{}{}
		ts.tcpConnsMu.Unlock()
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
	close(ts.broadcastChan)
	return nil
}

func (ts *tcpServer) Broadcast(data []byte) {
	defer func() {
		recover()
	}()
	ts.broadcastChan <- data
}

func (ts *tcpServer) Multicast(groupName string, data []byte) {
	ts.groupsMu.Lock()
	defer ts.groupsMu.Unlock()

	group, ok := ts.groups[groupName]
	if !ok {
		return
	}

	for conn := range group {
		conn.Send(data)
	}
}

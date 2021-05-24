package gamenet

import (
	"fmt"
	"gamenet/server"
)

type Server interface {
	Start()
	Stop()
}

func NewServer(network, addr string, handler EventHandler) Server {
	if network == "TCP" {
		return server.NewTCPServer(addr, handler)
	}

	panic(fmt.Errorf("wrong network: %s", network))
}

type Conn interface {
	Send(p []byte)
	Close()
}

type EventHandler interface {
	OnNewConn(c Conn)
	OnConnClosed(c Conn)
	OnRecv(c Conn, p []byte)
}

type Message struct {
}

type Packet struct {
}

package gamenet

import "gamenet/server"

type Server interface {
	ListenAndServe() error
	Shutdown() error
	Broadcast(data []byte)
	Multicast(groupName string, data []byte)
}

type Conn interface {
	Send(data []byte)
	Close() error
	AddToGroup(groupName string)
	RemoveFromGroup(groupName string)
}

type EventHandler interface {
	OnNewConn(c Conn)
	OnConnClosed(c Conn)
	OnRecv(c Conn, p *server.Packet)
}

type Message struct {
}

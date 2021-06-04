package gamenet

import (
	"gamenet/server"
)

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
	SetUserData(userData interface{})
	UserData() interface{}
}

type EventHandler interface {
	OnNewConn(c Conn)
	OnConnClosed(c Conn)
	OnRecvPacket(c Conn, p *server.Packet)
}

type Codec interface {
	Encode(data []byte) (*server.Packet, error)
	Decode(p *server.Packet) ([]byte, error)
}

type Marshaler interface {
	Marshal(message interface{}) ([]byte, error)
	Unmarshal(buf []byte, message interface{}) error
}

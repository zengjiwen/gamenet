package gamenet

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
	OnRecvData(c Conn, data []byte)
}

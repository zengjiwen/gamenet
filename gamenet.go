package gamenet

type Server interface {
	ListenAndServe() error
	Shutdown() error
}

type Conn interface {
	Send(data []byte)
	Close() error
	SetUserData(userData interface{})
	UserData() interface{}
}

type EventCallback interface {
	OnNewConn(c Conn)
	OnConnClosed(c Conn)
	OnRecvData(c Conn, data []byte)
}

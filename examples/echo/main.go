package main

import (
	"fmt"
	"gamenet"
	"gamenet/server"
	"os"
	"os/signal"
)

func main() {
	tcpServer := server.NewServer("tcp", "127.0.0.1:0", &echoHandler{}, nil)
	go tcpServer.ListenAndServe()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	tcpServer.Shutdown()
}

type echoHandler struct {
}

func (sc *echoHandler) OnNewConn(c gamenet.Conn) {
	fmt.Println("OnNewConn")
}

func (sc *echoHandler) OnConnClosed(c gamenet.Conn) {
	fmt.Println("OnConnClosed")
}

func (sc *echoHandler) OnRecvPacket(c gamenet.Conn, p *server.Packet) {
	fmt.Printf("OnRecvPacket: %s", p)
	c.Send(p.Data())
}

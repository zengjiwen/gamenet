package main

import (
	"fmt"
	"github.com/zengjiwen/gamenet"
	"github.com/zengjiwen/gamenet/server"
	"os"
	"os/signal"
)

func main() {
	tcpServer := server.NewServer("tcp", "127.0.0.1:0", echoHandler{})
	go tcpServer.ListenAndServe()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	tcpServer.Shutdown()
}

type echoHandler struct {
}

func (eh echoHandler) OnNewConn(c gamenet.Conn) {
	fmt.Println("OnNewConn")
}

func (eh echoHandler) OnConnClosed(c gamenet.Conn) {
	fmt.Println("OnConnClosed")
}

func (eh echoHandler) OnRecvData(c gamenet.Conn, data []byte) {
	fmt.Printf("OnRecvData: %s", data)
	c.Send(data)
}

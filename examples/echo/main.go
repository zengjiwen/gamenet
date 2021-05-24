package main

import (
	"fmt"
	"gamenet"
	"os"
	"os/signal"
)

func main() {
	server := gamenet.NewServer("TCP", "127.0.0.1:0", &echoHandler{})
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	server.Stop()
}

type echoHandler struct {
}

func (sc *echoHandler) OnNewConn(c gamenet.Conn) {
	fmt.Println("OnNewConn")
}

func (sc *echoHandler) OnConnClosed(c gamenet.Conn) {
	fmt.Println("OnConnClosed")
}

func (sc *echoHandler) OnRecv(c gamenet.Conn, p []byte) {
	fmt.Printf("OnRecv: %s", p)
	c.Send(p)
}

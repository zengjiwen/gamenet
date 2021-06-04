package server

import (
	"fmt"
	"gamenet"
	"strings"
)

func NewServer(network, addr string, handler gamenet.EventHandler, eventChan chan func()) gamenet.Server {
	network = strings.ToLower(network)
	switch network {
	case "tcp":
		return newTCPServer(addr, handler, eventChan)
	default:
		panic(fmt.Errorf("wrong network: %s", network))
	}
}

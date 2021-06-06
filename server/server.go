package server

import (
	"fmt"
	"gamenet"
	"strings"
)

func NewServer(network, addr string, handler gamenet.EventHandler, applies ...func(opts *options)) gamenet.Server {
	network = strings.ToLower(network)
	switch network {
	case "tcp":
		return newTCPServer(addr, handler, applies...)
	default:
		panic(fmt.Errorf("wrong network: %s", network))
	}
}

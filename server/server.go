package server

import (
	"fmt"
	"github.com/zengjiwen/gamenet"
	"strings"
)

func NewServer(network, addr string, callback gamenet.EventCallback, applies ...func(opts *options)) gamenet.Server {
	network = strings.ToLower(network)
	switch network {
	case "tcp":
		return newTCPServer(addr, callback, applies...)
	default:
		panic(fmt.Errorf("wrong network: %s", network))
	}
}

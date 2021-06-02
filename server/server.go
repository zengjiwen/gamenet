package server

import (
	"fmt"
	"gamenet"
	"runtime"
	"strings"
)

type Packet struct {
	Type     int
	Len      int
	data     []byte
	initData [_minDataCap]byte
}

func newPacket(len int) *Packet {
	if len > _maxDataCap {
		panic(fmt.Errorf("packet len too large"))
	}

	p := _packetPool.Get().(*Packet)
	if len > _minDataCap {
		for _, dataCap := range _packetDataCaps {
			if dataCap >= len {
				p.data = _packetDataPools[dataCap].Get().([]byte)
				return p
			}
		}
		p.data = _packetDataPools[_maxDataCap].Get().([]byte)
	}
	p.Len = len
	runtime.SetFinalizer(p, (*Packet).Release)
	return p
}

func (p *Packet) Data() []byte {
	return p.data[:p.Len]
}

func (p *Packet) Release() {
	dataCap := len(p.data)
	if dataCap > _minDataCap {
		_packetDataPools[dataCap].Put(p.data)
	}

	p.Reset()
	_packetPool.Put(p)
}

func (p *Packet) Reset() {
	p.Len = 0
	p.data = p.initData[:]
}

func NewServer(network, addr string, handler gamenet.EventHandler) gamenet.Server {
	network = strings.ToLower(network)
	switch network {
	case "tcp":
		return newTCPServer(addr, handler)
	default:
		panic(fmt.Errorf("wrong network: %s", network))
	}
}

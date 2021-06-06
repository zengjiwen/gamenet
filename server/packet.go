package server

import (
	"fmt"
)

type packet struct {
	Len      int
	data     []byte
	initData [_minDataCap]byte
}

func newPacket(len int) *packet {
	if len > _maxDataCap {
		panic(fmt.Errorf("packet len too large"))
	}

	p := _packetPool.Get().(*packet)
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
	return p
}

func (p *packet) getData() []byte {
	return p.data[:p.Len]
}

func (p *packet) release() {
	dataCap := len(p.data)
	if dataCap > _minDataCap {
		_packetDataPools[dataCap].Put(p.data)
	}

	p.reset()
	_packetPool.Put(p)
}

func (p *packet) reset() {
	p.Len = 0
	p.data = p.initData[:]
}

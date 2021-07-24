package server

import (
	"github.com/zengjiwen/gamenet"
	"time"
)

type options struct {
	eventChan       chan func()
	flushDelay      time.Duration
	maxFlushDelay   time.Duration
	maxFlushPackets int
	headLen         int
	littleEnd       bool
	rateLimit       *gamenet.RateLimit
}

func WithEventChan(eventChan chan func()) func(opts *options) {
	return func(opts *options) {
		opts.eventChan = eventChan
	}
}

func WithFlushDelay(flushDelay time.Duration) func(opts *options) {
	return func(opts *options) {
		opts.flushDelay = flushDelay
	}
}

func WithMaxFlushDelay(maxFlushDelay time.Duration) func(opts *options) {
	return func(opts *options) {
		opts.maxFlushDelay = maxFlushDelay
	}
}

func WithMaxFlushPackets(maxFlushPackets int) func(opts *options) {
	return func(opts *options) {
		opts.maxFlushPackets = maxFlushPackets
	}
}

func WithHeadLen(headLen int) func(opts *options) {
	return func(opts *options) {
		opts.headLen = headLen
	}
}

func WithLittleEnd() func(opts *options) {
	return func(opts *options) {
		opts.littleEnd = true
	}
}

func WithRateLimit(rateLimit *gamenet.RateLimit) func(opts *options) {
	return func(opts *options) {
		opts.rateLimit = rateLimit
	}
}

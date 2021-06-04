package server

import "gamenet"

type options struct {
	eventChan chan func()
	codec     gamenet.Codec
	marshaler gamenet.Marshaler
}

func WithEventChan(eventChan chan func()) func(opts *options) {
	return func(opts *options) {
		opts.eventChan = eventChan
	}
}

func WithCodec(codec gamenet.Codec) func(opts *options) {
	return func(opts *options) {
		opts.codec = codec
	}
}

func WithMarshaler(marshaler gamenet.Marshaler) func(opts *options) {
	return func(opts *options) {
		opts.marshaler = marshaler
	}
}

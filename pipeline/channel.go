package pipeline

import "piper"

type channelSource[T any] struct {
	in  <-chan T
	out chan any
}

func FromChannel[T any](ch <-chan T) piper.Pipeline {
	source := channelSource[T]{
		in:  ch,
		out: make(chan any),
	}
	go source.start()
	return piper.PipelineFrom(source)
}

func (c channelSource[T]) Out() <-chan any { return c.out }

func (c channelSource[T]) start() {
	defer close(c.out)
	for input := range c.in {
		c.out <- input
	}
}

type channelSink[T any] struct {
	in  chan any
	out chan<- T
}

func ToChannel[T any](ch chan<- T) piper.Sink {
	sink := channelSink[T]{
		in:  make(chan any),
		out: ch,
	}
	go sink.start()
	return sink
}

func (c channelSink[T]) In() chan<- any { return c.in }

func (c channelSink[T]) start() {
	defer close(c.out)
	for input := range c.in {
		c.out <- input.(T)
	}
}

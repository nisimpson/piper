package pipeline

import "github.com/nisimpson/piper"

// channelSource adapts a typed input channel to serve as a pipeline source.
// It converts the typed channel into the pipeline's generic any-typed channel system.
type channelSource[T any] struct {
	// in is the external typed channel from which data is read
	in <-chan T
	// out is the internal pipeline channel to which data is forwarded
	out chan any
}

// FromChannel creates a new [piper.Pipeline] from a typed channel.
// It allows existing channel-based code to be used as the input for a pipeline.
func FromChannel[T any](ch <-chan T) piper.Pipeline {
	source := channelSource[T]{
		in:  ch,
		out: make(chan any),
	}
	go source.start()
	return piper.PipelineFrom(source)
}

func (c channelSource[T]) Out() <-chan any { return c.out }

// start begins forwarding data from the input channel to the pipeline.
// It handles type conversion from T to any and ensures proper cleanup.
func (c channelSource[T]) start() {
	defer close(c.out)
	for input := range c.in {
		c.out <- input
	}
}

// channelSink adapts a typed output channel to serve as a pipeline sink.
// It converts from the pipeline's generic any-typed system back to a typed channel.
type channelSink[T any] struct {
	// in is the internal pipeline channel from which data is read
	in chan any
	// out is the external typed channel to which data is forwarded
	out chan<- T
}

// ToChannel creates a new pipeline [piper.Sink] that forwards data to a typed channel.
// It allows pipeline output to be connected to existing channel-based code.
func ToChannel[T any](ch chan<- T) piper.Sink {
	sink := channelSink[T]{
		in:  make(chan any),
		out: ch,
	}
	go sink.start()
	return sink
}

func (c channelSink[T]) In() chan<- any { return c.in }

// start begins forwarding data from the pipeline to the output channel.
// It handles type assertion from any to T and ensures proper cleanup.
func (c channelSink[T]) start() {
	defer close(c.out)
	for input := range c.in {
		c.out <- input.(T)
	}
}

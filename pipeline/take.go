package pipeline

import "github.com/nisimpson/piper"

// taker represents a pipeline stage that takes a specified number of items
// from the input stream and forwards them to the output stream.
type taker struct {
	// in is the input channel that receives items from the previous stage
	in chan any
	// out is the output channel that sends items to the next stage
	out chan any
	// count specifies the maximum number of items to take from the input
	count int
}

// TakeN creates a [piper.Pipe] that takes at most 'count' items from
// the input stream and forwards them to the output stream. Once 'count'
// items have been processed, any remaining input items are ignored.
// If count is negative, then TakeN is the equivalent of [Passthrough].
func TakeN(count int) piper.Pipe {
	pipe := taker{
		in:    make(chan any),
		out:   make(chan any),
		count: count,
	}
	go pipe.start()
	return pipe
}

// In returns the input channel for the taker stage.
// This channel is used to receive items from the previous stage in the pipeline.
func (t taker) In() chan<- any { return t.in }

// Out returns the output channel for the taker stage.
// This channel is used to send items to the next stage in the pipeline.
func (t taker) Out() <-chan any { return t.out }

// start begins the main processing loop for the taker stage.
// It reads items from the input channel and forwards them to the output
// channel until either:
//  1. The count reaches zero
//  2. The input channel is closed
//
// The output channel is automatically closed when processing is complete.
func (t taker) start() {
	defer close(t.out)
	count := t.count
	for i := range t.in {
		if count == 0 {
			continue
		}
		t.out <- i
		count--
	}
}

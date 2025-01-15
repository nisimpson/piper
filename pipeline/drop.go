package pipeline

import "github.com/nisimpson/piper"

// dropper represents a pipeline stage that drops a specified number of items
// from the input stream before forwarding remaining items to the output stream.
type dropper struct {
	// in is the input channel that receives items from the previous stage
	in chan any
	// out is the output channel that sends items to the next stage
	out chan any
	// count specifies the number of items to drop from the input
	count int
}

// DropN creates a new [piper.Pipe] that drops the first 'count' items from
// the input stream and forwards any remaining items to the output stream.
// Once 'count' items have been dropped, all subsequent items are forwarded.
// If count is negative, then DropN is the equivalent of [Passthrough].
func DropN(count int) piper.Pipe {
	pipe := dropper{
		in:    make(chan any),
		out:   make(chan any),
		count: count,
	}
	go pipe.start()
	return pipe
}

// In returns the input channel for the dropper stage.
// This channel is used to receive items from the previous stage in the pipeline.
func (d dropper) In() chan<- any { return d.in }

// Out returns the output channel for the dropper stage.
// This channel is used to send items to the next stage in the pipeline.
func (d dropper) Out() <-chan any { return d.out }

// start begins the main processing loop for the dropper stage.
// It reads items from the input channel, dropping the first 'count' items
// and forwarding all remaining items to the output channel until either:
//  1. The input channel is closed
//  2. All items have been processed
//
// The output channel is automatically closed wdhen processing is complete.
func (d dropper) start() {
	defer close(d.out)

	// If count is negative, pass through all items
	if d.count < 0 {
		for item := range d.in {
			d.out <- item
		}
		return
	}

	// Drop the first 'count' items
	dropped := 0
	for item := range d.in {
		if dropped < d.count {
			dropped++
			continue
		}
		// Forward all remaining items
		d.out <- item
	}
}

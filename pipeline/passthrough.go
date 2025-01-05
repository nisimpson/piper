package pipeline

import "piper"

// passthroughPipe implements a pipeline component that forwards items without modification.
// It acts as a simple relay between pipeline segments.
type passthroughPipe struct {
	// in receives items to be forwarded
	in chan any
	// out sends received items without modification
	out chan any
}

// Passthrough creates a new [piper.Pipe] component that forwards items without modification.
// This can be useful for debugging or when you need to maintain the pipeline structure without processing.
func Passthrough() piper.Pipe {
	pipe := passthroughPipe{
		in:  make(chan any),
		out: make(chan any),
	}
	go pipe.start()
	return pipe
}

func (p passthroughPipe) In() chan<- any  { return p.in }
func (p passthroughPipe) Out() <-chan any { return p.out }

// start begins forwarding items from the input channel to the output channel.
// Each item is passed through unchanged.
func (p passthroughPipe) start() {
	defer close(p.out)
	for v := range p.in {
		p.out <- v
	}
}

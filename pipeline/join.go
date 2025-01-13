package pipeline

import "github.com/nisimpson/piper"

// joinedPipe represents a composite pipe that connects two pipes together,
// where data flows from the source pipe to the target pipe.
type joinedPipe struct {
	source piper.Pipe
	target piper.Pipe
}

// Join connects multiple pipes together in sequence, where the output of each pipe
// feeds into the input of the next pipe. Returns a single composite [piper.Pipe] that
// represents the entire chain.
func Join(src piper.Pipe, into ...piper.Pipe) piper.Pipe {
	// join each pipe in indexed order
	for _, tgt := range into {
		src = newJoinedPipe(src, tgt)
	}
	return src
}

// newJoinedPipe creates a new joinedPipe instance that connects the source pipe
// to the target pipe and starts the data flow between them.
func newJoinedPipe(src, tgt piper.Pipe) joinedPipe {
	pipe := joinedPipe{source: src, target: tgt}
	go pipe.start()
	return pipe
}

// In returns the input channel of the joined pipe, which is the input channel
// of the source pipe.
func (p joinedPipe) In() chan<- any { return p.source.In() }

// Out returns the output channel of the joined pipe, which is the output channel
// of the target pipe.
func (p joinedPipe) Out() <-chan any { return p.target.Out() }

// start begins the process of moving data from the source pipe to the target pipe.
// It ensures proper cleanup by closing the target's input channel when complete.
func (p joinedPipe) start() {
	defer close(p.target.In())
	for input := range p.source.Out() {
		p.target.In() <- input
	}
}

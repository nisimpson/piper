package pipeline

import "piper"

type passthroughPipe struct {
	in  chan any
	out chan any
}

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

func (p passthroughPipe) start() {
	defer close(p.out)
	for v := range p.in {
		p.out <- v
	}
}

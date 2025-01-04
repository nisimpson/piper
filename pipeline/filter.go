package pipeline

import "piper"

type FilterFunction[In any] func(In) bool

type filterPipe[In any] struct {
	in         chan any
	out        chan any
	filterFunc FilterFunction[In]
}

func Filter[In any](fn FilterFunction[In]) piper.Pipe {
	pipe := filterPipe[In]{
		in:         make(chan any),
		out:        make(chan any),
		filterFunc: fn,
	}

	go pipe.start()
	return pipe
}

func KeepIf[In any](fn FilterFunction[In]) piper.Pipe {
	return Filter(fn)
}

func DropIf[In any](fn FilterFunction[In]) piper.Pipe {
	return Filter(func(in In) bool { return !fn(in) })
}

func (f filterPipe[In]) In() chan<- any  { return f.in }
func (f filterPipe[In]) Out() <-chan any { return f.out }

func (f filterPipe[In]) start() {
	defer close(f.out)
	for input := range f.in {
		if !f.filterFunc(input.(In)) {
			// drop and do not pass downstream
			continue
		}
		f.out <- input
	}
}

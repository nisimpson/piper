package pipeline

import "piper"

type flatmapper[In any, Out any] struct {
	in          chan any
	out         chan any
	mapFunction MapFunction[In, []Out]
}

func FlatMap[In any, Out any](fn MapFunction[In, []Out]) piper.Pipe {
	pipe := flatmapper[In, Out]{
		in:          make(chan any),
		out:         make(chan any),
		mapFunction: fn,
	}
	go pipe.start()
	return pipe
}

func (f flatmapper[In, Out]) In() chan<- any  { return f.in }
func (f flatmapper[In, Out]) Out() <-chan any { return f.out }

func (f flatmapper[In, Out]) start() {
	defer close(f.out)
	for input := range f.in {
		items := f.mapFunction(input.(In))
		for _, item := range items {
			f.out <- item
		}
	}
}

package pipeline

import (
	"piper"
)

type MapFunction[T any, U any] func(T) U

type mapper[In any, Out any] struct {
	in        chan any
	out       chan any
	transform MapFunction[In, Out]
}

func Map[In any, Out any](fn MapFunction[In, Out]) piper.Pipe {
	pipe := mapper[In, Out]{
		in:        make(chan any),
		out:       make(chan any),
		transform: fn,
	}

	go pipe.start()
	return pipe
}

func (m mapper[In, Out]) In() chan<- any {
	return m.in
}

func (m mapper[In, Out]) Out() <-chan any {
	return m.out
}

func (m mapper[In, Out]) start() {
	defer close(m.out)
	for input := range m.in {
		// execute the transformation
		output := m.transform(input.(In))

		// send along
		m.out <- output
	}
}

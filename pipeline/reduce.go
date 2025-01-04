package pipeline

import "piper"

type ReduceFunction[T any] func(acc T, item T) T

type reducer[T any] struct {
	in             chan any
	out            chan any
	reduceFunction ReduceFunction[T]
	acc            any
}

func Reduce[T any](fn ReduceFunction[T]) piper.Pipe {
	pipe := &reducer[T]{
		in:             make(chan any),
		out:            make(chan any),
		reduceFunction: fn,
	}

	go pipe.start()
	return pipe
}

func (r *reducer[T]) In() chan<- any {
	return r.in
}

func (r *reducer[T]) Out() <-chan any {
	return r.out
}

func (r *reducer[T]) start() {
	defer close(r.out)

	for item := range r.in {
		if r.acc == nil {
			r.acc = item
			r.out <- item
			continue
		}
		acc := r.reduceFunction(r.acc.(T), item.(T))
		r.acc = acc
		r.out <- acc
	}
}

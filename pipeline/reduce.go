package pipeline

import "piper"

// ReduceFunction represents a function that combines two values of the same type into one.
// acc is the accumulated result so far, and item is the next item to combine into the result.
type ReduceFunction[T any] func(acc T, item T) T

// reducer implements a pipeline component that combines multiple items into a single accumulated result.
// It processes items one at a time, maintaining and updating an accumulator value.
type reducer[T any] struct {
	// in receives items to be reduced.
	in chan any
	// out sends the current accumulated value after each reduction.
	out chan any
	// reduceFunction combines the current accumulator with each new item.
	reduceFunction ReduceFunction[T]
	// acc holds the current accumulated value.
	acc any
}

// Reduce creates a new [piper.Pipe] component that combines multiple items into one using the provided function.
// The function is called for each item with the current accumulated value and the new item.
func Reduce[T any](fn ReduceFunction[T]) piper.Pipe {
	pipe := &reducer[T]{
		in:             make(chan any),
		out:            make(chan any),
		reduceFunction: fn,
	}

	go pipe.start()
	return pipe
}

func (r *reducer[T]) In() chan<- any  { return r.in }
func (r *reducer[T]) Out() <-chan any { return r.out }

// start begins the reduction process, combining items one at a time and sending the current
// accumulated value downstream after each combination.
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

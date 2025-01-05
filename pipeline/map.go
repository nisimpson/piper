package pipeline

import (
	"piper"
)

// MapFunction represents a function that transforms an item from one type to another.
// T is the input type and U is the output type.
type MapFunction[T any, U any] func(T) U

// mapper implements a pipeline component that transforms items using a mapping function.
type mapper[In any, Out any] struct {
	// in receives items to be transformed
	in chan any
	// out sends transformed items
	out chan any
	// transform is the function that converts items from type In to type Out
	transform MapFunction[In, Out]
}

// Map creates a new [piper.Pipe] component that transforms items using the provided function.
// Each input item is transformed from type In to type Out using the [MapFunction] fn.
func Map[In any, Out any](fn MapFunction[In, Out]) piper.Pipe {
	pipe := mapper[In, Out]{
		in:        make(chan any),
		out:       make(chan any),
		transform: fn,
	}

	go pipe.start()
	return pipe
}

func (m mapper[In, Out]) In() chan<- any  { return m.in }
func (m mapper[In, Out]) Out() <-chan any { return m.out }

// start begins the transformation process, converting each input item to an output item
// using the mapping function. Each transformed item is sent downstream.
func (m mapper[In, Out]) start() {
	defer close(m.out)
	for input := range m.in {
		// execute the transformation
		output := m.transform(input.(In))

		// send along
		m.out <- output
	}
}

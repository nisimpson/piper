package pipeline

import "github.com/nisimpson/piper"

// flatmapper implements a pipeline component that transforms each input item into multiple output items.
// It executes a mapping function that returns a slice, then sends each element of that slice downstream individually.
type flatmapper[In any, Out any] struct {
	// in receives items to be transformed
	in chan any
	// out sends the transformed items
	out chan any
	// mapFunction converts each input item into a slice of output items
	mapFunction MapFunction[In, []Out]
}

// FlatMap creates a new [piper.Pipe] component that transforms items using the provided [MapFunction].
// Each input item is transformed into a slice of output items, which are then sent individually downstream.
func FlatMap[In any, Out any](fn MapFunction[In, []Out]) piper.Pipe {
	pipe := flatmapper[In, Out]{
		in:          make(chan any),
		out:         make(chan any),
		mapFunction: fn,
	}
	go pipe.start()
	return pipe
}

// Flatten creates a new [piper.Pipe] component that receives a batch of items and flattens it,
// sending each item from the upstream batch individually downstream.
// It is the semantic equivalent of:
//
//	FlatMap(func(in []T) []T { return in })
//
// For example, to handle an upstream csv in indexed row order, call:
//
//	// flatten a 2-D array of strings
//	Flatten[[][]string]() // [][]string - Flatten -> []string
func Flatten[In []Out, Out any]() piper.Pipe {
	return FlatMap(func(in In) []Out { return in })
}

func (f flatmapper[In, Out]) In() chan<- any  { return f.in }
func (f flatmapper[In, Out]) Out() <-chan any { return f.out }

// start begins the flat mapping process, transforming each input item into multiple output items.
// Each item in the output slice is sent individually downstream.
func (f flatmapper[In, Out]) start() {
	defer close(f.out)
	for input := range f.in {
		items := f.mapFunction(input.(In))
		for _, item := range items {
			f.out <- item
		}
	}
}

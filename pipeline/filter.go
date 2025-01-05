package pipeline

import "piper"

// FilterFunction represents a predicate that determines whether an item should be included in the output.
// It takes an item of type In and returns true if the item should be kept, false if it should be dropped.
type FilterFunction[In any] func(In) bool

// filterPipe implements a pipeline component that selectively passes items based on a filter function.
type filterPipe[In any] struct {
	// in receives items to be filtered
	in chan any
	// out sends items that pass the filter
	out chan any
	// filterFunc determines which items to keep
	filterFunc FilterFunction[In]
}

// Filter creates a new pipeline component that uses the provided [FilterFunction] to filter items.
// Only items for which fn returns true will be passed downstream.
func Filter[In any](fn FilterFunction[In]) piper.Pipe {
	pipe := filterPipe[In]{
		in:         make(chan any),
		out:        make(chan any),
		filterFunc: fn,
	}

	go pipe.start()
	return pipe
}

// KeepIf is an alias for [Filter] that creates a more readable pipeline when the intent
// is to keep items that match a condition.
func KeepIf[In any](fn FilterFunction[In]) piper.Pipe {
	return Filter(fn)
}

// DropIf creates a [Filter] that drops items matching the condition and keeps everything else.
// It inverts the behavior of the provided [FilterFunction].
func DropIf[In any](fn FilterFunction[In]) piper.Pipe {
	return Filter(func(in In) bool { return !fn(in) })
}

func (f filterPipe[In]) In() chan<- any  { return f.in }
func (f filterPipe[In]) Out() <-chan any { return f.out }

// start begins the filtering process, passing through only the items that satisfy the filter function.
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

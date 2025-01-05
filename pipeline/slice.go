package pipeline

import (
	"piper"
	"sync"
)

// source represents a pipeline source that sends items from a slice.
type source struct {
	// out is the channel where slice items are sent
	out chan any
}

// FromSlice creates a new [piper.Pipeline] that starts with the provided slice items.
// Items are sent one at a time through the pipeline in the order they appear in the slice.
func FromSlice[In any](items ...In) piper.Pipeline {
	source := source{
		out: make(chan any, len(items)),
	}
	for _, item := range items {
		source.out <- item
	}
	close(source.out)
	return piper.PipelineFrom(source)
}

func (s source) Out() <-chan any { return s.out }

// sink implements a pipeline sink that collects all received items into a slice.
// It provides synchronization capabilities to wait for and access the final slice.
type sink[In any] struct {
	// wg is used to signal when all items have been collected
	wg sync.WaitGroup
	// in receives items to be collected
	in chan any
	// output holds all received items in order
	output []In
}

// ToSlice creates a new [piper.Sink] that collects all pipeline items into a slice.
// The slice can be accessed using the [Slice] method after the pipeline completes.
func ToSlice[In any]() *sink[In] {
	sink := &sink[In]{
		wg: sync.WaitGroup{},
		in: make(chan any),
	}

	sink.wg.Add(1)
	go sink.start()
	return sink
}

func (s *sink[In]) In() chan<- any { return s.in }

// Slice waits for all items to be collected and returns them as a slice.
// This method will block until the pipeline has finished processing all items.
func (s *sink[In]) Slice() []In {
	s.wg.Wait()
	return s.output
}

// start begins collecting items from the input channel into a slice.
// Each received item is appended to the output slice in order.
func (s *sink[In]) start() {
	defer s.wg.Done()
	for data := range s.in {
		s.output = append(s.output, data.(In))
	}
}

package pipeline

import (
	"piper"
	"sync"
)

type source struct {
	out chan any
}

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

func (s source) Out() <-chan any {
	return s.out
}

type sink[In any] struct {
	wg     sync.WaitGroup
	in     chan any
	output []In
}

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

func (s *sink[In]) Slice() []In {
	s.wg.Wait()
	return s.output
}

func (s *sink[In]) start() {
	defer s.wg.Done()
	for data := range s.in {
		s.output = append(s.output, data.(In))
	}
}

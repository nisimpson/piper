package pipeline_test

import (
	"sync"

	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Fixture[In any] struct {
	wg    sync.WaitGroup
	in    chan any
	items []In
}

func NewFixture[In any]() *Fixture[In] {
	sink := Fixture[In]{
		in:    make(chan any),
		wg:    sync.WaitGroup{},
		items: make([]In, 0),
	}
	sink.wg.Add(1)
	go sink.start()
	return &sink
}

func Consume[In any](s piper.Source) []In {
	f := NewFixture[In]()
	return f.consume(pipeline.From(s))
}

func (s *Fixture[In]) In() chan<- any { return s.in }

func (s *Fixture[In]) consume(pipeline pipeline.Flow) []In {
	pipeline.To(s)
	s.wg.Wait()
	return s.items
}

func (s *Fixture[In]) start() {
	defer s.wg.Done()
	for item := range s.in {
		s.items = append(s.items, item.(In))
	}
}

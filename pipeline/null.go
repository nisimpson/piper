package pipeline

import (
	"sync"
)

type nullSink struct {
	wg *sync.WaitGroup
	in chan any
}

func ToNull() nullSink {
	sink := nullSink{
		in: make(chan any),
		wg: &sync.WaitGroup{},
	}
	sink.wg.Add(1)
	go sink.start()
	return sink
}

func (n nullSink) Wait()          { n.wg.Wait() }
func (n nullSink) In() chan<- any { return n.in }
func (nullSink) noop(any)         {}

func (n nullSink) start() {
	defer n.wg.Done()
	for i := range n.in {
		n.noop(i)
	}
}

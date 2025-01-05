package pipeline

import (
	"sync"
)

// nullSink implements a pipeline sink that discards all received items.
// It provides synchronization capabilities to wait for all items to be processed.
type nullSink struct {
	// wg is used to signal when all items have been processed
	wg *sync.WaitGroup
	// in receives items to be discarded
	in chan any
}

// ToNull creates a new sink that discards all items it receives.
// This is useful when you want to execute a pipeline but don't need its output.
func ToNull() nullSink {
	sink := nullSink{
		in: make(chan any),
		wg: &sync.WaitGroup{},
	}
	sink.wg.Add(1)
	go sink.start()
	return sink
}

// Wait blocks until all items have been processed and discarded
func (n nullSink) Wait()          { n.wg.Wait() }
// In returns the channel used to send items to be discarded
func (n nullSink) In() chan<- any { return n.in }
// noop is an empty operation that consumes an item without doing anything
func (nullSink) noop(any)         {}

// start begins consuming and discarding items from the input channel.
// It signals completion through the WaitGroup when all items have been processed.
func (n nullSink) start() {
	defer n.wg.Done()
	for i := range n.in {
		n.noop(i)
	}
}

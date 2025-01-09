package pipeline

import (
	"sync"

	"github.com/nisimpson/piper"
)

// ParallelPipeFactory creates and returns a new [piper.Pipe].
// This is used to generate multiple identical pipes for parallel processing.
type ParallelPipeFactory = func() piper.Pipe

// parallelizer implements a parallel processing [piper.Pipe] that distributes work
// across multiple identical pipes running concurrently.
type parallelizer struct {
	in        chan any            // in is the input channel that receives data to be processed
	out       chan any            // out is the output channel that sends processed results
	size      int                 // size determines the number of parallel pipes to create
	generator ParallelPipeFactory // generator is the function used to create new pipes for parallel processing
}

// Parallelize creates a [piper.Pipe] that processes data concurrently across
// multiple identical pipes. It distributes incoming data across 'size' number
// of worker pipes created with the [ParallelPipeFactory]. Note that this can
// potentially alter the initial ordering of items to a pipe or [piper.Sink]
// downstream.
//
// Parallelize panics if size is less than 1.
func Parallelize(size int, factory ParallelPipeFactory) piper.Pipe {
	if size < 1 {
		panic("parallelize size must be greater than 0")
	}

	if size == 1 {
		// no parallelization, return a single generated pipe
		return factory()
	}

	pipe := &parallelizer{
		in:        make(chan any),
		out:       make(chan any),
		size:      size,
		generator: factory,
	}

	go pipe.start()
	return pipe
}

// In returns the input channel for the parallelizer.
// This channel is used to send data into the parallel processing pipeline.
func (p parallelizer) In() chan<- any { return p.in }

// Out returns the output channel for the parallelizer.
// This channel receives the results from all parallel processing pipes.
func (p parallelizer) Out() <-chan any { return p.out }

// start initializes and manages the parallel processing operation.
// It creates the specified number of worker pipes and coordinates their execution.
// The function ensures all workers are properly started and cleaned up.
func (p parallelizer) start() {
	wg := sync.WaitGroup{}
	defer close(p.out)

	// create size number of workers, and immediately put them to work
	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go p.work(p.generator(), &wg)
	}

	// wait for all work to be completed.
	wg.Wait()
}

// work manages an individual worker pipe in the parallel processing system.
// It handles the flow of data through a single pipe and ensures proper cleanup.
//
// Parameters:
//   - pipe: The individual pipe instance to process data
//   - wg: WaitGroup for coordinating worker completion
func (p parallelizer) work(pipe piper.Pipe, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(pipe.In())
	for input := range p.in {
		pipe.In() <- input  // process upstream input
		p.out <- pipe.Out() // send output downstream
	}
}

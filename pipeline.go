package piper

import "sync"

// Pipeline represents a data processing pipeline that connects various components together.
// It manages the flow of data from a source through optional intermediate processing steps
// to eventual sinks.
type Pipeline struct {
	// outlet is the current endpoint from which data flows in this pipeline segment
	outlet Outlet
}

// PipelineFrom creates a new pipeline starting from the given source.
// This is typically used as the entry point for constructing a new pipeline.
func PipelineFrom(source Source) Pipeline {
	return Pipeline{outlet: source}
}

// Then adds one or more processing steps to the pipeline.
// Each pipe is connected in sequence, with data flowing from one to the next.
// Returns a new Pipeline instance representing the updated pipeline.
func (p Pipeline) Then(pipes ...Pipe) Pipeline {
	for _, pipe := range pipes {
		go p.transmit(pipe)
		p = Pipeline{outlet: pipe}
	}
	return p
}

// To connects a sink to the end of the pipeline.
// This is typically the final step in pipeline construction, establishing
// where the processed data will ultimately be delivered.
func (p Pipeline) To(sink Sink) {
	go p.transmit(sink)
}

// Tee splits the pipeline into two branches.
// The same data will be sent to both pipe1 and pipe2, allowing for parallel processing paths.
// Returns two new Pipeline instances, one for each branch.
func (p Pipeline) Tee(pipe1, pipe2 Pipe) (Pipeline, Pipeline) {
	go p.tee(pipe1, pipe2)
	return Pipeline{outlet: pipe1}, Pipeline{outlet: pipe2}
}

// Out returns the output channel of the pipeline.
// This channel can be used to read processed data directly from the pipeline.
func (p Pipeline) Out() <-chan any {
	return p.outlet.Out()
}

// transmit handles the movement of data from the pipeline's current outlet to the given inlet.
// It ensures proper cleanup by closing the inlet's channel when transmission is complete.
func (p Pipeline) transmit(in Inlet) {
	defer close(in.In())
	for b := range p.outlet.Out() {
		in.In() <- b
	}
}

// tee is an internal helper function that implements the data duplication logic for the Tee method.
// It reads from the pipeline's outlet and sends each item to both input channels.
func (p Pipeline) tee(in1, in2 Inlet) {
	send := func(wg *sync.WaitGroup, in Inlet, data any) {
		defer wg.Done()
		in.In() <- data
	}

	for b := range p.outlet.Out() {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go send(wg, in1, b)
		go send(wg, in2, b)
		wg.Wait()
	}

	close(in1.In())
	close(in2.In())
}

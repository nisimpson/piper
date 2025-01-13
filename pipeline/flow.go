package pipeline

import (
	"context"
	"sync"

	"github.com/nisimpson/piper"
)

// Flow represents a data processing pipeline that connects various components together.
// It manages the flow of data from a [piper.Source] through optional intermediate [piper.Pipe] processing steps
// to a [piper.Sink].
type Flow struct {
	// outlet is the current endpoint from which data flows in this pipeline segment.
	outlet piper.Outlet
	// ctx is the context associated with this pipeline, used for cancellation and other context-related operations.
	// It is propagated to downstream components in the pipeline.
	ctx context.Context
}

// From creates a new pipeline starting from the given source.
// This is typically used as the entry point for constructing a new pipeline.
func From(source piper.Source) Flow {
	ctx := context.TODO()
	if pipeline, ok := source.(Flow); ok {
		ctx = pipeline.ctx
	}
	return Flow{outlet: source, ctx: ctx}
}

// WithContext adds the target context to this [Flow]. If you want this pipeline to support
// cancellation, you must call this method before adding a [Pipe] or [Sink].
func (f Flow) WithContext(ctx context.Context) Flow {
	f.ctx = ctx
	return f
}

// Thru adds one or more processing steps to the pipeline.
// Each [Pipe] is connected in sequence (indexed order), with data flowing from one to the next.
// Returns a new [Flow] instance representing the updated pipeline.
func (f Flow) Thru(pipes ...piper.Pipe) Flow {
	for _, pipe := range pipes {
		go f.transmit(pipe)
		f = From(pipe).WithContext(f.ctx)
	}
	return f
}

// To connects a [Sink] to the end of the [Flow].
// This is typically the final step in pipeline construction, establishing
// where the processed data will ultimately be delivered.
func (f Flow) To(sink piper.Sink) {
	go f.transmit(sink)
}

// Tee splits the pipeline into two branches.
// The same data will be sent to both pipe1 and pipe2, allowing for parallelized processing paths.
// Returns two new [Flow] instances, one for each branch.
func (f Flow) Tee(pipe1, pipe2 piper.Pipe) (Flow, Flow) {
	go f.tee(pipe1, pipe2)
	var (
		f1 = From(pipe1).WithContext(f.ctx)
		f2 = From(pipe2).WithContext(f.ctx)
	)
	return f1, f2
}

// Out returns the output channel of the [Flow].
// This channel can be used to read processed data directly from the pipeline, and allows
// this flow to act as the [piper.Source] of another flow.
func (f Flow) Out() <-chan any {
	return f.outlet.Out()
}

// transmit handles the movement of data from the pipeline's current outlet to the given inlet.
// It ensures proper cleanup by closing the inlet's channel when transmission is complete.
func (f Flow) transmit(in piper.Inlet) {
	defer close(in.In())
	for {
		b, ok := <-f.outlet.Out()
		if !ok {
			return
		}
		select {
		case <-f.ctx.Done():
			return
		case in.In() <- b:
		}
	}
}

// tee is an internal helper function that implements the data duplication logic for the Tee method.
// It reads from the pipeline's outlet and sends each item to both input channels.
func (p Flow) tee(in1, in2 piper.Inlet) {
	send := func(wg *sync.WaitGroup, in piper.Inlet, data any) {
		defer wg.Done()
		in.In() <- data
	}

	defer close(in1.In())
	defer close(in2.In())

	for {
		b, ok := <-p.outlet.Out()
		if !ok {
			return
		}
		select {
		case <-p.ctx.Done():
			return
		default:
			wg := &sync.WaitGroup{}
			wg.Add(2)
			go send(wg, in1, b)
			go send(wg, in2, b)
			wg.Wait()
		}
	}
}

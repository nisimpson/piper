package piper

import "sync"

type Pipeline struct {
	outlet Outlet
}

func PipelineFrom(source Source) Pipeline {
	return Pipeline{outlet: source}
}

func (p Pipeline) Then(pipes ...Pipe) Pipeline {
	for _, pipe := range pipes {
		go p.transmit(pipe)
		p = Pipeline{outlet: pipe}
	}
	return p
}

func (p Pipeline) To(sink Sink) {
	go p.transmit(sink)
}

func (p Pipeline) Tee(pipe1, pipe2 Pipe) (Pipeline, Pipeline) {
	go p.tee(pipe1, pipe2)
	return Pipeline{outlet: pipe1}, Pipeline{outlet: pipe2}
}

func (p Pipeline) Out() <-chan any {
	return p.outlet.Out()
}

func (p Pipeline) transmit(in Inlet) {
	defer close(in.In())
	for b := range p.outlet.Out() {
		in.In() <- b
	}
}

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

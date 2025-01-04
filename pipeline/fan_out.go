package pipeline

import "piper"

type KeyFunction[T any] func(T) string

type FanOutPipelineFunction = func(source piper.Source) piper.Pipeline

type fanOutSink[In any] struct {
	in          chan any
	generators  map[string]FanOutPipelineFunction
	keyFunction KeyFunction[In]
	sources     []piper.Source
	channels    map[string]chan In
}

func FanOut[In any](keyfn KeyFunction[In], generators map[string]FanOutPipelineFunction) fanOutSink[In] {
	sink := fanOutSink[In]{
		in:          make(chan any),
		keyFunction: keyfn,
		generators:  generators,
		sources:     make([]piper.Source, 0, len(generators)),
		channels:    make(map[string]chan In),
	}

	for key, generator := range generators {
		var (
			channel  = make(chan In)
			pipeline = FromChannel(channel)
		)
		sink.channels[key] = channel
		sink.sources = append(sink.sources, generator(pipeline))
	}

	go sink.start()
	return sink
}

func (f fanOutSink[In]) Sources() []piper.Source { return f.sources }
func (f fanOutSink[In]) In() chan<- any          { return f.in }

func (f fanOutSink[In]) start() {
	for _, ch := range f.channels {
		defer close(ch)
	}
	for input := range f.in {
		item := input.(In)
		key := f.keyFunction(item)
		channel, ok := f.channels[key]
		if !ok {
			continue
		}
		channel <- item
	}
}

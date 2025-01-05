package pipeline

import "piper"

// KeyFunction is a function that determines the destination branch for each item in the fan-out.
// It takes an item of type T and returns a string key identifying the target branch.
type KeyFunction[T any] func(T) string

// FanOutPipelineFunction represents a function that constructs a pipeline segment for a fan-out branch.
// It takes a source and returns a pipeline that will process items sent to that branch.
type FanOutPipelineFunction = func(source piper.Source) piper.Pipeline

// fanOutSink implements a pipeline sink that distributes incoming items to multiple branches
// based on a key function. Each branch can have its own processing pipeline.
type fanOutSink[In any] struct {
	// in receives items to be distributed
	in chan any
	// generators maps branch keys to functions that create the processing pipeline for that branch
	generators map[string]FanOutPipelineFunction
	// keyFunction determines which branch should receive each item
	keyFunction KeyFunction[In]
	// sources holds the source end of each branch's pipeline
	sources []piper.Source
	// channels maps branch keys to the channels used to send items to each branch
	channels map[string]chan In
}

// ToMultiSource creates a fan-out sink that distributes items to multiple pipeline branches.
// The keyfn determines which branch receives each item, and generators provide the processing pipeline for each branch.
func ToMultiSource[In any](keyfn KeyFunction[In], generators map[string]FanOutPipelineFunction) fanOutSink[In] {
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

// Sources returns the source ends of all branch pipelines
func (f fanOutSink[In]) Sources() []piper.Source { return f.sources }
// In returns the channel used to send items into the fan-out sink
func (f fanOutSink[In]) In() chan<- any          { return f.in }

// start begins distributing incoming items to their appropriate branches based on the key function.
// It ensures proper cleanup by closing all branch channels when the input is exhausted.
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

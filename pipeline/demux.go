package pipeline

import "piper"

// DemuxKeyFunction is a function that determines the destination branch for each item upstream.
// It takes an item of type T and returns a string key identifying the target branch.
type DemuxKeyFunction[T any] func(T) string

// DemuxPipelineFunction represents a function that constructs a pipeline segment for a fork branch.
// It takes a [piper.Source] and returns a [piper.Pipeline] that will process items sent to that branch.
type DemuxPipelineFunction = func(source piper.Source) piper.Pipeline

// demuxer implements a pipeline sink that distributes incoming items to multiple branches
// based on a key function. Each branch can have its own processing pipeline.
type demuxer[In any] struct {
	// in receives items to be distributed.
	in chan any
	// generators maps branch keys to functions that create the processing pipeline for that branch.
	generators map[string]DemuxPipelineFunction
	// keyFunction determines which branch should receive each item.
	keyFunction DemuxKeyFunction[In]
	// sources holds the source end of each branch's pipeline.
	sources []piper.Source
	// channels maps branch keys to the channels used to send items to each branch.
	channels map[string]chan In
}

// Demux creates a fan-out [piper.Sink] that distributes items to multiple [piper.Pipeline] branches.
// The [DemuxKeyFunction] keyfn determines which branch receives each item, and [DemuxPipelineFunction] generators
// provide the processing pipeline for each branch.
func Demux[In any](keyfn DemuxKeyFunction[In], generators map[string]DemuxPipelineFunction) demuxer[In] {
	sink := demuxer[In]{
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

// Sources returns the source ends of all branch pipelines.
func (d demuxer[In]) Sources() []piper.Source { return d.sources }

// In returns the channel used to send items into the fan-out sink.
func (d demuxer[In]) In() chan<- any { return d.in }

// start begins distributing incoming items to their appropriate branches based on the key function.
// It ensures proper cleanup by closing all branch channels when the input is exhausted.
func (d demuxer[In]) start() {
	for _, ch := range d.channels {
		defer close(ch)
	}
	for input := range d.in {
		item := input.(In)
		key := d.keyFunction(item)
		channel, ok := d.channels[key]
		if !ok {
			continue
		}
		channel <- item
	}
}

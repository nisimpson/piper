package pipeline

import (
	"math/rand"
	"piper"
	"slices"
)

// fanInSource implements a pipeline source that combines multiple input sources into a single output stream.
type fanInSource struct {
	// out is the channel where combined data from all sources is sent
	out chan any
	// sources is the collection of input sources to read from
	sources []piper.Source
}

// FromMultiSource creates a new [piper.Pipeline] that reads from multiple sources simultaneously.
// Data from all sources is interleaved into a single source stream.
func FromMultiSource(sources ...piper.Source) piper.Pipeline {
	fanin := fanInSource{
		out: make(chan any),
	}
	fanin.sources = append(fanin.sources, sources...)
	go fanin.start()
	return piper.PipelineFrom(fanin)
}

// Out returns the channel containing the combined output from all sources.
func (f fanInSource) Out() <-chan any { return f.out }

// start begins reading from all sources and combining their output into a single stream.
// It continues until all sources are exhausted.
func (f fanInSource) start() {
	defer close(f.out)
	var (
		sources = f.sources
	)
	for len(sources) > 0 {
		idx := rand.Intn(len(sources))
		output, next := <-sources[idx].Out()
		if !next {
			sources = slices.Delete(sources, idx, idx+1)
			continue
		}
		f.out <- output
	}
}

package pipeline

import (
	"math/rand"
	"piper"
	"slices"
)

// muxer implements a pipeline source that combines multiple input sources into a single output stream.
type muxer struct {
	// out is the channel where combined data from all sources is sent
	out chan any
	// sources is the collection of input sources to read from
	sources []piper.Source
}

// Mux creates a new [piper.Pipeline] that reads from multiple sources simultaneously.
// Data from all sources is interleaved into a single source stream.
func Mux(sources ...piper.Source) piper.Pipeline {
	fanin := muxer{
		out: make(chan any),
	}
	fanin.sources = append(fanin.sources, sources...)
	go fanin.start()
	return piper.PipelineFrom(fanin)
}

// Out returns the channel containing the combined output from all sources.
func (m muxer) Out() <-chan any { return m.out }

// start begins reading from all sources and combining their output into a single stream.
// It continues until all sources are exhausted.
func (m muxer) start() {
	defer close(m.out)
	var (
		sources = m.sources
	)
	for len(sources) > 0 {
		idx := rand.Intn(len(sources))
		output, next := <-sources[idx].Out()
		if !next {
			sources = slices.Delete(sources, idx, idx+1)
			continue
		}
		m.out <- output
	}
}

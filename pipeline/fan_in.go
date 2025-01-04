package pipeline

import (
	"math/rand"
	"piper"
	"slices"
)

type fanInSource struct {
	out     chan any
	sources []piper.Source
}

func FanIn(sources ...piper.Source) piper.Pipeline {
	fanin := fanInSource{
		out: make(chan any),
	}
	fanin.sources = append(fanin.sources, sources...)
	go fanin.start()
	return piper.PipelineFrom(fanin)
}

func (f fanInSource) Out() <-chan any { return f.out }

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

package pipeline_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

func TestDemux(t *testing.T) {
	t.Parallel()

	var (
		source = pipeline.FromSlice(0, 1, 2, 3, 4)
		double = func(in int) int { return in * 2 }
		triple = func(in int) int { return in * 3 }
	)

	generators := map[string]pipeline.DemuxPipelineFunction{
		"evens": func(s piper.Source) piper.Pipeline {
			return piper.PipelineFrom(s).Thru(pipeline.Map(double))
		},
		"odds": func(s piper.Source) piper.Pipeline {
			return piper.PipelineFrom(s).Thru(pipeline.Map(triple))
		},
	}

	keyFn := func(in int) string {
		if in == 0 {
			// "zeros" is a key without an associated generator, so any value
			// emitting this key will be dropped by the pipeline.
			return "zeros"
		}
		switch in % 2 {
		case 0:
			return "evens"
		default:
			return "odds"
		}
	}

	sink := pipeline.Demux(keyFn, generators)
	source.To(sink)

	sources := sink.Sources()
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
		return
	}

	var (
		out1 = Consume[int](sources[0])
		out2 = Consume[int](sources[1])
		got  = make([]int, 0)
	)

	got = append(got, out1...)
	got = append(got, out2...)

	// evens doubled, odds tripled
	want := []int{3, 4, 9, 8}

	// sort for deep comparison
	slices.Sort(want)
	slices.Sort(got)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

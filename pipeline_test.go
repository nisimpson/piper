package piper_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

func TestPipeline(t *testing.T) {
	t.Run("source to sink", func(t *testing.T) {
		sink := pipeline.ToSlice[int]()
		piper.PipelineFrom(pipeline.FromSlice(1, 2, 3, 4)).To(sink)
		want := []int{1, 2, 3, 4}
		got := sink.Slice()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("source then action to sink", func(t *testing.T) {
		var (
			source = piper.PipelineFrom(pipeline.FromSlice(1, 2, 3, 4))
			action = pipeline.Map(func(i int) int { return i * 2 })
			sink   = pipeline.ToSlice[int]()
		)

		source.Then(action).To(sink)

		var (
			want = []int{2, 4, 6, 8}
			got  = sink.Slice()
		)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tee into two pipelines", func(t *testing.T) {
		var (
			sink1        = pipeline.ToSlice[int]()
			sink2        = pipeline.ToSlice[int]()
			pipe1, pipe2 = piper.PipelineFrom(pipeline.FromSlice(1, 2, 3, 4)).Tee(
				pipeline.Passthrough(),
				pipeline.Passthrough(),
			)
		)

		pipe1.To(sink1)
		pipe2.To(sink2)

		var (
			want = []int{1, 2, 3, 4}
			got1 = sink1.Slice()
			got2 = sink2.Slice()
		)

		if !reflect.DeepEqual(got1, want) {
			t.Errorf("got %v, want %v", got1, want)
		}
		if !reflect.DeepEqual(got2, want) {
			t.Errorf("got %v, want %v", got2, want)
		}
	})
}

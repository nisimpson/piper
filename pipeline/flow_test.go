package pipeline_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestFlow(t *testing.T) {
	t.Parallel()

	t.Run("source to sink", func(t *testing.T) {
		sink := pipeline.ToSlice[int]()
		source := pipeline.FromSlice(1, 2, 3, 4)
		pipeline.From(source).To(sink)
		want := []int{1, 2, 3, 4}
		got := sink.Slice()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("cancel transmission with context", func(t *testing.T) {
		var (
			source = pipeline.FromSlice(1, 2, 3, 4)
			action = pipeline.Map(func(i int) int { return i * 2 })
			sink   = pipeline.ToSlice[int]()
		)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		source.WithContext(ctx).Thru(action).To(sink)

		var (
			got = sink.Slice()
		)

		if len(got) == 4 {
			t.Errorf("got %v", got)
		}
	})

	t.Run("source then action to sink", func(t *testing.T) {
		var (
			source = pipeline.From(pipeline.FromSlice(1, 2, 3, 4))
			action = pipeline.Map(func(i int) int { return i * 2 })
			sink   = pipeline.ToSlice[int]()
		)

		source.Thru(action).To(sink)

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
			pipe1, pipe2 = pipeline.From(pipeline.FromSlice(1, 2, 3, 4)).Tee(
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

	t.Run("cancel tee", func(t *testing.T) {
		var (
			sink1        = pipeline.ToSlice[int]()
			sink2        = pipeline.ToSlice[int]()
			ctx, cancel  = context.WithCancel(context.Background())
			pipe1, pipe2 = pipeline.
					FromSlice(1, 2, 3, 4).
					WithContext(ctx).
					Tee(
					pipeline.Passthrough(),
					pipeline.Passthrough(),
				)
		)

		cancel()
		pipe1.To(sink1)
		pipe2.To(sink2)

		var (
			got1 = sink1.Slice()
			got2 = sink2.Slice()
		)

		if len(got1) == 4 {
			t.Errorf("got %v", got1)
		}
		if len(got2) == 4 {
			t.Errorf("got %v", got2)
		}
	})
}

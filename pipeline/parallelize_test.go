package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

func TestParallelize(t *testing.T) {
	t.Parallel()

	t.Run("should run in parallel", func(t *testing.T) {
		var (
			source = pipeline.FromSlice(1, 2, 3, 4)
			pipe   = pipeline.Parallelize(2,
				func() piper.Pipe {
					return pipeline.Map(func(i int) int { return i * 2 })
				},
			)
			got = Consume[int](source.Thru(pipe))
		)

		if len(got) != 4 {
			t.Fatalf("expected 4 items, got %d", len(got))
		}

		for _, i := range got {
			if i%2 != 0 {
				t.Fatalf("expected even numbers, got %d", i)
			}
		}
	})

	t.Run("no parallization", func(t *testing.T) {
		var (
			source = pipeline.FromSlice(1, 2, 3, 4)
			pipe   = pipeline.Parallelize(1,
				func() piper.Pipe {
					return pipeline.Map(func(i int) int { return i * 2 })
				},
			)
			want = []int{2, 4, 6, 8}
			got  = Consume[int](source.Thru(pipe))
		)

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("expected %v, got %v", want, got)
		}

		for _, i := range got {
			if i%2 != 0 {
				t.Fatalf("expected even numbers, got %d", i)
			}
		}
	})

	t.Run("panics if parallelize size is non positive integer", func(t *testing.T) {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic")
				}
			}()

			pipeline.Parallelize(0, func() piper.Pipe { return nil })
		}()
	})
}

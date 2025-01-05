package pipeline_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nisimpson/piper/pipeline"
)

func TestBatchN(t *testing.T) {
	t.Parallel()

	source := pipeline.
		FromSlice(1, 2, 3, 4).
		Then(pipeline.BatchN[int](2))

	var (
		want = [][]int{{1, 2}, {3, 4}}
		got  = Consume[[]int](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestBatchEvery(t *testing.T) {
	t.Parallel()

	t.Run("flushes during each duration", func(t *testing.T) {
		var (
			in     = make(chan int)
			source = pipeline.
				FromChannel(in).
				Then(pipeline.BatchEvery[int](time.Millisecond))
		)

		go func() {
			defer close(in)
			in <- 1
			in <- 2
			time.Sleep(2 * time.Millisecond)
			in <- 3
			in <- 4
		}()

		var (
			want = [][]int{{1, 2}, {3, 4}}
			got  = Consume[[]int](source)
		)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
		}
	})

	t.Run("flushes when the input channel closes", func(t *testing.T) {
		// the source will close before the batch timer is triggered.
		source := pipeline.
			FromSlice(1, 2, 3, 4).
			Then(pipeline.BatchEvery[int](1 * time.Hour))

		var (
			want = [][]int{{1, 2, 3, 4}}
			got  = Consume[[]int](source)
		)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
		}
	})
}

func TestBatch(t *testing.T) {
	t.Parallel()

	t.Run("flushes if maximum size is reached", func(t *testing.T) {
		var (
			in     = make(chan int)
			source = pipeline.
				FromChannel(in).
				Then(pipeline.Batch[int](func(bo *pipeline.BatcherOptions) {
					bo.Interval = time.Millisecond
					bo.MaxSize = 3
				}))
		)

		go func() {
			defer close(in)
			for i := 1; i < 6; i++ {
				in <- i
				time.Sleep(10 * time.Nanosecond)
			}
		}()

		var (
			want = [][]int{{1, 2, 3}, {4, 5}}
			got  = Consume[[]int](source)
		)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
		}
	})
}

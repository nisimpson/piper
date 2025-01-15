package pipeline_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nisimpson/piper/pipeline"
)

func TestSlidingWindow(t *testing.T) {
	t.Parallel()

	t.Run("basic usage", func(t *testing.T) {
		// Create a sliding window with size 3 and step 1
		pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
			o.WindowSize = 3
			o.StepSize = 1
		})

		// Send input
		in := pipe.In()
		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		var (
			want = [][]int{
				{1, 2, 3},
				{2, 3, 4},
				{3, 4, 5},
			}
			got = make([][]int, 0)
		)

		// Receive windows
		for window := range pipe.Out() {
			got = append(got, window.([]int))
		}

		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}
	})

	t.Run("with step", func(t *testing.T) {
		// Create a sliding window with size 3 and step 2
		pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
			o.WindowSize = 3
			o.StepSize = 2
		})

		// Send input
		in := pipe.In()
		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		var (
			want = [][]int{
				{1, 2, 3},
				{3, 4, 5},
			}
			got = make([][]int, 0)
		)

		// Receive windows
		for window := range pipe.Out() {
			got = append(got, window.([]int))
		}

		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}
	})

	t.Run("with interval", func(t *testing.T) {
		// Create a sliding window with size 3 and step 1
		pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
			o.WindowSize = 3
			o.StepSize = 1
			o.Interval = 20 * time.Millisecond
		})

		// Send input
		in := pipe.In()
		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(30 * time.Millisecond)
			}
			close(in)
		}()

		var (
			want = [][]int{
				{1, 2, 3},
				{2, 3, 4},
				{3, 4, 5},
			}
			got = make([][]int, 0)
		)

		// Receive windows
		for window := range pipe.Out() {
			got = append(got, window.([]int))
		}

		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
			return
		}
	})
}

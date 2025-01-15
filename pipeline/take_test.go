package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestTakeN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		count    int
		input    []any
		expected []any
	}{
		{
			name:     "take three items",
			count:    3,
			input:    []any{1, 2, 3, 4, 5},
			expected: []any{1, 2, 3},
		},
		{
			name:     "take zero items",
			count:    0,
			input:    []any{1, 2, 3},
			expected: []any{},
		},
		{
			name:     "take with negative count",
			count:    -1,
			input:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "take more items than available",
			count:    5,
			input:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "take strings",
			count:    2,
			input:    []any{"apple", "banana", "cherry", "date"},
			expected: []any{"apple", "banana"},
		},
		{
			name:     "take from empty input",
			count:    3,
			input:    []any{},
			expected: []any{},
		},
		{
			name:     "take mixed types",
			count:    4,
			input:    []any{1, "two", 3.0, true, "five"},
			expected: []any{1, "two", 3.0, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the pipeline
			pipe := pipeline.TakeN(tt.count)

			// Create channel for collecting results
			results := make([]any, 0)
			done := make(chan struct{})

			// Start collector goroutine
			go func() {
				for item := range pipe.Out() {
					results = append(results, item)
				}
				close(done)
			}()

			// Send input
			in := pipe.In()
			for _, item := range tt.input {
				in <- item
			}
			close(in)

			// Wait for collection to complete
			<-done

			// Check results
			if !reflect.DeepEqual(results, tt.expected) {
				t.Errorf("got %v, want %v", results, tt.expected)
			}
		})
	}

	t.Run("chained operations", func(t *testing.T) {
		// Create a pipeline that doubles numbers and takes first 3
		input := []any{1, 2, 3, 4, 5}
		expected := []any{2, 4, 6}

		source := pipeline.FromSlice(input...)
		result := source.
			Thru(pipeline.Map(func(x any) any {
				return x.(int) * 2
			})).
			Thru(pipeline.TakeN(3))

		// Collect results
		results := make([]any, 0)
		for val := range result.Out() {
			results = append(results, val)
		}

		// Verify results
		if !reflect.DeepEqual(results, expected) {
			t.Errorf("got %v, want %v", results, expected)
		}
	})

	t.Run("take last", func(t *testing.T) {
		// Create a pipeline that takes the last 3 items
		input := []any{1, 2, 3, 4, 5}
		expected := []any{3, 4, 5}

		source := pipeline.FromSlice(input...)
		result := source.Thru(pipeline.TakeLastN(3))

		// Collect results
		results := make([]any, 0)
		for val := range result.Out() {
			results = append(results, val)
		}

		// Verify results
		if !reflect.DeepEqual(results, expected) {
			t.Errorf("got %v, want %v", results, expected)
		}
	})
}

func BenchmarkTakeN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipe := pipeline.TakeN(5)
		go func() {
			in := pipe.In()
			for j := 0; j < 10; j++ {
				in <- j
			}
			close(in)
		}()

		// Consume all output
		for range pipe.Out() {
		}
	}
}

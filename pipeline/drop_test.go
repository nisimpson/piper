package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestDropN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		count    int
		input    []any
		expected []any
	}{
		{
			name:     "drop three items",
			count:    3,
			input:    []any{1, 2, 3, 4, 5},
			expected: []any{4, 5},
		},
		{
			name:     "drop zero items",
			count:    0,
			input:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "drop with negative count",
			count:    -1,
			input:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "drop more items than available",
			count:    5,
			input:    []any{1, 2, 3},
			expected: []any{},
		},
		{
			name:     "drop strings",
			count:    2,
			input:    []any{"apple", "banana", "cherry", "date"},
			expected: []any{"cherry", "date"},
		},
		{
			name:     "drop from empty input",
			count:    3,
			input:    []any{},
			expected: []any{},
		},
		{
			name:     "drop mixed types",
			count:    2,
			input:    []any{1, "two", 3.0, true, "five"},
			expected: []any{3.0, true, "five"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the pipeline
			pipe := pipeline.DropN(tt.count)

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
		// Create a pipeline that doubles numbers and drops first 2
		input := []any{1, 2, 3, 4, 5}
		expected := []any{6, 8, 10}

		source := pipeline.FromSlice(input...)
		result := source.
			Thru(pipeline.Map(func(x any) any {
				return x.(int) * 2
			})).
			Thru(pipeline.DropN(2))

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

func BenchmarkDropN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipe := pipeline.DropN(5)
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

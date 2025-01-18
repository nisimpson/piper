package pipeline_test

import (
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestChunk(t *testing.T) {
	t.Parallel()

	t.Run("panics if size is less than 1", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()

		pipeline.Chunk[[]any](0)
	})
}

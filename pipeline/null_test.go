package pipeline_test

import (
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestNull(t *testing.T) {
	t.Parallel()

	pipe := pipeline.FromSlice(1, 2, 3)
	null := pipeline.ToNull()
	pipe.To(null)
	null.Wait()
}

package pipeline_test

import (
	"piper/pipeline"
	"testing"
)

func TestNull(t *testing.T) {
	pipe := pipeline.FromSlice(1, 2, 3)
	null := pipeline.ToNull()
	pipe.To(null)
	null.Wait()
}

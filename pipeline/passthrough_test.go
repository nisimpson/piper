package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestPassthrough(t *testing.T) {
	t.Parallel()

	var (
		source = pipeline.FromSlice(1, 2, 3, 4)
		action = pipeline.Passthrough()
	)

	source = source.Thru(action)

	var (
		want = []int{1, 2, 3, 4}
		got  = Consume[int](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
}

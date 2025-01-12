package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestJoin(t *testing.T) {
	var (
		source = pipeline.FromSlice(1, 2, 3, 4)
		pipe1  = pipeline.Map(func(i int) int { return i * 2 })
		pipe2  = pipeline.Map(func(i int) int { return i * 3 })
		pipe   = pipeline.Join(pipe1, pipe2)
		want   = []int{6, 12, 18, 24}
		got    = Consume[int](source.Thru(pipe))
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
}

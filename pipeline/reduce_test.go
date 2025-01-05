package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	var (
		adder  = func(acc, cur int) int { return acc + cur }
		source = pipeline.FromSlice(1, 2, 3, 4)
		action = pipeline.Reduce(adder)
	)

	source = source.Then(action)

	var (
		want = []int{1, 3, 6, 10} // [(1), (1 + 2), (3 + 3), (6 + 4)]
		got  = Consume[int](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

package pipeline_test

import (
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestFlatMap(t *testing.T) {
	t.Parallel()

	type node struct {
		children []node
	}

	getChildren := func(n node) []node { return n.children }

	source := pipeline.FromSlice(
		node{children: []node{{}, {}}},
		node{children: []node{{}, {}, {}}},
	).Thru(pipeline.FlatMap(getChildren))

	want := []node{{}, {}, {}, {}, {}}
	got := Consume[node](source)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestFlatten(t *testing.T) {
	t.Parallel()

	source := pipeline.FromSlice(
		[]int{1, 2, 3},
		[]int{4, 5, 6},
	).Thru(pipeline.Flatten[[]int]())

	want := []int{1, 2, 3, 4, 5, 6}
	got := Consume[int](source)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

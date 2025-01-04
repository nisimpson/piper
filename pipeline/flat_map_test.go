package pipeline_test

import (
	"piper/pipeline"
	"reflect"
	"testing"
)

func TestFlatMap(t *testing.T) {
	type node struct {
		children []node
	}

	getChildren := func(n node) []node { return n.children }

	source := pipeline.FromSlice(
		node{children: []node{{}, {}}},
		node{children: []node{{}, {}, {}}},
	).Then(pipeline.FlatMap(getChildren))

	want := []node{{}, {}, {}, {}, {}}
	got := Consume[node](source)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

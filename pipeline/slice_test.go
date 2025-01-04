package pipeline_test

import (
	"piper/pipeline"
	"reflect"
	"testing"
)

func TestFromSlice(t *testing.T) {
	t.Parallel()

	want := []int{1, 2, 3, 4}
	got := Consume[int](pipeline.FromSlice(want...))
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	want := []int{1, 2, 3, 4}
	source := pipeline.FromSlice(want...)
	sink := pipeline.ToSlice[int]()
	source.To(sink)
	got := sink.Slice()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

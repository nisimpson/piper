package pipeline_test

import (
	"piper/pipeline"
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	isEven := func(i int) bool { return i%2 == 0 }
	source := pipeline.FromSlice(1, 2, 3, 4)
	source = source.Then(pipeline.Filter(isEven))
	want := []int{2, 4}
	got := Consume[int](source)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestKeepIf(t *testing.T) {
	t.Parallel()

	isEven := func(i int) bool { return i%2 == 0 }
	source := pipeline.FromSlice(1, 2, 3, 4)
	source = source.Then(pipeline.KeepIf(isEven))
	want := []int{2, 4}
	got := Consume[int](source)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestDropIf(t *testing.T) {
	t.Parallel()

	isEven := func(i int) bool { return i%2 == 0 }
	source := pipeline.FromSlice(1, 2, 3, 4)
	source = source.Then(pipeline.DropIf(isEven))
	want := []int{1, 3}
	got := Consume[int](source)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

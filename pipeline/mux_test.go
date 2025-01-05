package pipeline_test

import (
	"piper/pipeline"
	"reflect"
	"slices"
	"testing"
)

func TestFromMux(t *testing.T) {
	t.Parallel()

	var (
		s1     = pipeline.FromSlice(1, 2, 3, 4)
		s2     = pipeline.FromSlice(5, 6, 7, 8)
		source = pipeline.Mux(s1, s2)
		got    = Consume[int](source)
	)

	if len(got) != 8 {
		t.Errorf("wanted %d items, got %d", 8, len(got))
		return
	}

	// actual result is likely out of order so let's sort
	// to verify them.
	slices.Sort(got)

	want := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

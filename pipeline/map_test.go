package pipeline_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
)

func TestMap(t *testing.T) {
	t.Parallel()

	source := pipeline.FromSlice(1, 2, 3, 4)
	double := func(in int) string { return fmt.Sprintf("%d", in*2) }
	source = source.Thru(pipeline.Map(double))
	want := []string{"2", "4", "6", "8"}
	got := Consume[string](source)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

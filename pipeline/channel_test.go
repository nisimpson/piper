package pipeline_test

import (
	"piper/pipeline"
	"reflect"
	"testing"
)

func TestFromChannel(t *testing.T) {
	ch := make(chan int, 4)
	ch <- 1
	ch <- 2
	ch <- 3
	ch <- 4
	close(ch)

	source := pipeline.FromChannel(ch)
	want := []int{1, 2, 3, 4}
	got := Consume[int](source)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

func TestToChannel(t *testing.T) {
	var (
		ch     = make(chan int)
		source = pipeline.FromSlice(1, 2, 3, 4)
		sink   = pipeline.ToChannel(ch)
	)

	source.To(sink)

	var (
		want = []int{1, 2, 3, 4}
		got  = make([]int, 0)
	)

	for item := range ch {
		got = append(got, item)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
	}
}

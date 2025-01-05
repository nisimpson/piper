package pipeline_test

import (
	"errors"
	"piper/pipeline"
	"reflect"
	"testing"
)

func EchoCommand(input string, err error) pipeline.Command[string, string] {
	return pipeline.CommandFunc(func(in string) (string, int, error) {
		if in != "" {
			return in, 0, err
		}
		return input, 0, err
	})
}

func TestFromCmd(t *testing.T) {
	t.Parallel()

	var (
		source = pipeline.FromCmd(EchoCommand("hello", nil))
		want   = []string{"hello"}
		got    = Consume[string](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
		return
	}
}

func TestExecCmd(t *testing.T) {
	t.Parallel()

	var (
		source = pipeline.FromSlice("hello")
		action = pipeline.ExecCmd(EchoCommand("", nil))
		want   = []string{"hello"}
		got    = Consume[string](source.Then(action))
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
		return
	}

	t.Run("handles command error", func(t *testing.T) {
		var (
			handled = []bool{false}
			source  = pipeline.FromSlice("hello")
			action  = pipeline.ExecCmd(EchoCommand("", errors.New("an error")),
				func(cpo *pipeline.CommandPipeOptions[string]) {
					cpo.HandleError = func(err error) { handled[0] = true }
				},
			)
			want = []string{}
			got  = Consume[string](source.Then(action))
		)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
			return
		}

		if !handled[0] {
			t.Errorf("error was not handled")
		}
	})
}

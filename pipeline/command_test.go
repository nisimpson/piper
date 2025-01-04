package pipeline_test

import (
	"errors"
	"piper/pipeline"
	"reflect"
	"testing"
)

type EchoCommand struct {
	input string
	err   error
}

func (e EchoCommand) Execute(input string) (out string, exitcode int, err error) {
	if e.input != "" {
		return e.input, 0, e.err
	}
	return input, 0, e.err
}

func TestFromCmd(t *testing.T) {
	var (
		source = pipeline.FromCmd(EchoCommand{input: "hello"})
		want   = []string{"hello"}
		got    = Consume[string](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
		return
	}
}

func TestExecCmd(t *testing.T) {
	var (
		source = pipeline.FromSlice("hello")
		action = pipeline.ExecCmd(EchoCommand{})
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
			action  = pipeline.ExecCmd(EchoCommand{err: errors.New("an error")},
				func(cpo *pipeline.CommandPipeOptions) {
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

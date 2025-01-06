package command_test

import (
	"os/exec"
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
	"github.com/nisimpson/piper/pipeline/command"
)

func TestFromShell(t *testing.T) {
	t.Parallel()

	t.Run("single command", func(t *testing.T) {
		var (
			commands = command.FromShell(
				exec.Command("echo", "hello world"),
			)
			source = pipeline.FromCmd(commands)
			sink   = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			want = []string{"hello world\n"}
			got  = sink.Slice()
		)

		if len(want) != len(got) {
			t.Fatalf("want %d, got %d", len(want), len(got))
			return
		}

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("chain multiple calls", func(t *testing.T) {
		var (
			commands = command.FromShell(
				exec.Command("echo", "hello world"), // echo
				exec.Command("wc", "-l"),            // word count: number of lines
			)
			source = pipeline.FromCmd(commands)
			sink   = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			want = []string{"1\n"}
			got  = sink.Slice()
		)

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

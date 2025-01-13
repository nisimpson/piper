package command_test

import (
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/nisimpson/piper/pipeline"
	"github.com/nisimpson/piper/pipeline/command"
)

var (
	skipShellTests = false
)

func TestMain(m *testing.M) {
	skipShellTests = os.Getenv("SKIP_SHELL_TESTS") != ""
	m.Run()
}

func TestShell(t *testing.T) {
	if skipShellTests {
		t.Skip("skipping shell tests")
	}

	t.Parallel()

	t.Run("nothing to do", func(t *testing.T) {
		var (
			commands = command.Shell()
			source   = pipeline.FromCmd(commands)
			sink     = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			got = sink.Slice()
		)

		if len(got) != 0 {
			t.Fatalf("want %d, got %d", 0, len(got))
		}
	})

	t.Run("single command", func(t *testing.T) {
		var (
			commands = command.Shell(
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

	t.Run("command fails with exit code", func(t *testing.T) {
		var (
			commands = command.Shell(
				exec.Command("false"),
			)
			source = pipeline.FromCmd(commands)
			sink   = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			got = sink.Slice()
		)

		if len(got) != 0 {
			t.Fatalf("want %d, got %d", 0, len(got))
		}
	})

	t.Run("chain commands fail with exit code", func(t *testing.T) {
		var (
			commands = command.Shell(
				exec.Command("echo", "hello world"),
				exec.Command("wc", "-z"),
			)
			source = pipeline.FromCmd(commands)
			sink   = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			got = sink.Slice()
		)

		if len(got) != 0 {
			t.Fatalf("want %d, got %d", 0, len(got))
		}
	})

	t.Run("chain multiple calls", func(t *testing.T) {
		var (
			commands = command.Shell(
				exec.Command("echo", "hello world"),   // echo
				exec.Command("wc", "-c"),              // word count: number of characters
				exec.Command("tr", "-d", "[:space:]"), // tr: translate/delete: delete spaces
			)
			source = pipeline.FromCmd(commands)
			sink   = pipeline.ToSlice[string]()
		)

		source.To(sink)

		var (
			want = []string{"12"} // "hello" + " " + "world" + "\n"
			got  = sink.Slice()
		)

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestShellCommandsAsPipe(t *testing.T) {
	if skipShellTests {
		t.Skip("skipping shell tests")
	}

	t.Parallel()

	t.Run("single command", func(t *testing.T) {
		var (
			source = pipeline.FromSlice("hello world")
			pipe   = pipeline.ExecCmd(command.Shell(exec.Command("echo")))
			sink   = pipeline.ToSlice[string]()
		)

		source.Thru(pipe).To(sink)

		var (
			want = []string{"hello world\n"}
			got  = sink.Slice()
		)

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("multiple commands", func(t *testing.T) {
		var (
			source = pipeline.FromSlice("hello world")
			pipe   = pipeline.ExecCmd(command.Shell(
				exec.Command("echo"),
				exec.Command("wc", "-c"),
				exec.Command("tr", "-d", "[:space:]"),
			))
			sink = pipeline.ToSlice[string]()
		)

		source.Thru(pipe).To(sink)

		var (
			want = []string{"12"}
			got  = sink.Slice()
		)

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

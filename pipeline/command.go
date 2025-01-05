package pipeline

import (
	"piper"
	"piper/internal/must"
)

// CommandPipeOptions configure how command execution errors and output are handled in the pipeline.
type CommandPipeOptions[Out any] struct {
	// HandleError is called when a command execution results in an error.
	HandleError func(error)
	// HandleOutput processes command output before sending it downstream.
	// It receives the command output string and exit code, and returns a modified output string.
	HandleOutput func(out Out, exitcode int) Out
}

// Command represents an executable operation that can be run in a [piper.Pipeline].
// The generic type In represents the type of input the command accepts.
type Command[In any, Out any] interface {
	// Execute runs the command with the given input and returns its output, exit code, and any error.
	Execute(input In) (out Out, exitcode int, err error)
}

// CommandFunction is a function that implements [Command].
type CommandFunction[In any, Out any] func(input In) (out Out, exitcode int, err error)

// CommandFunc wraps a function into a [CommandFunction] for use by [FromCmd] or [ExecCmd].
func CommandFunc[In any, Out any](f CommandFunction[In, Out]) Command[In, Out] {
	return f
}

// Execute implements [Command], invoking the underlying function.
func (f CommandFunction[In, Out]) Execute(input In) (out Out, exitcode int, err error) {
	return f(input)
}

// executor implements a pipeline component that executes commands.
// It can be configured to handle errors and process command output in custom ways.
type executor[In any, Out any] struct {
	// cmd is the command to be executed
	cmd Command[In, Out]
	// in receives inputs to be passed to the command
	in chan any
	// out sends processed command outputs
	out chan any
	// options configure error handling and output processing
	options []func(*CommandPipeOptions[Out])
}

// FromCmd creates a new [piper.Pipeline] that starts with command execution.
// The command will be executed once with an empty input, making it suitable for commands
// that don't require input (like 'ls' or 'date').
func FromCmd[In any, Out any](cmd Command[In, Out], opts ...func(*CommandPipeOptions[Out])) piper.Pipeline {
	source := executor[In, Out]{
		cmd:     cmd,
		in:      make(chan any, 1),
		out:     make(chan any),
		options: opts,
	}

	source.in <- ""
	close(source.in)

	go source.start()
	return piper.PipelineFrom(source)
}

// ExecCmd creates a [piper.Pipe] component that executes a command for each input it receives.
// This is suitable for commands that process input (like 'grep' or 'sed').
func ExecCmd[In any, Out any](cmd Command[In, Out], opts ...func(*CommandPipeOptions[Out])) piper.Pipe {
	source := executor[In, Out]{
		cmd:     cmd,
		in:      make(chan any),
		out:     make(chan any),
		options: opts,
	}

	go source.start()
	return source
}

func (c executor[In, Out]) In() chan<- any {
	return c.in
}

func (c executor[In, Out]) Out() <-chan any {
	return c.out
}

// start begins the command execution process.
// It processes each input by executing the command and handling its output according to the configured options.
func (c executor[In, Out]) start() {
	defer close(c.out)

	opts := CommandPipeOptions[Out]{
		HandleError:  must.IgnoreError,
		HandleOutput: c.passCommandOutput,
	}

	for _, opt := range c.options {
		opt(&opts)
	}

	for input := range c.in {
		// execute command
		output, exitcode, err := c.cmd.Execute(input.(In))

		// handle error
		if err != nil {
			opts.HandleError(err)
			continue
		}

		// handle output
		c.out <- opts.HandleOutput(output, exitcode)
	}
}

// passCommandOutput is the default output handler that simply passes through the command's output string.
// It ignores the exit code and returns the output unchanged.
func (executor[In, Out]) passCommandOutput(out Out, _ int) Out {
	return out
}

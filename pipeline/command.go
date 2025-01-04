package pipeline

import (
	"piper"
	"piper/internal/must"
)

type CommandPipeOptions struct {
	HandleError  func(error)
	HandleOutput func(out string, exitcode int) string
}

type Command[In any] interface {
	Execute(input In) (out string, exitcode int, err error)
}

type commandPipe[In any] struct {
	cmd     Command[In]
	in      chan any
	out     chan any
	options []func(*CommandPipeOptions)
}

func FromCmd[In any](cmd Command[In], opts ...func(*CommandPipeOptions)) piper.Pipeline {
	source := commandPipe[In]{
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

func ExecCmd[In any](cmd Command[In], opts ...func(*CommandPipeOptions)) piper.Pipe {
	source := commandPipe[In]{
		cmd:     cmd,
		in:      make(chan any),
		out:     make(chan any),
		options: opts,
	}

	go source.start()
	return source
}

func (c commandPipe[In]) In() chan<- any {
	return c.in
}

func (c commandPipe[In]) Out() <-chan any {
	return c.out
}

func (c commandPipe[In]) start() {
	defer close(c.out)

	opts := CommandPipeOptions{
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

func (commandPipe[In]) passCommandOutput(out string, _ int) string {
	return out
}

package command

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/nisimpson/piper/pipeline"
)

type shellCommands []*exec.Cmd

func Shell(cmds ...*exec.Cmd) pipeline.Command[string, string] {
	return shellCommands(cmds)
}

func (s shellCommands) Execute(input string) (out string, exitcode int, err error) {
	var (
		errbuf bytes.Buffer
		outbuf bytes.Buffer
		pipes  = make([]*io.PipeWriter, len(s)-1)
		i      = 0
	)

	for ; i < len(s)-1; i++ {
		stdin, stdout := io.Pipe()
		s[i].Stdout = stdout
		s[i].Stderr = &errbuf
		s[i+1].Stdin = stdin
		pipes[i] = stdout
	}

	s[i].Stdin = bytes.NewBufferString(input)
	s[i].Stdout = &outbuf
	s[i].Stderr = &errbuf

	err = s.call(s, pipes)
	out = outbuf.String()

	if werr, ok := err.(*exec.ExitError); ok {
		exitcode = werr.ExitCode()
	}

	return out, exitcode, err
}

func (s shellCommands) call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = s.call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

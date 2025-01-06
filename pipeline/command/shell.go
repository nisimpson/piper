package command

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"slices"

	"github.com/nisimpson/piper/internal/must"
	"github.com/nisimpson/piper/pipeline"
)

// shellCommands represent a sequence of shell commands that can be executed in a pipeline.
// Each command's output is connected to the input of the next command in the sequence.
type shellCommands []*exec.Cmd

// FromShell creates a new [pipeline.Command] that executes shell commands in sequence.
// The commands are executed in the order they are provided, with each command's output
// piped to the next command's input.
func FromShell(cmds ...*exec.Cmd) pipeline.Command[string, string] {
	return shellCommands(cmds)
}

func (s shellCommands) Execute(input string) (out string, exitcode int, err error) {
	if len(s) == 0 {
		// nothing to do
		return "", 0, nil
	}

	var (
		// errbuf collects error output from all commands
		errbuf bytes.Buffer
		// capture all error codes generated
		exitcodes = make([]int, len(s))
		// capture all errors generated
		errs = make([]error, len(s))
	)

	cur := s[0]
	cur.Stdin = bytes.NewBufferString(input) // the first command's input is the input string
	cur.Stderr = &errbuf                     // all commands share the same error buffer

	for i := 1; i < len(s); i++ {
		stdout := must.Return(cur.StdoutPipe())

		// execute the next command
		next := s[i]
		next.Stdin = stdout
		next.Stderr = &errbuf

		// update pointer to next command
		cur = next
	}

	output :=

		// reverse the slice of commands to execute them in the correct order
		slices.Reverse(s)
	for _, cmd := range s {

		var werr *exec.ExitError
		if err := cmd.Wait(); errors.As(err, &werr) {
			// Extract exit code if command failed
			exitcodes = append(exitcodes, werr.ExitCode())
			errs = append(errs, err)
		} else if err != nil {
			errs = append(errs, err)
		}
	}

	out = outbuf.String()
	err = errors.Join(errs...)
	if len(exitcodes) > 0 {
		exitcode = exitcodes[0]
	}
	return
}

// Execute runs the chain of shell commands with the given input string.
// It connects the commands using pipes and executes them in sequence.
// Returns:
//   - out: the final command's standard output as a string
//   - exitcode: the exit code if any command failed (0 otherwise)
//   - err: any error that occurred during execution
func (s shellCommands) Execute2(input string) (out string, exitcode int, err error) {
	var (
		// errbuf collects error output from all commands
		errbuf bytes.Buffer
		// outbuf stores the final command's output
		outbuf bytes.Buffer
		// pipes holds the pipe writers connecting each command
		pipes = make([]*io.PipeWriter, len(s)-1)
		// i is used as an index for setting up command pipes
		i = 0
	)

	// Set up pipes between sequential commands
	for ; i < len(s)-1; i++ {
		stdin, stdout := io.Pipe()
		s[i].Stdout = stdout  // Connect command output to pipe
		s[i].Stderr = &errbuf // Collect stderr output
		s[i+1].Stdin = stdin  // Connect next command's input to pipe
		pipes[i] = stdout     // Store pipe for later cleanup
	}

	// Configure the last command's I/O
	s[i].Stdin = bytes.NewBufferString(input) // Feed input string to first command
	s[i].Stdout = &outbuf                     // Collect final output
	s[i].Stderr = &errbuf                     // Collect stderr output

	// Execute the command chain
	err = s.call(s, pipes)
	out = outbuf.String()

	// Extract exit code if command failed
	if werr, ok := err.(*exec.ExitError); ok {
		exitcode = werr.ExitCode()
	}

	return out, exitcode, err
}

// call recursively executes a chain of commands connected by pipes.
// It ensures proper sequential execution of commands where each command's output
// feeds into the next command's input through the provided pipes.
func (s shellCommands) call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	// Start the first command if it hasn't been started yet
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}

	// If there are more commands in the chain, start the next one
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		// After the current command finishes, close its pipe and continue the chain
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = s.call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

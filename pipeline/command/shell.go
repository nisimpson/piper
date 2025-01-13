package command

import (
	"bytes"
	"errors"
	"io"
	"os/exec"

	"github.com/nisimpson/piper/internal/must"
	"github.com/nisimpson/piper/pipeline"
)

// shellCommands represent a sequence of shell commands that can be executed in a pipeline.
// Each command's output is connected to the input of the next command in the sequence.
type shellCommands []*exec.Cmd

// Shell creates a new [pipeline.Command] that executes shell commands in sequence.
// The commands are executed in the order they are provided, with each command's output
// piped to the next command's input.
//
// If Shell is used as a pipe (via [pipeline.ExecCmd]), any nonempty upstream input value
// is used as the last positional argument to the first command in the sequence.
func Shell(cmds ...*exec.Cmd) pipeline.Command[string, string] {
	return shellCommands(cmds)
}

// Execute runs the chain of shell commands with the given input string.
// It connects the commands using pipes and executes them in sequence.
// Returns:
//   - out: the final command's standard output as a string
//   - exitcode: the exit code if any command failed (0 otherwise)
//   - err: any error that occurred during execution
func (s shellCommands) Execute(input string) (out string, exitcode int, err error) {
	if len(s) == 1 {
		return s.executeOne(input)
	} else if len(s) > 1 {
		return s.executeAll(input)
	}
	// nothing to do.
	return "", 0, io.EOF
}

// executeAll executes all commands in the sequence, connecting their outputs and inputs.
// It collects the output of the last command and any errors that occurred during execution.
// Returns:
//   - out: the final command's standard output as a string
//   - exitcode: the exit code if any command failed (0 otherwise)
//   - err: any error that occurred during execution
func (s shellCommands) executeAll(input string) (out string, exitcode int, err error) {
	var (
		errbuf    bytes.Buffer            // errbuf collects error output from all commands
		outbuf    bytes.Buffer            // outbuf collects output from last command
		exitcodes = make([]int, len(s))   // capture all error codes generated
		errs      = make([]error, len(s)) // capture all errors generated
	)

	cur := s[0]          // first command in the pipe
	cur.Stderr = &errbuf // all commands share the same error buffer

	// set the first command's input
	if input != "" {
		cur.Args = append(cur.Args, input)
	}

	// for each command, set the next command's stdin to the current command's stdout
	for i := 1; i < len(s); i++ {
		// set up stdout routing between commands
		stdout := must.Return(cur.StdoutPipe())
		next := s[i]
		next.Stdin = stdout
		next.Stderr = &errbuf

		// update pointer to next command
		cur = next
	}

	// last command's output
	var (
		stdout = must.Return(cur.StdoutPipe())
		done   = make(chan struct{})
	)

	go func() {
		must.Return(io.Copy(&outbuf, stdout))
		done <- struct{}{}
	}()

	// execute commands in reverse
	for i := len(s) - 1; i >= 0; i-- {
		must.PanicOnError(s[i].Start())
	}

	<-done

	// wait for all of the commands to finish
	for _, cmd := range s {
		var werr *exec.ExitError
		cmderr := cmd.Wait()
		errs = append(errs, cmderr)
		if errors.As(cmderr, &werr) {
			// capture error code if available
			exitcodes = append(exitcodes, werr.ExitCode())
		}
	}

	out = outbuf.String()
	err = errors.Join(errs...)
	if len(exitcodes) > 0 {
		exitcode = exitcodes[0]
	}
	return
}

// executeOne executes the first and only command in the slice.
// It collects the output of the last command and any errors that occurred during execution.
// Returns:
//   - out: the final command's standard output as a string
//   - exitcode: the exit code if any command failed (0 otherwise)
//   - err: any error that occurred during execution
func (s shellCommands) executeOne(input string) (out string, exitcode int, err error) {
	cmd := s[0]
	if input != "" {
		cmd.Args = append(cmd.Args, input) // the first command's argument is the input string
	}
	output, err := cmd.Output()
	var werr *exec.ExitError
	if errors.As(err, &werr) {
		// Extract exit code if command failed
		exitcode = werr.ExitCode()
	}
	return string(output), exitcode, err
}

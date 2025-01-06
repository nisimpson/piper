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
		// errbuf collects error output from all commands
		errbuf bytes.Buffer
		// outbuf collects output from last command
		outbuf bytes.Buffer
		// capture all error codes generated
		exitcodes = make([]int, len(s))
		// capture all errors generated
		errs = make([]error, len(s))
	)

	cur := s[0]
	cur.Stdin = bytes.NewBufferString(input) // the first command's input is the input string
	cur.Stderr = &errbuf                     // all commands share the same error buffer

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
	cur.Stdout = &outbuf

	// reverse the slice of commands to execute them in the correct order
	reversed := make(shellCommands, 0, len(s))
	reversed = append(reversed, s...)
	slices.Reverse(reversed)

	// start commands
	for _, cmd := range reversed {
		must.PanicOnError(cmd.Start())
	}

	// wait for all of the commands to finish
	for _, cmd := range s {
		var werr *exec.ExitError
		if err := cmd.Wait(); errors.As(err, &werr) {
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

// executeOne executes the first and only command in the slice.
// It collects the output of the last command and any errors that occurred during execution.
// Returns:
//   - out: the final command's standard output as a string
//   - exitcode: the exit code if any command failed (0 otherwise)
//   - err: any error that occurred during execution
func (s shellCommands) executeOne(input string) (out string, exitcode int, err error) {
	cmd := s[0]
	cmd.Stdin = bytes.NewBufferString(input) // the first command's input is the input string
	output, err := cmd.Output()
	var werr *exec.ExitError
	if errors.As(err, &werr) {
		// Extract exit code if command failed
		exitcode = werr.ExitCode()
	}
	return string(output), exitcode, err
}

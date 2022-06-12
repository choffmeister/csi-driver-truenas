package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExecOpts struct {
	Name       string
	Args       []string
	Env        []string
	Input      string
	Output     io.Writer
	Retries    int
	RetryDelay time.Duration
}

func Exec(name string, args ...string) (string, int, error) {
	return ExecWithOpts(ExecOpts{
		Name: name,
		Args: args,
	})
}

func ExecRetry(name string, args ...string) (string, int, error) {
	return ExecWithOpts(ExecOpts{
		Name:       name,
		Args:       args,
		Retries:    300 / 5,
		RetryDelay: 5 * time.Second,
	})
}

func ExecWithOutput(name string, args ...string) (string, int, error) {
	return ExecWithOpts(ExecOpts{
		Name:   name,
		Args:   args,
		Output: os.Stdout,
	})
}

func ExecRetryWithOutput(name string, args ...string) (string, int, error) {
	return ExecWithOpts(ExecOpts{
		Name:       name,
		Args:       args,
		Output:     os.Stdout,
		Retries:    300 / 5,
		RetryDelay: 5 * time.Second,
	})
}

func ExecWithOpts(opts ExecOpts) (string, int, error) {
	fmt.Printf("Executing command %s %v\n", opts.Name, opts.Args)

	lastOutput := ""
	lastCode := 0
	lastErr := (error)(nil)
	attempt := 0
	for attempt <= opts.Retries {
		cmd := exec.Command(opts.Name, opts.Args...)
		if opts.Input != "" {
			cmd.Stdin = strings.NewReader(opts.Input)
		}
		var outputBuffer bytes.Buffer
		writer := ioutil.Discard
		if opts.Output != nil {
			writer = opts.Output
		}
		writer = io.MultiWriter(&outputBuffer, writer)
		cmd.Stdout = writer
		cmd.Stderr = writer
		cmd.Env = append(os.Environ(), opts.Env...)

		err := cmd.Run()
		output := outputBuffer.String()

		if err == nil {
			return output, 0, nil
		}

		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return output, 0, err
		}
		lastOutput = output
		lastCode = exitError.ExitCode()
		lastErr = fmt.Errorf("%w\n%s", exitError, output)

		time.Sleep(opts.RetryDelay)
		attempt++
	}
	return lastOutput, lastCode, lastErr
}

func Retry(fn func() error, maxAttempts int, delayBetweenAttempts time.Duration) error {
	lastErr := (error)(nil)
	attempt := 1
	for attempt <= maxAttempts {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		attempt++
		time.Sleep(delayBetweenAttempts)
	}
	return lastErr
}

package cliwrappers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

type CliExecutorInterface interface {
	Execute(command string, args ...string) (stdout, stderr string, exitCode int, err error)
	ExecuteInDir(wordir, command string, args ...string) (stdout, stderr string, exitCode int, err error)
	ExecuteWithOutput(command string, args ...string) (stdout, stderr string, exitCode int, err error)
}

var _ CliExecutorInterface = &CliExecutor{}

type CliExecutor struct {
	Verbose bool
}

func NewCliExecutor(verbose bool) *CliExecutor {
	return &CliExecutor{Verbose: verbose}
}

// Execute runs specified command with given arguments.
// Returns stdout, stderr, exit code, error
func (e *CliExecutor) Execute(command string, args ...string) (string, string, int, error) {
	return e.ExecuteInDir("", command, args...)
}

// ExecuteInDir runs given command in the specified directory.
// Returns stdout, stderr, exit code, error
func (e *CliExecutor) ExecuteInDir(wordir, command string, args ...string) (string, string, int, error) {
	cmd := exec.Command(command, args...)
	if wordir != "" {
		cmd.Dir = wordir
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if e.Verbose {
		l.Logger.Infof("Executing command: %s %s", command, strings.Join(args, " "))
	}

	err := cmd.Run()

	return stdoutBuf.String(), stderrBuf.String(), getExitCodeFromError(err), err
}

// ExecuteWithOutput runs a command with args while printing stdout and stderr in real time.
// Returns stdout, stderr, exit code, error
func (e *CliExecutor) ExecuteWithOutput(command string, args ...string) (string, string, int, error) {
	cmd := exec.Command(command, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to get stdout: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to get stderr: %w", err)
	}

	if e.Verbose {
		l.Logger.Infof("Executing command: %s %s", command, strings.Join(args, " "))
	}

	if err := cmd.Start(); err != nil {
		return "", "", -1, fmt.Errorf("failed to start command: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	readStream := func(r io.Reader, w io.Writer, buf *bytes.Buffer) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(w, line)
			buf.WriteString(line + "\n")
		}
	}

	done := make(chan struct{}, 2)
	go func() {
		readStream(stdoutPipe, os.Stdout, &stdoutBuf)
		done <- struct{}{}
	}()
	go func() {
		readStream(stderrPipe, os.Stderr, &stderrBuf)
		done <- struct{}{}
	}()

	err = cmd.Wait()
	// Wait for both output streams to finish
	<-done
	<-done

	return stdoutBuf.String(), stderrBuf.String(), getExitCodeFromError(err), err
}

func getExitCodeFromError(cmdErr error) int {
	if cmdErr == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(cmdErr, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}

func CheckCliToolsAvailable(cliTools []string) (bool, error) {
	for _, cliTool := range cliTools {
		isAvailable, err := CheckCliToolAvailable(cliTool)
		if !isAvailable {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func CheckCliToolAvailable(cliTool string) (bool, error) {
	cmd := exec.Command("sh", "-c", "command -v "+cliTool)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(output)) == "" {
		return false, nil
	}
	return true, nil
}

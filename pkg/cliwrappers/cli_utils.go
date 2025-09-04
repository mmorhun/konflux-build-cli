package cliwrappers

import (
	"bytes"
	"os/exec"
	"strings"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

type CliExecutorInterface interface {
	Execute(command string, args ...string) (stdout, stderr bytes.Buffer, err error)
	ExecuteInDir(wordir, command string, args ...string) (stdout, stderr bytes.Buffer, err error)
}

var _ CliExecutorInterface = &CliExecutor{}

type CliExecutor struct {
	Verbose bool
}

func NewCliExecutor(verbose bool) *CliExecutor {
	return &CliExecutor{Verbose: verbose}
}

func (e *CliExecutor) Execute(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
	return e.ExecuteInDir("", command, args...)
}

func (e *CliExecutor) ExecuteInDir(wordir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
	cmd := exec.Command(command, args...)
	if wordir != "" {
		cmd.Dir = wordir
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if e.Verbose {
		l.Logger.Infof("Executing command: %s %s", command, strings.Join(args, " "))
	}

	err = cmd.Run()
	return stdout, stderr, err
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

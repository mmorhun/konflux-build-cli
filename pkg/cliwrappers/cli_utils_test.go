package cliwrappers_test

import "bytes"

type mockExecutor struct {
	executeFunc      func(command string, args ...string) (stdout, stderr bytes.Buffer, err error)
	executeInDirFunc func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error)
}

func (m *mockExecutor) Execute(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
	if m.executeFunc != nil {
		return m.executeFunc(command, args...)
	}
	return bytes.Buffer{}, bytes.Buffer{}, nil
}

func (m *mockExecutor) ExecuteInDir(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
	if m.executeInDirFunc != nil {
		return m.executeInDirFunc(workdir, command, args...)
	}
	return bytes.Buffer{}, bytes.Buffer{}, nil
}

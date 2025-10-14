package commands_test

import (
	"github.com/mmorhun/konflux-task-cli/pkg/common"
)

var _ common.ResultsWriterInterface = &MockResultsWriter{}

type MockResultsWriter struct {
	Verbose bool

	WriteResultStringFunc func(result, path string) error

	// Result file path => result data
	WrittenResults map[string]string
}

func (m *MockResultsWriter) WriteResultString(result, path string) error {
	if m.WriteResultStringFunc != nil {
		if err := m.WriteResultStringFunc(result, path); err != nil {
			return err
		}
	}

	if m.WrittenResults == nil {
		m.WrittenResults = make(map[string]string)
	}
	m.WrittenResults[path] = result
	return nil
}

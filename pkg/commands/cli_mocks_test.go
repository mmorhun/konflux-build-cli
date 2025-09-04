package commands_test

import (
	"github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
)

var _ cliwrappers.GitCliInterface = &MockGitCli{}

type MockGitCli struct {
	CloneFunc              func(url, branch string, depth int) (string, error)
	GetRepoHeadFullShaFunc func(gitRepoDir string) (string, error)
}

func (m *MockGitCli) Clone(url, branch string, depth int) (string, error) {
	if m.CloneFunc != nil {
		return m.CloneFunc(url, branch, depth)
	}
	return "", nil
}

func (m *MockGitCli) GetRepoHeadFullSha(gitRepoDir string) (string, error) {
	if m.GetRepoHeadFullShaFunc != nil {
		return m.GetRepoHeadFullShaFunc(gitRepoDir)
	}
	return "", nil
}

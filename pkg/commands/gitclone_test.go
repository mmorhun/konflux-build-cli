package commands_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/mmorhun/konflux-task-cli/pkg/commands"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
)

const (
	repoUrl       = "https://github.com/test/repo.git"
	defaultBranch = "branch"
	gitSha        = "abcdef1234567890abcdef1234567890abcdef12"
	shortGitSha   = "abcdef1"
	clonedPath    = "repo"

	resultRepoUrlPath   = "/result/dir/repo_url"
	resultSourceDirPath = "/result/dir/source_dir"
	resultShaPath       = "/result/dir/sha"
	resultShortShaPath  = "/result/dir/short_sha"
)

func setupTestGitClone(mockResultsWriter *common.MockResultsWriter, mockGitCli *MockGitCli) *commands.GitClone {
	return &commands.GitClone{
		Params: &commands.GitCloneParams{
			RepoUrl: repoUrl,
			Branch:  defaultBranch,
			Depth:   1,
			Verbose: false,
		},
		Results: &commands.GitCloneResultFilesPath{
			Url:         resultRepoUrlPath,
			SourceDir:   resultSourceDirPath,
			Commit:      resultShaPath,
			ShortCommit: resultShortShaPath,
		},
		ResultsWriter: mockResultsWriter,
		CliWrappers: commands.GitCloneCliWrappers{
			GitCli: mockGitCli,
		},
	}
}

func TestGitClone_Success(t *testing.T) {
	g := NewWithT(t)

	mockGitCli := &MockGitCli{}
	mockResultsWriter := &common.MockResultsWriter{}
	gitClone := setupTestGitClone(mockResultsWriter, mockGitCli)

	mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
		g.Expect(url).To(Equal(repoUrl))
		g.Expect(branch).To(Equal(defaultBranch))
		g.Expect(depth).To(Equal(1))
		return clonedPath, nil
	}
	mockGitCli.GetRepoHeadFullShaFunc = func(gitRepoDir string) (string, error) {
		g.Expect(gitRepoDir).To(Equal(clonedPath))
		return gitSha, nil
	}

	err := gitClone.Run()
	g.Expect(err).ToNot(HaveOccurred())

	// We've written 4 results: repo url, source dir, git sha, short git sha
	g.Expect(mockResultsWriter.WrittenResults).To(HaveLen(4))

	g.Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultRepoUrlPath))
	g.Expect(mockResultsWriter.WrittenResults[resultRepoUrlPath]).To(Equal(repoUrl))

	g.Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultSourceDirPath))
	g.Expect(mockResultsWriter.WrittenResults[resultSourceDirPath]).To(Equal(clonedPath))

	g.Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultShaPath))
	g.Expect(mockResultsWriter.WrittenResults[resultShaPath]).To(Equal(gitSha))

	g.Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultShortShaPath))
	g.Expect(mockResultsWriter.WrittenResults[resultShortShaPath]).To(Equal(shortGitSha))
}

func TestGitClone_CloneError(t *testing.T) {
	g := NewWithT(t)

	mockGitCli := &MockGitCli{}
	mockResultsWriter := &common.MockResultsWriter{}
	gitClone := setupTestGitClone(mockResultsWriter, mockGitCli)

	mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
		return "", errors.New("clone failed")
	}

	err := gitClone.Run()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("clone failed"))
}

func TestGitClone_GetShaError(t *testing.T) {
	g := NewWithT(t)

	mockGitCli := &MockGitCli{}
	mockResultsWriter := &common.MockResultsWriter{}
	gitClone := setupTestGitClone(mockResultsWriter, mockGitCli)

	mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
		return repoUrl, nil
	}
	mockGitCli.GetRepoHeadFullShaFunc = func(gitRepoDir string) (string, error) {
		return "", errors.New("failed to get SHA")
	}

	err := gitClone.Run()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("failed to get SHA"))
}

func TestGitClone_WriteError(t *testing.T) {
	g := NewWithT(t)

	mockGitCli := &MockGitCli{}
	mockResultsWriter := &common.MockResultsWriter{}
	gitClone := setupTestGitClone(mockResultsWriter, mockGitCli)

	mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
		return repoUrl, nil
	}
	mockGitCli.GetRepoHeadFullShaFunc = func(gitRepoDir string) (string, error) {
		return gitSha, nil
	}
	mockResultsWriter.WriteResultStringFunc = func(content, path string) error {
		if path == resultShaPath {
			return errors.New("permission denied")
		}
		return nil
	}

	err := gitClone.Run()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("permission denied"))
}

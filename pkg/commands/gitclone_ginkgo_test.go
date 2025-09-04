package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mmorhun/konflux-task-cli/pkg/commands"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
)

var _ = Describe("Git Clone Command", func() {
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

	var (
		gitClone          *commands.GitClone
		mockGitCli        *MockGitCli
		mockResultsWriter *common.MockResultsWriter
	)

	setupTestGitClone := func(mockResultsWriter *common.MockResultsWriter, mockGitCli *MockGitCli) *commands.GitClone {
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

	Context("When succeeds", func() {
		BeforeEach(func() {
			mockGitCli = &MockGitCli{}
			mockResultsWriter = &common.MockResultsWriter{}
			gitClone = setupTestGitClone(mockResultsWriter, mockGitCli)
		})

		It("should successfully clone and write results", func() {
			mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
				Expect(url).To(Equal(repoUrl))
				Expect(branch).To(Equal(defaultBranch))
				Expect(depth).To(Equal(1))
				return clonedPath, nil
			}
			mockGitCli.GetRepoHeadFullShaFunc = func(gitRepoDir string) (string, error) {
				Expect(gitRepoDir).To(Equal(clonedPath))
				return gitSha, nil
			}

			err := gitClone.Run()
			Expect(err).ToNot(HaveOccurred())

			// We've written 4 results: repo url, source dir, git sha, short git sha
			Expect(mockResultsWriter.WrittenResults).To(HaveLen(4))

			Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultRepoUrlPath))
			Expect(mockResultsWriter.WrittenResults[resultRepoUrlPath]).To(Equal(repoUrl))

			Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultSourceDirPath))
			Expect(mockResultsWriter.WrittenResults[resultSourceDirPath]).To(Equal(clonedPath))

			Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultShaPath))
			Expect(mockResultsWriter.WrittenResults[resultShaPath]).To(Equal(gitSha))

			Expect(mockResultsWriter.WrittenResults).To(HaveKey(resultShortShaPath))
			Expect(mockResultsWriter.WrittenResults[resultShortShaPath]).To(Equal(shortGitSha))
		})
	})

	Context("When fails", func() {
		BeforeEach(func() {
			mockGitCli = &MockGitCli{}
			mockResultsWriter = &common.MockResultsWriter{}
			gitClone = setupTestGitClone(mockResultsWriter, mockGitCli)
		})

		It("should return the clone error", func() {
			mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
				return "", errors.New("clone failed")
			}

			err := gitClone.Run()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("clone failed"))
		})

		It("should return get full SHA error", func() {
			mockGitCli.CloneFunc = func(url, branch string, depth int) (string, error) {
				return repoUrl, nil
			}
			mockGitCli.GetRepoHeadFullShaFunc = func(gitRepoDir string) (string, error) {
				return "", errors.New("failed to get SHA")
			}

			err := gitClone.Run()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("failed to get SHA"))
		})

		It("should return a write error when writing result files fails", func() {
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
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("permission denied"))
		})
	})
})

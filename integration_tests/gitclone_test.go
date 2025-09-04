package integration_tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

const GitCloneImage = "quay.io/konflux-ci/git-clone@sha256:4e53ebd9242f05ca55bfc8d58b3363d8b9d9bc3ab439d9ab76cdbdf5b1fd42d9"

type GitCloneParams struct {
	RepoUrl string
	Branch  string
}

type GitCloneResults struct {
	Url         string
	SourceDir   string
	Commit      string
	CommitShort string
}

func RunGitClone(gitCloneParams GitCloneParams, volumeHostPath string) (GitCloneResults, error) {
	var err error

	container := NewTestContainer("gitclone", GitCloneImage, true)

	// Params
	container.AddEnv("GIT_REPO_URL", gitCloneParams.RepoUrl)
	container.AddEnv("GIT_BRANCH", gitCloneParams.Branch)
	container.AddEnv("VERBOSE", "true")
	// Results
	container.AddTaskResult("RESULT_URL")
	container.AddTaskResult("RESULT_SOURCE_DIR")
	container.AddTaskResult("RESULT_COMMIT")
	container.AddTaskResult("RESULT_SHORT_COMMIT")

	if volumeHostPath != "" {
		container.AddVolume(volumeHostPath, "/pvc")
		container.SetWorkdir("/pvc")
	}

	if Debug {
		container.AddPort("2345", "2345")
	}
	err = container.Start()
	Expect(err).ToNot(HaveOccurred())
	defer container.Delete()

	err = container.CopyFileIntoContainer("../"+KonfluxCli, "/usr/bin/")
	Expect(err).ToNot(HaveOccurred())

	if Debug {
		err = container.DebugCli("gitclone")
	} else {
		err = container.ExecuteAndWait(KonfluxCli, "gitclone")
	}
	Expect(err).ToNot(HaveOccurred())

	resultUrl, err := container.GetTaskResultValue("RESULT_URL")
	Expect(err).ToNot(HaveOccurred())
	sourceDir, err := container.GetTaskResultValue("RESULT_SOURCE_DIR")
	Expect(err).ToNot(HaveOccurred())
	resultCommit, err := container.GetTaskResultValue("RESULT_COMMIT")
	Expect(err).ToNot(HaveOccurred())
	resultCommitShort, err := container.GetTaskResultValue("RESULT_SHORT_COMMIT")
	Expect(err).ToNot(HaveOccurred())

	return GitCloneResults{
		Url:         resultUrl,
		SourceDir:   sourceDir,
		Commit:      resultCommit,
		CommitShort: resultCommitShort,
	}, nil
}

func TestGitClone(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Printf("Test Failure: %s\n", message)
		t.FailNow() // Terminate the test immediately
	})
	ExpectKonfluxCliCompiled()

	volumeDir := CreateTempDir("git-clone-pvc-*")
	defer os.RemoveAll(volumeDir)

	const repoUrl = "https://github.com/devfile-samples/devfile-sample-go-basic"
	const branch = "main"
	gitCloneParams := GitCloneParams{
		RepoUrl: repoUrl,
		Branch:  branch,
	}

	gitCloneResults, err := RunGitClone(gitCloneParams, volumeDir)
	Expect(err).ToNot(HaveOccurred())

	Expect(gitCloneResults.Url).To(Equal(repoUrl))
	Expect(gitCloneResults.SourceDir).To(Equal("devfile-sample-go-basic"))
	Expect(gitCloneResults.Commit).To(MatchRegexp(`^[0-9a-f]{40}$`))
	Expect(gitCloneResults.CommitShort).To(MatchRegexp(`^[0-9a-f]{7}$`))
}

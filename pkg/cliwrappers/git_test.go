package cliwrappers_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
)

func setupGitCli() (*cliwrappers.GitCli, *mockExecutor) {
	executor := &mockExecutor{}
	gitCli := &cliwrappers.GitCli{
		Executor: executor,
		Verbose:  false,
	}
	return gitCli, executor
}

func TestGitCli_Clone_DefaultBranch(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	var capturedArgs []string
	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		g.Expect(command).To(Equal("git"))
		capturedArgs = args
		stderr.WriteString("Cloning into 'test-repo'...\n")
		return stdout, stderr, nil
	}

	_, err := gitCli.Clone("https://github.com/test/repo.git", "", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(capturedArgs).To(HaveLen(4))
	g.Expect(capturedArgs).To(ContainElement("clone"))
	g.Expect(capturedArgs).To(ContainElement("https://github.com/test/repo.git"))
	g.Expect(capturedArgs).To(ContainElement("--branch"))
	g.Expect(capturedArgs).To(ContainElement("main"))
}

func TestGitCli_Clone_SpecifiedBranch(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	var capturedArgs []string
	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		capturedArgs = args
		stderr.WriteString("Cloning into 'test-repo'...\n")
		return stdout, stderr, nil
	}

	_, err := gitCli.Clone("https://github.com/test/repo.git", "devel", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(capturedArgs).To(ContainElement("--branch"))
	g.Expect(capturedArgs).To(ContainElement("devel"))
}

func TestGitCli_Clone_WithDepth(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	var capturedArgs []string
	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		capturedArgs = args
		stderr.WriteString("Cloning into 'test-repo'...\n")
		return stdout, stderr, nil
	}

	_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 5)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(capturedArgs).To(ContainElement("--depth"))
	g.Expect(capturedArgs).To(ContainElement("5"))
}

func TestGitCli_Clone_NoDepthWhenZero(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	var capturedArgs []string
	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		capturedArgs = args
		stderr.WriteString("Cloning into 'test-repo'...\n")
		return stdout, stderr, nil
	}

	_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(capturedArgs).NotTo(ContainElement("--depth"))
}

func TestGitCli_Clone_ReturnsRepoPath(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("Cloning into 'my-custom-repo'...\n")
		return stdout, stderr, nil
	}

	repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(repoPath).To(ContainSubstring("my-custom-repo"))
}

func TestGitCli_Clone_ComplexRepoNames(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("Cloning into 'repo-with-dashes_and_underscores.git'...\n")
		return stdout, stderr, nil
	}

	repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(repoPath).To(ContainSubstring("repo-with-dashes_and_underscores.git"))
}

func TestGitCli_Clone_WithAdditionalHeaders(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("Note: switching to 'main'.\n")
		stderr.WriteString("Cloning into 'test-repo'...\n")
		stderr.WriteString("remote: Counting objects: 100, done.\n")
		return stdout, stderr, nil
	}

	repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(repoPath).To(ContainSubstring("test-repo"))
}

func TestGitCli_Clone_FailsOnGitError(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("fatal: repository not found")
		return stdout, stderr, errors.New("exit status 128")
	}

	_, err := gitCli.Clone("https://github.com/test/nonexistent.git", "main", 0)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("git clone failed"))
}

func TestGitCli_Clone_FailsOnEmptyURL(t *testing.T) {
	g := NewWithT(t)
	gitCli, _ := setupGitCli()

	_, err := gitCli.Clone("", "main", 0)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("url must be set to clone"))
}

func TestGitCli_Clone_FailsOnUnparseableOutput(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("Some unexpected output without cloning info\n")
		return stdout, stderr, nil
	}

	_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("failed to obtain cloned repository directory"))
}

func TestGitCli_GetRepoHeadFullSha_Success(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
	executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		g.Expect(workdir).To(Equal("/path/to/repo"))
		g.Expect(command).To(Equal("git"))
		g.Expect(args).To(Equal([]string{"rev-parse", "HEAD"}))
		stdout.WriteString(expectedSha + "\n")
		return stdout, stderr, nil
	}

	sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(sha).To(Equal(expectedSha))
}

func TestGitCli_GetRepoHeadFullSha_TrimsWhitespace(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
	executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stdout.WriteString("  " + expectedSha + "  \n\t")
		return stdout, stderr, nil
	}

	sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(sha).To(Equal(expectedSha))
}

func TestGitCli_GetRepoHeadFullSha_VerboseMode(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()
	gitCli.Verbose = true

	expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
	executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stdout.WriteString(expectedSha + "\n")
		return stdout, stderr, nil
	}

	sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(sha).To(Equal(expectedSha))
}

func TestGitCli_GetRepoHeadFullSha_FailsOnGitError(t *testing.T) {
	g := NewWithT(t)
	gitCli, executor := setupGitCli()

	executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
		stderr.WriteString("fatal: not a git repository")
		return stdout, stderr, errors.New("exit status 128")
	}

	_, err := gitCli.GetRepoHeadFullSha("/invalid/path")

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("git rev-parse failed"))
}

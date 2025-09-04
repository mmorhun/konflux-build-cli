package cliwrappers_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
)

var _ = Describe("Git Clone Cli Wrapper", func() {
	var (
		gitCli   *cliwrappers.GitCli
		executor *mockExecutor
	)

	BeforeEach(func() {
		executor = &mockExecutor{}
		gitCli = &cliwrappers.GitCli{
			Executor: executor,
			Verbose:  false,
		}
	})

	Describe("Clone", func() {
		Context("When succeeds", func() {
			It("should clone with default branch when branch is empty", func() {
				var capturedArgs []string
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					Expect(command).To(Equal("git"))
					capturedArgs = args
					stderr.WriteString("Cloning into 'test-repo'...\n")
					return stdout, stderr, nil
				}

				_, err := gitCli.Clone("https://github.com/test/repo.git", "", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedArgs).To(ContainElement("clone"))
				Expect(capturedArgs).To(ContainElement("https://github.com/test/repo.git"))
				Expect(capturedArgs).To(ContainElement("--branch"))
				Expect(capturedArgs).To(ContainElement("main"))
			})

			It("should clone with specified branch", func() {
				var capturedArgs []string
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					capturedArgs = args
					stderr.WriteString("Cloning into 'test-repo'...\n")
					return stdout, stderr, nil
				}

				_, err := gitCli.Clone("https://github.com/test/repo.git", "devel", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedArgs).To(ContainElement("--branch"))
				Expect(capturedArgs).To(ContainElement("devel"))
			})

			It("should clone with specified depth", func() {
				var capturedArgs []string
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					capturedArgs = args
					stderr.WriteString("Cloning into 'test-repo'...\n")
					return stdout, stderr, nil
				}

				_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 5)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedArgs).To(ContainElement("--depth"))
				Expect(capturedArgs).To(ContainElement("5"))
			})

			It("should not include depth when depth is 0", func() {
				var capturedArgs []string
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					capturedArgs = args
					stderr.WriteString("Cloning into 'test-repo'...\n")
					return stdout, stderr, nil
				}

				_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedArgs).NotTo(ContainElement("--depth"))
			})

			It("should return repository path on successful clone", func() {
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("Cloning into 'my-custom-repo'...\n")
					return stdout, stderr, nil
				}

				repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(repoPath).To(ContainSubstring("my-custom-repo"))
			})

			It("should handle complex repository names", func() {
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("Cloning into 'repo-with-dashes_and_underscores.git'...\n")
					return stdout, stderr, nil
				}

				repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(repoPath).To(ContainSubstring("repo-with-dashes_and_underscores.git"))
			})

			It("should handle git clone output with additional headers", func() {
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("Note: switching to 'main'.\n")
					stderr.WriteString("Cloning into 'test-repo'...\n")
					stderr.WriteString("remote: Counting objects: 100, done.\n")
					return stdout, stderr, nil
				}

				repoPath, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(repoPath).To(ContainSubstring("test-repo"))
			})

		})

		Context("When fails", func() {
			It("should handle git clone failure", func() {
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("fatal: repository not found")
					return stdout, stderr, errors.New("exit status 128")
				}

				_, err := gitCli.Clone("https://github.com/test/nonexistent.git", "main", 0)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("git clone failed"))
			})

			It("should fail if git repo url is not provided", func() {
				_, err := gitCli.Clone("", "main", 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("url must be set to clone"))
			})

			It("should fail when unable to parse repository directory", func() {
				executor.executeFunc = func(command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("Some unexpected output without cloning info\n")
					return stdout, stderr, nil
				}

				_, err := gitCli.Clone("https://github.com/test/repo.git", "main", 0)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to obtain cloned repository directory"))
			})
		})
	})

	Describe("GetRepoHeadFullSha", func() {
		Context("When git rev-parse succeeds", func() {
			It("should return the full SHA", func() {
				expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
				executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					Expect(workdir).To(Equal("/path/to/repo"))
					Expect(command).To(Equal("git"))
					Expect(args).To(Equal([]string{"rev-parse", "HEAD"}))
					stdout.WriteString(expectedSha + "\n")
					return stdout, stderr, nil
				}

				sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(sha).To(Equal(expectedSha))
			})

			It("should trim whitespace from SHA", func() {
				expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
				executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stdout.WriteString("  " + expectedSha + "  \n\t")
					return stdout, stderr, nil
				}

				sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(sha).To(Equal(expectedSha))
			})

			It("should handle verbose output", func() {
				gitCli.Verbose = true
				expectedSha := "abcd1234567890abcdef1234567890abcdef1234"
				executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stdout.WriteString(expectedSha + "\n")
					return stdout, stderr, nil
				}

				sha, err := gitCli.GetRepoHeadFullSha("/path/to/repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(sha).To(Equal(expectedSha))
			})
		})

		Context("When git rev-parse fails", func() {
			It("should return an error", func() {
				executor.executeInDirFunc = func(workdir, command string, args ...string) (stdout, stderr bytes.Buffer, err error) {
					stderr.WriteString("fatal: not a git repository")
					return stdout, stderr, errors.New("exit status 128")
				}

				_, err := gitCli.GetRepoHeadFullSha("/invalid/path")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("git rev-parse failed"))
			})
		})
	})
})
